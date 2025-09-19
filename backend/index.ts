
import { ProxyAgent, setGlobalDispatcher } from 'undici';
import Koa from 'koa';
import Router from '@koa/router';
import run from './run.ts';
import { email, email_config } from './communication.ts';

//测试环境下使用代理
const httpDispatcher = new ProxyAgent({ uri: 'http://127.0.0.1:7890' });
setGlobalDispatcher(httpDispatcher);

const app = new Koa();
const router = new Router();

router.get('/', async (ctx, next) => {
    ctx.body = 'Hello World!'
})
router.get('/config/email', async (ctx, next) => {
    ctx.body = await email_config();
})
router.get('/testemail', async (ctx, next) => {
    email(
        '1.59⏰',
        '代币空投 将在2021-09-01 00:00:00 开始',
        'Hello',
        true
    )

    ctx.body = {
        'status': 'done'
    }
})

app.use(router.routes());
app.use(router.allowedMethods());

app.listen(3000);

//程序逻辑
// await run();
// const interval = setInterval(run, 10000);

// process.on('SIGINT', () => {
//     clearInterval(interval);
//     console.log('\n程序已停止');
//     process.exit();
// });

// console.log('程序已启动,每10秒执行一次函数...');
// console.log('按 Ctrl+C 停止程序');

