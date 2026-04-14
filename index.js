#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

const binPath = path.join(__dirname, 'bin', 'auto-i18n');

const args = process.argv.slice(2);
const child = spawn(binPath, args, {
  stdio: 'inherit',
  shell: true
});

child.on('exit', (code) => {
  process.exit(code || 0);
});
