const { execSync } = require('child_process');
const os = require('os');
const path = require('path');
const fs = require('fs');

const platform = os.platform();
const arch = os.arch();
const version = require('../package.json').version;

const platforms = {
  darwin: { ext: '', name: 'darwin' },
  linux: { ext: '', name: 'linux' },
  win32: { ext: '.exe', name: 'win.exe' }
};

const archs = {
  x64: 'x64'
};

function getPlatformKey() {
  const p = platforms[platform];
  const a = archs[arch];
  if (!p || !a) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    process.exit(1);
  }
  return `${p.name}-${a}`;
}

function getAssetName() {
  const platformKey = getPlatformKey();
  const ext = platform === 'win32' ? '.exe' : '';
  return `auto-i18n-${platformKey}${ext}`;
}

function getDownloadUrl() {
  const assetName = getAssetName();
  return `https://github.com/flylea/auto-i18n/releases/download/v${version}/${assetName}`;
}

const binDir = path.join(__dirname, '..', 'bin');
const binPath = path.join(binDir, 'auto-i18n');

if (fs.existsSync(binPath)) {
  console.log('auto-i18n already installed');
  return;
}

console.log(`Downloading auto-i18n v${version} for ${platform}-${arch}...`);

try {
  const https = require('https');
  const http = require('http');
  const url = require('url');

  const parsedUrl = url.parse(getDownloadUrl());
  const client = parsedUrl.protocol === 'https:' ? https : http;

  console.log(`Downloading from ${getDownloadUrl()}`);

  const file = fs.createWriteStream(binPath);
  client.get(parsedUrl, (response) => {
    if (response.statusCode === 302 || response.statusCode === 301) {
      // Follow redirect
      const redirectUrl = url.parse(response.headers.location);
      const redirectClient = redirectUrl.protocol === 'https:' ? https : http;
      redirectClient.get(response.headers.location, (redirectResponse) => {
        redirectResponse.pipe(file);
        file.on('finish', () => {
          file.close();
          fs.chmodSync(binPath, 0o755);
          console.log('auto-i18n installed successfully');
        });
      });
    } else {
      response.pipe(file);
      file.on('finish', () => {
        file.close();
        fs.chmodSync(binPath, 0o755);
        console.log('auto-i18n installed successfully');
      });
    }
  }).on('error', (err) => {
    fs.unlinkSync(binPath);
    console.error(`Failed to download: ${err.message}`);
    console.error('Please download manually from: https://github.com/flylea/auto-i18n/releases');
  });
} catch (err) {
  console.error(`Download failed: ${err.message}`);
  console.error('Please download manually from: https://github.com/flylea/auto-i18n/releases');
}
