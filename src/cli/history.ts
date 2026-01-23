import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import prompts from 'prompts';
import type { HistoryItem } from '../types';

const HISTORY_FILE = join(homedir(), '.config', 'koda', 'history.json');
const MAX_HISTORY_ITEMS = 50;

function ensureHistoryFileExists(): void {
  const dir = join(homedir(), '.config', 'koda');
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  if (!existsSync(HISTORY_FILE))
    writeFileSync(HISTORY_FILE, JSON.stringify([], null, 2), 'utf-8');
}

function loadHistory(): HistoryItem[] {
  ensureHistoryFileExists();
  try {
    const content = readFileSync(HISTORY_FILE, 'utf-8');
    return JSON.parse(content) as HistoryItem[];
  } catch {
    return [];
  }
}

function saveHistory(history: HistoryItem[]): void {
  ensureHistoryFileExists();
  try {
    writeFileSync(HISTORY_FILE, JSON.stringify(history, null, 2), 'utf-8');
  } catch (error) {
    console.error('Error saving history:', error);
  }
}

export function addToHistory(args: string[]): void {
  const history = loadHistory();
  const command = args.join(' ');
  const existingIndex = history.findIndex((item) => item.command === command);
  if (existingIndex !== -1) history.splice(existingIndex, 1);
  history.unshift({ command, timestamp: Date.now() });
  if (history.length > MAX_HISTORY_ITEMS) history.length = MAX_HISTORY_ITEMS;
  saveHistory(history);
}

export async function showHistoryMenu(): Promise<string[] | null> {
  const history = loadHistory();

  if (history.length === 0) {
    console.log('No history found.');
    return null;
  }

  const choices = history.map((item, index) => ({
    title: item.command,
    value: index,
    description: new Date(item.timestamp).toLocaleString(),
  }));

  const { value } = await prompts({
    type: 'select',
    name: 'value',
    message: 'Select a command to run',
    choices: [...choices, { title: 'Cancel', value: -1 }],
  });

  if (value === -1 || value === undefined || !history[value]) {
    return null;
  }

  return history[value].command.split(' ');
}

export function getHistory(): HistoryItem[] {
  return loadHistory();
}
