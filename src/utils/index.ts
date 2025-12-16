import chalk from 'chalk';
import cliProgress from "cli-progress";
import colors from "colors";
import figlet from "figlet";

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

export function createProgressBar(total: number) {
  return new cliProgress.SingleBar(
    {
      format: colors.cyan("Progress") + " |" +
              colors.green("{bar}") + "| " +
              colors.yellow("{percentage}%") + " | " +
              colors.magenta(`{value}/${total} Keys`),
      barCompleteChar: "█",
      barIncompleteChar: "░",
      hideCursor: true,
      barsize: 30,
    },
    cliProgress.Presets.shades_classic
  );
}

export const print = async (message:string) => {
 const text = figlet.textSync(message);
 return await console.log(text)
}

export const pickKeys = (keys: string[] | string) => {
 
  const keyList = typeof keys === "string" ? [keys] : [...(keys || [])];
  
  const result = new Set<string>();

  keyList.forEach((rawKey) => {
    
    const cleanKey = (rawKey || "").trim();
    if (!cleanKey) return;

    const keySegments = cleanKey.split(".");
    const finalKey = keySegments.pop() || ""; 
    
    if (finalKey) {
      result.add(finalKey);
    }
  });

  return Array.from(result);
};
