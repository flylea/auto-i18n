import fs from "node:fs/promises";
import { constants } from "node:fs";
import fg from "fast-glob";
import { logger } from "./utils/index.js";

export const scanFiles = async (): Promise<string[]> => {
  let config;
  try {
    const imported = await import("../i18n.config.js");
    config = imported.default;
  } catch {
    throw new Error("i18n.config.js not exists or format error");
  }

  const { entryDirs, extensions } = config;
  if (!entryDirs || !extensions) {
    throw new Error("i18n.config.js missing entryDirs/extensions config");
  }

  const existingDirs = [];
  for (const dir of entryDirs) {
    try {
      await fs.access(dir, constants.F_OK);
      existingDirs.push(dir);
    } catch {
      logger.warn(`Directory does not exist: ${dir}`);
    }
  }

  if (existingDirs.length === 0) {
    logger.error("None of the entryDirs exist. Please check your i18n.config.js");
    return [];
  }

  const patterns = existingDirs.map(dir => `${dir.replace(/\\/g, "/")}/**/*.${extensions.join(",")}`);
  logger.info(`Scanning directories: ${existingDirs.join(", ")}`);

  const files = await fg(patterns, { onlyFiles: true, absolute: true });

  return files;
}
