#!/usr/bin/env node
import { Command } from "commander";
import { scanFiles } from "./scanFiles.js";
import { extractKeys } from "./extractKeys.js";
import { logger } from "./utils/index.js";

const program = new Command();

program.name("auto-i18n").description("I18n a11y CLI").version("1.0.0");

program
  .command("scan")
  .description("scan project and print i18n keys")
  .action(async () => {
    const files = await scanFiles();
    for (const file of files) {
      const keys = await extractKeys(file);
      logger.info(`📄: ${file}\n🧩: ${keys}\n`);
    }
  });

program.parse(process.argv);

export {};
