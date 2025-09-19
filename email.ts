
// 邮箱配置
const smtpCode = '';
const smtpCodeType = '';
const smtpEmail = '';
const ColaKey = '';

const _send = (body: object): Promise<Response> => {
    return fetch('https://luckycola.com.cn/tools/customMail', {
        method: 'POST',
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(body)
    });
}

// "tomail": "1712881363@qq.com",
// "fromTitle": "币安空投提醒⏰",
// "subject": `${token} 代币空投 将在2021-09-01 00:00:00 开始`,
// "content": "<div style='color: red'>我是邮件内容(因为isTextContent=false所以我可以解析html标签,是红色的)</div>",
// "isTextContent": false,

export default (
    tomail: string | string[],
    fromTitle: string,
    subject: string,
    content: string,
    isTextContent: boolean = true // 是否是纯文本
): Promise<Response> | Promise<Response[]> => {
    const body = {
        fromTitle,
        subject,
        content,
        isTextContent,
        tomail: '',
        ColaKey,
        smtpCode,
        smtpCodeType,
        smtpEmail,
    };

    if (Array.isArray(tomail)) {
        const promises = tomail.map(email => {
            return _send({
                ...body,
                tomail: email
            });
        });
        return Promise.all(promises);
    } else {
        return _send({
            ...body,
            tomail
        });
    }
}