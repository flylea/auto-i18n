import fs from "node:fs";
import path from "node:path";
import inquirer from "inquirer";
import chalk from "chalk";

export const genI18nConfig = async (): Promise<void> => {
  console.log(
    chalk.cyan.bold("\n🚀 Auto I18n Config Generator\n")
  );

  // Step 1: entry directories
  const { entryDirs } = await inquirer.prompt<{
    entryDirs: string[];
  }>([
    {
      type: "input",
      name: "entryDirs",
      message: chalk.green(
        "Step 1: Enter entry directories (comma separated)"
      ),
      default: "src/views,src/pages,src/components",
      filter: (input: string) =>
        input
          .split(",")
          .map((dir) => dir.trim())
          .filter(Boolean),
    },
  ]);

  // Step 2: file extensions
  const { extensions } = await inquirer.prompt<{
    extensions: string[];
  }>([
    {
      type: "checkbox",
      name: "extensions",
      message: chalk.green(
        "Step 2: Select file extensions to scan"
      ),
      choices: ["vue", "ts", "js"],
      default: ["vue"],
    },
  ]);

  // Step 3: output format
  const { output } = await inquirer.prompt<{
    output: "ts" | "js" | "json";
  }>([
    {
      type: "list",
      name: "output",
      message: chalk.green("Step 3: Select output format"),
      choices: ["ts", "js", "json"],
      default: "ts",
    },
  ]);

  // Step 4: output directory
  const { outputDir } = await inquirer.prompt<{
    outputDir: string;
  }>([
    {
      type: "input",
      name: "outputDir",
      message: chalk.green(
        "Step 4: Enter output language directory"
      ),
      default: "src/i18n/locales",
    },
  ]);

  // Step 5: file header comment
  const { withFileComment } = await inquirer.prompt<{
    withFileComment: boolean;
  }>([
    {
      type: "confirm",
      name: "withFileComment",
      message: chalk.green(
        "Step 5: Include file header comment?"
      ),
      default: true,
    },
  ]);

  // Step 6: key → file path comment
  const { withKeyFileComment } = await inquirer.prompt<{
    withKeyFileComment: boolean;
  }>([
    {
      type: "confirm",
      name: "withKeyFileComment",
      message: chalk.green(
        "Step 6: Include key-to-source-file comments?"
      ),
      default: true,
    },
  ]);

  // Step 7: i18n functions
  const { i18nFns } = await inquirer.prompt<{
    i18nFns: string[];
  }>([
    {
      type: "checkbox",
      name: "i18nFns",
      message: chalk.green(
        "Step 7: Select i18n functions to match"
      ),
      choices: ["t", "$t", "i18n.t"],
      default: ["t", "$t", "i18n.t"],
    },
  ]);

  // Step 8: support Chinese key
  const { supportChineseKey } = await inquirer.prompt<{
    supportChineseKey: boolean;
  }>([
    {
      type: "confirm",
      name: "supportChineseKey",
      message: chalk.green(
        "Step 8: Support Chinese keys?"
      ),
      default: true,
    },
  ]);

  // Step 9: languages
  const { languages } = await inquirer.prompt<{
    languages: string[];
  }>([
    {
      type: "input",
      name: "languages",
      message: chalk.green(
        "Step 9: Enter languages to generate (comma separated)"
      ),
      default: "zh-CN,en-US",
      filter: (input: string) =>
        input
          .split(",")
          .map((lang) => lang.trim())
          .filter(Boolean),
    },
  ]);

  // Step 10: generate config content
  const configObject = {
    entryDirs,
    extensions,
    output,
    outputDir,
    withFileComment,
    withKeyFileComment,
    i18nFns,
    supportChineseKey,
    languages,
  };

  const configContent = `export default ${JSON.stringify(
    configObject,
    null,
    2
  )};\n`;

  // Step 11: write config file
  const filePath = path.resolve(
    process.cwd(),
    "i18n.config.ts"
  );

  fs.writeFileSync(filePath, configContent, "utf-8");

  console.log(
    chalk.yellow.bold(
      `\n✅ i18n.config.ts created successfully`
    )
  );
  console.log(
    chalk.gray(`📄 Path: ${filePath}\n`)
  );
};
