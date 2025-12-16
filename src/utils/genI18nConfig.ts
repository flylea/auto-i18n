import fs from 'fs';
import path from 'path';
import inquirer from 'inquirer';
import chalk from 'chalk';

export const genI18nConfig = async () => {
  // step 1：scan entry directories
  const { entryDirs } = await inquirer.prompt([
    {
      type: 'input',
      name: 'entryDirs',
      message: chalk.green('Step 1: Enter entry directories (comma separated)'),
      default: 'src/views,src/pages,src/components',
      filter: (input: string) => input.split(',').map(dir => dir.trim()),
    },
  ]);

  // step 2：scan file extensions
  const { extensions } = await inquirer.prompt([
    {
      type: 'checkbox',
      name: 'extensions',
      message: chalk.green('Step 2: Select file extensions to scan'),
      choices: ['vue', 'tsx', 'jsx', 'ts', 'js'],
      default: ['vue', 'tsx', 'jsx'],
    },
  ]);

  // step 3：output format
  const { output } = await inquirer.prompt([
    {
      type: 'list',
      name: 'output',
      message: chalk.green('Step 3: Select output format'),
      choices: ['ts', 'js', 'json'],
      default: 'ts',
    },
  ]);

  // step 4：output language directory
  const { outputDir } = await inquirer.prompt([
    {
      type: 'input',
      name: 'outputDir',
      message: chalk.green('Step 4: Enter output language directory'),
      default: 'src/i18n/locale',
    },
  ]);

  // step 5：include file comment (source info)?
  const { withFileComment } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'withFileComment',
      message: chalk.green('Step 5: Include file comment (source info)?'),
      default: true,
    },
  ]);

  // step 6：i18n functions to match
  const { i18nFns } = await inquirer.prompt([
    {
      type: 'checkbox',
      name: 'i18nFns',
      message: chalk.green('Step 6: Select i18n functions to match'),
      choices: ['t', '$t', 'i18n.t'],
      default: ['t', '$t', 'i18n.t'],
    },
  ]);

  // step 7：support chinese keys?
  const { supportChineseKey } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'supportChineseKey',
      message: chalk.green('Step 7: Support Chinese keys?'),
      default: true,
    },
  ]);
    // step 8：language packages to generate
    const { languages } = await inquirer.prompt([
    {
      type: 'input',
      name: 'languages',
      message: chalk.green('Step 8: Enter languages to generate (comma separated)'),
      default: 'zh-CN,en-US', 
      filter: (input: string) => input.split(',').map(lang => lang.trim()).filter(Boolean),
    },
  ]);

  // step 9：generate config object
  const configContent = `export default ${JSON.stringify(
    {
      entryDirs,
      extensions,
      output,
      outputDir,
      withFileComment,
      i18nFns,
      supportChineseKey,
      languages,
    },
    null,
    2
  )};\n`;

  // step 9：write config file
  const filePath = path.resolve(process.cwd(), 'i18n.config.ts');
  fs.writeFileSync(filePath, configContent, 'utf-8');

  console.log(chalk.yellow.bold(`\n🎉i18n.config.ts created successfully at ${filePath}\n`));
};
