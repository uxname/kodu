#!/usr/bin/env node
// Тонкий лаунчер: запускает нативный бинарь kodu, скачанный postinstall'ом.
'use strict';

const path = require('node:path');
const fs = require('node:fs');
const { spawnSync } = require('node:child_process');

const binName = process.platform === 'win32' ? 'kodu.exe' : 'kodu';
const binPath = path.join(__dirname, binName);

if (!fs.existsSync(binPath)) {
  console.error('[kodu] Нативный бинарь не найден. Переустановите пакет или соберите из исходников:');
  console.error('[kodu]   go install github.com/uxname/kodu/cmd/kodu@latest');
  process.exit(1);
}

const res = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
if (res.error) {
  console.error(`[kodu] ${res.error.message}`);
  process.exit(1);
}
process.exit(res.status ?? 0);
