#!/usr/bin/env node
// Thin launcher: runs the native kodu binary downloaded by postinstall.
// If the binary is missing (the package manager skipped the postinstall hook —
// bun does this by default, or `npm i --ignore-scripts` — or the download failed),
// it is fetched lazily from GitHub Releases on first run.
'use strict';

const path = require('node:path');
const fs = require('node:fs');
const { spawnSync } = require('node:child_process');
const { install, binNameFor, REPO } = require('../scripts/install');

const goos = process.platform === 'win32' ? 'windows' : process.platform;
const binName = binNameFor(goos);
const binDir = __dirname;
const binPath = path.join(binDir, binName);

function run() {
  const res = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
  if (res.error) {
    console.error(`[kodu] ${res.error.message}`);
    process.exit(1);
  }
  process.exit(res.status ?? 0);
}

if (fs.existsSync(binPath)) {
  run();
} else {
  console.error('[kodu] Native binary not found, downloading it now…');
  install(binDir)
    .then(run)
    .catch((err) => {
      console.error(`[kodu] Failed to download the binary: ${err.message}`);
      console.error('[kodu] Reinstall the package or build it from source:');
      console.error('[kodu]   go install github.com/uxname/kodu/cmd/kodu@latest');
      console.error(`[kodu]   https://github.com/${REPO}/releases`);
      process.exit(1);
    });
}
