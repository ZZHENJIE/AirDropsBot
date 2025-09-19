import { readFile } from "fs/promises";

export const email_config = async () => {
    const data = await readFile('./config/email.json', 'utf8');
    return JSON.parse(data);
}

export const email = async (
    fromTitle: string,
    subject: string,
    content: string,
    isTextContent: boolean = true // 是否是纯文本
): Promise<Response[]> => {
    const { smtpCode, smtpCodeType, smtpEmail, ColaKey, tomail } = await email_config();

    const template = {
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

    const promises = tomail.map((email: string) => {
        return fetch('https://luckycola.com.cn/tools/customMail', {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                ...template,
                tomail: email
            })
        });
    });

    return Promise.all(promises);
}

// "tomail": "1712881363@qq.com",
// "fromTitle": "币安空投提醒⏰",
// "subject": `${token} 代币空投 将在2021-09-01 00:00:00 开始`,
// "content": "<div style='color: red'>我是邮件内容(因为isTextContent=false所以我可以解析html标签,是红色的)</div>",
// "isTextContent": false,

export default {
    email_config,
    email
}