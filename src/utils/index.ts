import chalk from 'chalk';
import figlet from "figlet";
import fs from "fs";
import path from "path";
import dotenv from "dotenv";

const getTimestamp = () => new Date().toLocaleString();

export const logger = {
  info: (...args: unknown[]) => 
    console.log(chalk.blue(`[INFO] [${getTimestamp()}]`), ...args),
  success: (...args: unknown[]) => 
    console.log(chalk.green(`[SUCCESS] [${getTimestamp()}]`), ...args),
  warn: (...args: unknown[]) => 
    console.log(chalk.yellow(`[WARN] [${getTimestamp()}]`), ...args),
  error: (...args: unknown[]) => 
    console.error(chalk.red(`[ERROR] [${getTimestamp()}]`), ...args),
};

export const print = async (message:string) => {
 const text = figlet.textSync(message);
 return await console.log(text)
}

export const pickKeys = (keys: string[] | string) => {
  const keyList = typeof keys === "string" ? [keys] : keys || [];
  const result = new Set<string>();

  keyList.forEach((rawKey) => {
    const cleanKey = (rawKey || "").trim();
    if (!cleanKey) return;

    result.add(cleanKey);
  });

  return Array.from(result);
};

export function loadEnv() {
  const cwd = process.cwd();

  const candidates = [
    path.resolve(cwd, "env/.env.local"),
    path.resolve(cwd, "env/.env"),
    path.resolve(cwd, ".env.local"),
    path.resolve(cwd, ".env"),
  ];

  for (const filePath of candidates) {
    if (fs.existsSync(filePath)) {
      dotenv.config({ path: filePath });
      return filePath;
    }
  }

  return null;
}


