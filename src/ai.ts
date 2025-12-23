import { loadEnv } from "./utils/index.js";  
loadEnv(); 

import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: process.env.DEEPSEEK_API_URL,
  apiKey: process.env.DEEPSEEK_API_KEY,
});

const rules = `
你是一个国际化语言包翻译助手。

输入一个 key, 例如:
- "email"
- "title"
- "name"
- "user.avatar"

根据 key 翻译成目标语言，遵循以下 

规则：
1. 输出必须是字符串。
2. 只返回 target 目标语言的内容，不要解释、不要日志、不要 markdown。
3. 如果 key 包含. 表示嵌套对象的路径，说明是一个页面的模块，只翻译最后一级。
  eg: "user.avatar" 只翻译 "avatar"

示例：
输入 translate ("title", "zh-CN")

输出：标题
`;

 const translate = async (key: string, target: string) => {
  const completion = await openai.chat.completions.create({
    model: "deepseek-chat",
    messages: [
      {
        role: "system",
        content: rules,
      },
      {
        role: "user",
        content: `翻译 key: "${key}"，目标语言：${target}`,
      },
    ],
    temperature: 0,
  });

  const content = completion.choices[0].message.content;
  return content?.trim() || "";
}

export default translate;