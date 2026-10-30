// Exporta cada bloque ```mermaid de docs/ a docs/exports/<formato>/, con la misma estructura de carpetas.
// Uso: node scripts/render.mjs [svg|png]
import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { basename, dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const docsDir = join(dirname(fileURLToPath(import.meta.url)), '..');
const format = process.argv[2] === 'png' ? 'png' : 'svg';
const exportsDir = join(docsDir, 'exports', format);
const cli = join(docsDir, 'node_modules', '@mermaid-js', 'mermaid-cli', 'src', 'cli.js');
const config = join(docsDir, 'mermaid.config.json');
const css = join(docsDir, 'mermaid.css');
const ignored = new Set(['node_modules', 'exports', 'scripts']);

function listMarkdown(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) return ignored.has(entry.name) ? [] : listMarkdown(path);
    return entry.name.endsWith('.md') ? [path] : [];
  });
}

function extractDiagrams(markdown) {
  return [...markdown.matchAll(/```mermaid\r?\n([\s\S]*?)```/g)].map((m) => m[1]);
}

// Si Puppeteer no pudo descargar su Chromium, se usa un navegador instalado en el sistema.
const systemBrowsers = [
  process.env.PUPPETEER_EXECUTABLE_PATH,
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  '/usr/bin/chromium',
  '/usr/bin/chromium-browser',
  '/usr/bin/google-chrome',
].filter(Boolean);

// Si un .md tiene varios diagramas, cada imagen se nombra con el título de su frontmatter.
function slugFromTitle(source) {
  const match = source.match(/^title:\s*"?([^"\r\n]+)"?\s*$/m);
  if (!match) return null;
  return match[1]
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

const workDir = join(tmpdir(), `docs-render-${process.pid}`);
mkdirSync(workDir, { recursive: true });

const browser = systemBrowsers.find((path) => existsSync(path));
const puppeteerConfig = join(workDir, 'puppeteer.json');
writeFileSync(
  puppeteerConfig,
  JSON.stringify(browser ? { executablePath: browser, headless: true } : { headless: true }),
);

let failures = 0;
let rendered = 0;

rmSync(exportsDir, { recursive: true, force: true });

for (const file of listMarkdown(docsDir)) {
  const diagrams = extractDiagrams(readFileSync(file, 'utf8'));
  const rel = relative(docsDir, file);
  if (diagrams.length === 0) continue;
  const outDir = join(exportsDir, dirname(rel));
  mkdirSync(outDir, { recursive: true });

  diagrams.forEach((source, index) => {
    const name = diagrams.length > 1
      ? `${basename(file, '.md')}-${slugFromTitle(source) ?? index + 1}`
      : basename(file, '.md');
    const input = join(workDir, `${name}.mmd`);
    const output = join(outDir, `${name}.${format}`);
    writeFileSync(input, source);

    const args = [cli, '-i', input, '-o', output, '-c', config, '-C', css, '-p', puppeteerConfig, '-b', 'white', '-q'];
    if (format === 'png') args.push('-s', '3');
    const result = spawnSync(process.execPath, args, { encoding: 'utf8' });

    if (result.status === 0) {
      rendered += 1;
      console.log(`ok    ${relative(docsDir, output)}`);
    } else {
      failures += 1;
      console.error(`error ${rel} (${name})\n${result.stderr || result.stdout}`);
    }
  });
}

rmSync(workDir, { recursive: true, force: true });
console.log(`\n${rendered} diagramas exportados, ${failures} con error.`);
process.exit(failures > 0 ? 1 : 0);
