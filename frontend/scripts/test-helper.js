#!/usr/bin/env node

/**
 * Test Helper Script
 *
 * Script utilitario para facilitar tarefas comuns de teste
 *
 * Uso:
 *   node scripts/test-helper.js <comando>
 *
 * Comandos:
 *   coverage-report - Gera e abre relatorio de cobertura
 *   watch-file <filename> - Executa testes de um arquivo especifico em watch mode
 *   clean - Limpa arquivos de teste temporarios
 *   stats - Mostra estatisticas dos testes
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const command = process.argv[2];
const arg = process.argv[3];

// Cores para output no terminal
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  red: '\x1b[31m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function exec(cmd, silent = false) {
  try {
    const output = execSync(cmd, { encoding: 'utf8', stdio: silent ? 'pipe' : 'inherit' });
    return output;
  } catch (error) {
    if (!silent) {
      log(`Erro ao executar: ${cmd}`, 'red');
    }
    return null;
  }
}

// ============================================================================
// Comandos
// ============================================================================

function coverageReport() {
  log('Gerando relatorio de cobertura...', 'blue');
  exec('npm run test:coverage');

  const coveragePath = path.join(__dirname, '..', 'coverage', 'index.html');

  if (fs.existsSync(coveragePath)) {
    log('\nRelatorio gerado com sucesso!', 'green');
    log(`Abrindo: ${coveragePath}`, 'yellow');

    // Tenta abrir o arquivo no navegador padrao
    const openCommand = process.platform === 'win32' ? 'start' :
                       process.platform === 'darwin' ? 'open' : 'xdg-open';
    exec(`${openCommand} ${coveragePath}`, true);
  } else {
    log('Erro: Relatorio de cobertura nao encontrado', 'red');
  }
}

function watchFile() {
  if (!arg) {
    log('Erro: Especifique um arquivo para watch', 'red');
    log('Uso: node scripts/test-helper.js watch-file <filename>', 'yellow');
    process.exit(1);
  }

  log(`Executando testes de ${arg} em watch mode...`, 'blue');
  exec(`npx vitest ${arg}`);
}

function clean() {
  log('Limpando arquivos temporarios...', 'blue');

  const dirsToClean = [
    path.join(__dirname, '..', 'coverage'),
    path.join(__dirname, '..', '.vitest'),
  ];

  dirsToClean.forEach(dir => {
    if (fs.existsSync(dir)) {
      fs.rmSync(dir, { recursive: true, force: true });
      log(`Removido: ${dir}`, 'green');
    }
  });

  log('Limpeza concluida!', 'green');
}

function stats() {
  log('Coletando estatisticas dos testes...', 'blue');

  const testFiles = [];
  const findTestFiles = (dir) => {
    const files = fs.readdirSync(dir);
    files.forEach(file => {
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);

      if (stat.isDirectory() && !file.includes('node_modules') && !file.includes('dist')) {
        findTestFiles(filePath);
      } else if (file.endsWith('.test.ts') || file.endsWith('.test.tsx')) {
        testFiles.push(filePath);
      }
    });
  };

  findTestFiles(path.join(__dirname, '..'));

  log('\n=== Estatisticas dos Testes ===\n', 'bright');
  log(`Total de arquivos de teste: ${testFiles.length}`, 'green');

  let totalTests = 0;
  let totalDescribes = 0;

  testFiles.forEach(file => {
    const content = fs.readFileSync(file, 'utf8');
    const itMatches = content.match(/it\(|test\(/g);
    const describeMatches = content.match(/describe\(/g);

    const numTests = itMatches ? itMatches.length : 0;
    const numDescribes = describeMatches ? describeMatches.length : 0;

    totalTests += numTests;
    totalDescribes += numDescribes;

    const relativePath = path.relative(path.join(__dirname, '..'), file);
    log(`  ${relativePath}`, 'yellow');
    log(`    - ${numTests} testes`, 'reset');
    log(`    - ${numDescribes} describe blocks\n`, 'reset');
  });

  log(`\nTotal de testes: ${totalTests}`, 'bright');
  log(`Total de describe blocks: ${totalDescribes}`, 'bright');

  // Tenta obter cobertura
  log('\nExecutando testes para obter cobertura...', 'blue');
  const coverageOutput = exec('npx vitest run --coverage --reporter=json', true);

  if (coverageOutput) {
    log('Cobertura calculada!', 'green');
  }
}

function listTests() {
  log('Listando todos os arquivos de teste...', 'blue');

  const testFiles = [];
  const findTestFiles = (dir) => {
    const files = fs.readdirSync(dir);
    files.forEach(file => {
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);

      if (stat.isDirectory() && !file.includes('node_modules') && !file.includes('dist')) {
        findTestFiles(filePath);
      } else if (file.endsWith('.test.ts') || file.endsWith('.test.tsx')) {
        testFiles.push(filePath);
      }
    });
  };

  findTestFiles(path.join(__dirname, '..'));

  log('\n=== Arquivos de Teste ===\n', 'bright');
  testFiles.forEach(file => {
    const relativePath = path.relative(path.join(__dirname, '..'), file);
    log(`  ${relativePath}`, 'green');
  });

  log(`\nTotal: ${testFiles.length} arquivos`, 'bright');
}

function help() {
  log('\n=== Test Helper - Comandos Disponiveis ===\n', 'bright');
  log('coverage-report', 'green');
  log('  Gera relatorio de cobertura e abre no navegador\n', 'reset');

  log('watch-file <filename>', 'green');
  log('  Executa testes de um arquivo especifico em watch mode', 'reset');
  log('  Exemplo: node scripts/test-helper.js watch-file validation.test.ts\n', 'yellow');

  log('clean', 'green');
  log('  Remove arquivos temporarios (coverage, .vitest)\n', 'reset');

  log('stats', 'green');
  log('  Mostra estatisticas dos testes (arquivos, numero de testes, etc)\n', 'reset');

  log('list', 'green');
  log('  Lista todos os arquivos de teste no projeto\n', 'reset');

  log('help', 'green');
  log('  Mostra esta mensagem de ajuda\n', 'reset');
}

// ============================================================================
// Main
// ============================================================================

switch (command) {
  case 'coverage-report':
    coverageReport();
    break;

  case 'watch-file':
    watchFile();
    break;

  case 'clean':
    clean();
    break;

  case 'stats':
    stats();
    break;

  case 'list':
    listTests();
    break;

  case 'help':
  default:
    help();
    break;
}
