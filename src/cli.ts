#!/usr/bin/env node
import { Command } from "commander";
import path from "path";
import fs from "fs";
import { scanFiles } from "./scanFiles.js";
import { extractKeys } from "./extractKeys.js";
import translate from "./ai.ts";
import { logger, createProgressBar, print, pickKeys } from "./utils/index.js";

const program = new Command();
program.name("auto-i18n").description("I18n a11y CLI").version("1.0.0");

const onAction = async () => {
  try {

    await print("AUTO I18N");

    const configPath = path.resolve(process.cwd(), "i18n.config.ts");
    const { genI18nConfig } = await import("./utils/genI18nConfig.ts");

    if (!fs.existsSync(configPath)) {
      await genI18nConfig();
    }

    /* scan files */
    const files = await scanFiles();
    if (files.length === 0) return

    const config = await import("../i18n.config.ts");
    /* extract keys */
    let totalKeys = 0;
    for (const file of files) {
      try {
        const keys = await extractKeys(file);
        totalKeys += keys.length;

        const allKeys = pickKeys(keys)
        for (const key of allKeys) {

          for(const lang of config.default.languages) {
            const result = await translate(key, lang);

            logger.info(`result: ${result}`)
          }
          
        }
      } catch (err) {
        logger.error(`Extract keys from ${file} failed: ${(err as Error).message}`);
      }
    }

  } catch (err) {
    logger.error(`Execute failed: ${(err as Error).message}`);
    process.exit(1);
  }
};

program
  .command("scan")
  .description("scan project and print i18n keys")
  .action(onAction);

program.parse(process.argv);