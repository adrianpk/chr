#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const {execFileSync} = require('node:child_process');

function options(argv) {
  const result = {kind: argv[0]};
  for (let i = 1; i < argv.length; i += 2) {
    if (!argv[i].startsWith('--') || argv[i + 1] === undefined) {
      throw new Error('Options require a --name value pair.');
    }
    result[argv[i].slice(2)] = argv[i + 1];
  }
  if (!['web', 'tui'].includes(result.kind) || !result.output) {
    throw new Error('Use web --url URL --output PNG or tui --socket NAME --target PANE --output PNG.');
  }
  return result;
}

async function captureWeb(browser, args) {
  if (!args.url) throw new Error('Web capture requires --url.');
  const width = Number(args.width || 1440);
  const height = Number(args.height || 1000);
  const top = Number(args.top || 0);
  const page = await browser.newPage({
    viewport: {width, height},
    deviceScaleFactor: 1,
  });
  await page.goto(args.url, {waitUntil: 'domcontentloaded'});
  if (args.scroll) await page.mouse.wheel(0, Number(args.scroll));
  if (args.form) {
    const fields = JSON.parse(fs.readFileSync(args.form, 'utf8'));
    for (const [selector, value] of Object.entries(fields)) {
      await page.locator(selector).fill(value);
    }
  }
  if (args.wait) await page.locator(args.wait).waitFor();
  await settle(page);
  await page.screenshot({path: args.output, clip: {x: 0, y: top, width, height: height - top}});
}

async function settle(page) {
  await page.evaluate(() => document.fonts.ready);
  await page.evaluate(() => new Promise(resolve => {
    requestAnimationFrame(() => requestAnimationFrame(resolve));
  }));
}

async function captureTerminal(browser, args) {
  if (!args.socket || !args.target) throw new Error('Terminal capture requires --socket and --target.');
  const tmux = command => execFileSync('tmux', ['-L', args.socket, ...command], {encoding: 'utf8'});
  const size = tmux(['display-message', '-p', '-t', args.target, '#{pane_width} #{pane_height}']);
  const [cols, rows] = size.trim().split(/\s+/).map(Number);
  const pane = tmux(['capture-pane', '-p', '-e', '-t', args.target]).replace(/\r?\n$/, '');
  const xtermMain = require.resolve('@xterm/xterm');
  const xtermRoot = path.dirname(path.dirname(xtermMain));
  const page = await browser.newPage({
    viewport: {width: Math.max(1900, cols * 13), height: Math.max(1200, rows * 26)},
    deviceScaleFactor: 1,
  });
  await page.setContent('<html><head><style>html,body{margin:0;background:#1b1b27}#frame{display:inline-block;padding:16px;background:#1b1b27}.xterm-viewport{overflow:hidden!important}</style></head><body><div id="frame"><div id="terminal"></div></div></body></html>');
  await page.addStyleTag({path: path.join(xtermRoot, 'css/xterm.css')});
  await page.addScriptTag({path: xtermMain});
  await page.evaluate(size => {
    window.terminal = new Terminal({
      cols: size.cols,
      rows: size.rows,
      fontFamily: '"CaskaydiaMono Nerd Font Mono", monospace',
      fontSize: 18,
      lineHeight: 1.1,
      convertEol: true,
      cursorBlink: false,
      disableStdin: true,
      theme: {background: '#1b1b27', foreground: '#abb0d4', cursor: '#5dd0d3'},
    });
    terminal.open(document.getElementById('terminal'));
  }, {cols, rows});
  // Render the real pane's ANSI cells; no screen content is synthesized.
  await page.evaluate(data => new Promise(resolve => {
    terminal.write(Uint8Array.from(atob(data), ch => ch.charCodeAt(0)), resolve);
  }), Buffer.from(pane).toString('base64'));
  await settle(page);
  await page.locator('#frame').screenshot({path: args.output});
}

async function main() {
  const args = options(process.argv.slice(2));
  const {chromium} = require('playwright');
  const browser = await chromium.launch({
    executablePath: process.env.CHR_CAPTURE_BROWSER || '/usr/bin/chromium',
    headless: true,
    args: ['--disable-gpu', '--disable-dev-shm-usage'],
  });
  try {
    fs.mkdirSync(path.dirname(path.resolve(args.output)), {recursive: true});
    if (args.kind === 'web') await captureWeb(browser, args);
    else await captureTerminal(browser, args);
  } finally {
    await browser.close();
  }
}

main().catch(error => {
  console.error(error.message);
  process.exitCode = 1;
});
