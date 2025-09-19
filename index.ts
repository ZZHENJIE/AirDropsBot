
import { ProxyAgent, setGlobalDispatcher } from 'undici';
import run from './run.ts';

//测试环境下使用代理
const httpDispatcher = new ProxyAgent({ uri: 'http://127.0.0.1:7890' });
setGlobalDispatcher(httpDispatcher);

//程序逻辑
await run();
// const interval = setInterval(run, 10000);

// process.on('SIGINT', () => {
//     clearInterval(interval);
//     console.log('\n程序已停止');
//     process.exit();
// });

// console.log('程序已启动,每10秒执行一次函数...');
// console.log('按 Ctrl+C 停止程序');

