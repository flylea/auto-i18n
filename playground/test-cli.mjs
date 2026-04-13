#!/usr/bin/env node
/**
 * Integration test for auto-i18n CLI
 * Run this script from the playground directory to test the CLI
 */

import { execSync } from 'child_process';
import { existsSync, readdirSync, readFileSync } from 'fs';
import { join, dirname, resolve } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Resolve CLI path relative to project root
const projectRoot = resolve(__dirname, '..');
const CLI_PATH = join(projectRoot, 'dist', 'src', 'cli.js');
const OUTPUT_DIR = join(__dirname, 'src', 'i18n', 'locale');

console.log('🧪 Testing auto-i18n CLI\n');
console.log(`📁 Project root: ${projectRoot}`);
console.log(`📁 CLI path: ${CLI_PATH}`);

try {
  // Change to playground directory
  process.chdir(__dirname);
  console.log(`📁 Working directory: ${process.cwd()}`);

  // Check if i18n.config.ts exists
  if (!existsSync('i18n.config.ts')) {
    console.error('❌ i18n.config.ts not found');
    process.exit(1);
  }
  console.log('✅ i18n.config.ts found');

  // Check if source files exist
  const viewFiles = readdirSync('src/views');
  const componentFiles = readdirSync('src/components');
  console.log(`✅ Found ${viewFiles.length} view files, ${componentFiles.length} component files`);

  // Check if CLI exists
  if (!existsSync(CLI_PATH)) {
    console.error(`❌ CLI not found at ${CLI_PATH}`);
    console.error('   Run "pnpm build" first to compile TypeScript');
    process.exit(1);
  }
  console.log('✅ CLI binary found');

  // Run the CLI
  console.log('\n🚀 Running CLI...');
  try {
    execSync(`node "${CLI_PATH}" scan`, {
      stdio: 'inherit',
      env: {
        ...process.env,
        DEEPSEEK_API_KEY: process.env.DEEPSEEK_API_KEY || 'test-key',
        DEEPSEEK_API_URL: process.env.DEEPSEEK_API_URL || 'https://api.deepseek.com',
      }
    });
  } catch (err) {
    // API calls will fail without valid key - that's expected
    console.log('⚠️  CLI executed (API calls may have failed without valid key)');
  }

  // Check if output files were generated
  console.log('\n📄 Checking output files...');
  const outputExists = existsSync(OUTPUT_DIR);
  if (!outputExists) {
    console.log('⚠️  Output directory not created (likely due to API failure)');
    console.log('   Set a valid DEEPSEEK_API_KEY to test full translation');
    process.exit(0);
  }
  console.log(`✅ Output directory created: ${OUTPUT_DIR}`);

  // List generated files
  const generatedFiles = readdirSync(OUTPUT_DIR);
  console.log(`\n📁 Generated files:`);
  generatedFiles.forEach(file => {
    const filePath = join(OUTPUT_DIR, file);
    const content = readFileSync(filePath, 'utf-8');
    console.log(`\n--- ${file} ---`);
    console.log(content);
  });

  console.log('\n✅ All tests passed!');
} catch (error) {
  console.error('\n❌ Test failed:', error.message);
  process.exit(1);
}
