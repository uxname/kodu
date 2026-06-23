#!/usr/bin/env node
// Скачивает нативный бинарь kodu из GitHub Releases под текущую платформу.
// При неудаче не валит установку — печатает подсказку (graceful degradation).
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const https = require('node:https');
const { execFileSync } = require('node:child_process');

const REPO = 'uxname/kodu';
const pkg = require('../package.json');

const PLATFORMS = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const ARCHES = { x64: 'amd64', arm64: 'arm64' };

function fail(msg) {
  console.warn(`\n[kodu] ${msg}`);
  console.warn('[kodu] Установите вручную: go install github.com/uxname/kodu/cmd/kodu@latest');
  console.warn(`[kodu] или скачайте бинарь: https://github.com/${REPO}/releases\n`);
  process.exit(0); // не блокируем npm install
}

const goos = PLATFORMS[process.platform];
const goarch = ARCHES[process.arch];
if (!goos || !goarch) fail(`Платформа ${process.platform}/${process.arch} не поддерживается готовым бинарём.`);

const ext = goos === 'windows' ? 'zip' : 'tar.gz';
const binName = goos === 'windows' ? 'kodu.exe' : 'kodu';
const asset = `kodu_v${pkg.version}_${goos}_${goarch}.${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${pkg.version}/${asset}`;
const binDir = path.join(__dirname, '..', 'bin');

function download(u, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) return reject(new Error('too many redirects'));
    https.get(u, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        return resolve(download(res.headers.location, dest, redirects + 1));
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`HTTP ${res.statusCode} для ${u}`));
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on('finish', () => file.close(resolve));
      file.on('error', reject);
    }).on('error', reject);
  });
}

(async () => {
  try {
    fs.mkdirSync(binDir, { recursive: true });
    const archivePath = path.join(binDir, asset);
    await download(url, archivePath);
    // bsdtar (tar) распаковывает и .tar.gz, и .zip на Linux/macOS/Win10+.
    execFileSync('tar', ['-xf', archivePath, '-C', binDir], { stdio: 'ignore' });
    fs.rmSync(archivePath, { force: true });
    const binPath = path.join(binDir, binName);
    if (!fs.existsSync(binPath)) throw new Error('бинарь не найден после распаковки');
    if (goos !== 'windows') fs.chmodSync(binPath, 0o755);
    console.log(`[kodu] Установлен бинарь ${goos}/${goarch} v${pkg.version}`);
  } catch (err) {
    fail(`Не удалось скачать бинарь: ${err.message}`);
  }
})();
