import OpenAI from "openai";
import dotenv from "dotenv";
import path from "path";
dotenv.config({ path: path.resolve(process.cwd(), ".env.local") });

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

规则：
1. 输出必须是字符串。
2. 只返回 target 目标语言的内容，不要解释、不要日志、不要 markdown。

示例：
输入 translate ("title", "zh-CN")

输出：标题
`;

export default async function translate(key: string, target: string) {
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