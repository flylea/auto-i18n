#!/usr/bin/env node
import { Command } from "commander";
import path from "node:path";
import fs from "node:fs/promises";
import { scanFiles } from "./scanFiles.js";
import { extractKeys } from "./extractKeys.js";
import translate from "./ai.js";
import { logger, print, pickKeys } from "./utils/index.js";

const program = new Command();
program.name("auto-i18n").description("I18n a11y CLI").version("1.0.0");

const onAction = async () => {
  try {
    await print("AUTO I18N");

    const configPath = path.resolve(process.cwd(), "i18n.config.ts");
    const { genI18nConfig } = await import("./utils/genI18nConfig.js");

    if (!(await fs.access(configPath).then(() => true).catch(() => false))) {
      await genI18nConfig();
    }

    const files = await scanFiles();
    if (!files.length) return;

    const { default: config } = await import("../i18n.config.js");
    const translateResult: Record<string, Record<string, string>> = {};
    const keyFileMap = new Map<string, Set<string>>();

    for (const file of files) {
      try {
        const keys = await extractKeys(file);
        const allKeys = pickKeys(keys);

        for (const key of allKeys) {
          if (!keyFileMap.has(key)) keyFileMap.set(key, new Set());
          keyFileMap.get(key)?.add(file);

          for (const lang of config.languages) {
            if (!translateResult[key]) translateResult[key] = {};
            const result = await translate(key, lang);
            translateResult[key][lang] = result;
            logger.info(`Translated "${key}" to ${lang}: ${result}`);
          }
        }
      } catch (err) {
        logger.error(`Extract keys from ${file} failed: ${(err as Error).message}`);
      }
    }

    const { writeI18nFiles } = await import("./writeFile.js");
    await writeI18nFiles(translateResult, keyFileMap, config, logger);
  } catch (err) {
    logger.error(`Execute failed: ${(err as Error).message}`);
    process.exit(1);
  }
};

program.command("scan").description("scan project and print i18n keys").action(onAction);
program.parse(process.argv);
