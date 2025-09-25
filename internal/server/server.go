package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	cfgpkg "air-drops-bot/internal/config"
	"air-drops-bot/internal/scheduler"
	"air-drops-bot/internal/task"
)

// Server HTTP服务器结构体，提供 API 接口和状态管理
type Server struct {
	httpServer *http.Server
	cfg        *cfgpkg.Manager
	sched      *scheduler.Scheduler
	task       *task.Task
}

// New 创建新的HTTP服务器实例
// addr: 监听地址
// cfg: 配置管理器
// sched: 调度器
// task: 任务实例
func New(addr string, cfg *cfgpkg.Manager, sched *scheduler.Scheduler, task *task.Task) *Server {
	mux := http.NewServeMux()
	s := &Server{cfg: cfg, sched: sched, task: task}
	mux.HandleFunc("/start", s.postAuth(s.handleStart))
	mux.HandleFunc("/stop", s.postAuth(s.handleStop))
	mux.HandleFunc("/status", s.postAuth(s.handleStatus))
	mux.HandleFunc("/config/get", s.postAuth(s.handleGetConfig))
	mux.HandleFunc("/config/update", s.postAuth(s.handleUpdateConfig))

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// postAuth 请求认证中间件，验证POST请求和密码
func (s *Server) postAuth(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		pwd, _ := req["password"].(string)
		if pwd == "" {
			http.Error(w, "password required", http.StatusUnauthorized)
			return
		}
		cfg := s.cfg.Get()
		if pwd != cfg.Password {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// attach parsed body to context for handlers to reuse
		ctx := context.WithValue(r.Context(), ctxKeyBody{}, req)
		next(w, r.WithContext(ctx))
	}
}

type ctxKeyBody struct{}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
	}()
	return s.httpServer.ListenAndServe()
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg.Get()
	if s.sched.IsRunning() {
		writeJSON(w, http.StatusOK, map[string]any{"message": "already running"})
		return
	}
	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if err := s.sched.Start(interval, s.task.MainTask); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "started"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if !s.sched.IsRunning() {
		writeJSON(w, http.StatusOK, map[string]any{"message": "already stopped"})
		return
	}
	s.sched.Stop()
	writeJSON(w, http.StatusOK, map[string]any{"message": "stopped"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	running := s.sched.IsRunning()
	cpuPct, memUsedPct, err := systemMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"timer_running": running,
		"cpu_percent":   cpuPct,
		"mem_percent":   memUsedPct,
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// Always reload to ensure latest on disk
	if err := s.cfg.Reload(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cfg := s.cfg.Get()
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if s.sched.IsRunning() {
		http.Error(w, "stop scheduler before updating config", http.StatusConflict)
		return
	}
	var body map[string]any
	if v := r.Context().Value(ctxKeyBody{}); v != nil {
		body, _ = v.(map[string]any)
	}
	if body == nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	// Accept full config object in "config" field or top-level fields
	var newCfg cfgpkg.Config
	if cfgRaw, ok := body["config"]; ok {
		bytes, _ := json.Marshal(cfgRaw)
		if err := json.Unmarshal(bytes, &newCfg); err != nil {
			http.Error(w, "invalid config payload", http.StatusBadRequest)
			return
		}
	} else {
		// merge with current config to preserve unspecified fields
		current := s.cfg.Get()
		newCfg = current

		// 如果提交的是部分配置，解析并合并
		if configData, ok := body["config"].(map[string]interface{}); ok {
			// 更新 email 配置（如果提供）
			if emailData, ok := configData["email"].(map[string]interface{}); ok {
				if err := json.Unmarshal(mustJSON(emailData), &newCfg.Email); err != nil {
					http.Error(w, "invalid email config", http.StatusBadRequest)
					return
				}
			}
			// 更新其他顶层字段（如果提供）
			if password, ok := configData["password"].(string); ok && password != "" {
				newCfg.Password = password
			}
			if interval, ok := configData["interval_seconds"].(float64); ok {
				newCfg.IntervalSeconds = int(interval)
			}
		}

		// 处理旧版本的顶层字段更新方式
		if pw, ok := body["new_password"].(string); ok && pw != "" {
			newCfg.Password = pw
		}
		if iv, ok := body["interval_seconds"].(float64); ok {
			newCfg.IntervalSeconds = int(iv)
		}
	}
	// Ensure not changing to empty password
	if newCfg.Password == "" {
		http.Error(w, "password must not be empty", http.StatusBadRequest)
		return
	}
	if err := s.cfg.Update(newCfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "config updated"})
}

func systemMetrics() (cpuPercent float64, memPercent float64, err error) {
	// CPU: average over 200ms sample
	cpus, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		return 0, 0, err
	}
	if len(cpus) == 0 {
		return 0, 0, errors.New("cpu percent unavailable")
	}
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	return cpus[0], vm.UsedPercent, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// mustJSON 将任意值转换为 JSON 字节数组
func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// withCORS wraps an http.Handler and sets permissive CORS headers.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			// Preflight request - no body, just headers
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
