package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EmailConfig 邮件服务相关配置
type EmailConfig struct {
	SmtpCode     string   `json:"smtpCode"`
	SmtpCodeType string   `json:"smtpCodeType"`
	SmtpEmail    string   `json:"smtpEmail"`
	ColaKey      string   `json:"ColaKey"`
	ToMail       []string `json:"tomail"`
}

// Config 应用程序主配置结构，从磁盘读取
type Config struct {
	Password        string      `json:"password"`
	IntervalSeconds int         `json:"interval_seconds"`
	Email           EmailConfig `json:"email"`
}

// Manager 配置管理器，提供线程安全的配置访问和更新
type Manager struct {
	mu      sync.RWMutex
	path    string
	current *Config
}

// NewManager 创建新的配置管理器并加载指定路径的配置文件
func NewManager(configPath string) (*Manager, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	m := &Manager{path: absPath}
	if err := m.Reload(); err != nil {
		return nil, err
	}
	return m, nil
}

// Get 获取当前配置的快照副本
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.current == nil {
		return Config{}
	}
	cpy := *m.current
	return cpy
}

// Reload forces re-reading the configuration from disk.
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := readConfigFile(m.path)
	if err != nil {
		return err
	}
	m.current = cfg
	return nil
}

// Update writes the given configuration atomically to disk and updates memory.
func (m *Manager) Update(newCfg Config) error {
	if err := validate(newCfg); err != nil {
		return err
	}
	// Write atomically: write to temp file then rename
	tmpPath := m.path + ".tmp"
	data, err := json.MarshalIndent(newCfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, m.path); err != nil {
		return err
	}
	m.mu.Lock()
	m.current = &newCfg
	m.mu.Unlock()
	return nil
}

func readConfigFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(c Config) error {
	if c.Password == "" {
		return errors.New("password must not be empty")
	}
	if c.IntervalSeconds <= 0 {
		return errors.New("interval_seconds must be > 0")
	}
	if c.IntervalSeconds > int((24*time.Hour)/time.Second) {
		return errors.New("interval_seconds too large")
	}
	return nil
}
