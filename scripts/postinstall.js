#!/usr/bin/env node
// Downloads the native kodu binary for the current platform from GitHub Releases.
// On failure it does not break the install — it prints a hint (graceful degradation).
// The launcher (bin/kodu.js) will also download lazily on first run if this hook
// was skipped (bun) or failed, so a missing binary here is never fatal.
'use strict';

const path = require('node:path');
const { install, REPO } = require('./install');

const binDir = path.join(__dirname, '..', 'bin');

install(binDir)
  .then(({ goos, goarch, version }) =>
    console.log(`[kodu] Installed binary ${goos}/${goarch} v${version}`))
  .catch((err) => {
    console.warn(`\n[kodu] Failed to download the binary: ${err.message}`);
    console.warn('[kodu] It will be downloaded automatically on first run, or install manually:');
    console.warn('[kodu]   go install github.com/uxname/kodu/cmd/kodu@latest');
    console.warn(`[kodu]   https://github.com/${REPO}/releases\n`);
    process.exit(0); // do not block npm install
  });
