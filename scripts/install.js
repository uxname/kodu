// Shared installer: downloads the native kodu binary for the current platform
// from GitHub Releases. Used by the postinstall hook AND by the launcher's
// lazy fallback (so the package works even when the package manager skipped
// postinstall — e.g. bun by default, or `npm i --ignore-scripts`).
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const https = require('node:https');
const { execFileSync } = require('node:child_process');

const REPO = 'uxname/kodu';
const pkg = require('../package.json');

const PLATFORMS = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const ARCHES = { x64: 'amd64', arm64: 'arm64' };

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
        return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on('finish', () => file.close(resolve));
      file.on('error', reject);
    }).on('error', reject);
  });
}

// Resolves the binary path for the running platform, or null if unsupported.
function binNameFor(goos) {
  return goos === 'windows' ? 'kodu.exe' : 'kodu';
}

// Downloads + extracts the native binary into binDir. Returns the absolute
// binary path on success; throws on any failure (caller decides how to log).
async function install(binDir) {
  const goos = PLATFORMS[process.platform];
  const goarch = ARCHES[process.arch];
  if (!goos || !goarch) {
    throw new Error(`Platform ${process.platform}/${process.arch} is not supported by a prebuilt binary.`);
  }
  const ext = goos === 'windows' ? 'zip' : 'tar.gz';
  const binName = binNameFor(goos);
  const asset = `kodu_v${pkg.version}_${goos}_${goarch}.${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${pkg.version}/${asset}`;

  fs.mkdirSync(binDir, { recursive: true });
  const archivePath = path.join(binDir, asset);
  await download(url, archivePath);
  // bsdtar (tar) extracts both .tar.gz and .zip on Linux/macOS/Win10+.
  execFileSync('tar', ['-xf', archivePath, '-C', binDir], { stdio: 'ignore' });
  fs.rmSync(archivePath, { force: true });
  const binPath = path.join(binDir, binName);
  if (!fs.existsSync(binPath)) throw new Error('binary not found after extraction');
  if (goos !== 'windows') fs.chmodSync(binPath, 0o755);
  return { binPath, goos, goarch, version: pkg.version };
}

module.exports = { install, download, binNameFor, REPO, pkg };
