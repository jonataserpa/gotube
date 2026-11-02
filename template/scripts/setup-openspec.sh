#!/usr/bin/env bash
set -euo pipefail
PROJECT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"
command -v node >/dev/null || { echo "Instale Node.js compatível e tente novamente." >&2; exit 1; }
command -v npm >/dev/null || { echo "npm não encontrado." >&2; exit 1; }
node -e 'const [a,b]=process.versions.node.split(".").map(Number); if(a<20||(a===20&&b<19)){console.error("OpenSpec requer Node >=20.19; confira também requisito Next."); process.exit(1)}'
if [[ $# -gt 1 ]]; then
  echo "Uso: bash scripts/setup-openspec.sh [versao-openspec]" >&2
  exit 2
fi
if [[ $# -eq 1 ]]; then
  npm install --prefix "$PROJECT_ROOT/.tooling/openspec" --save-dev --save-exact "@fission-ai/openspec@$1"
elif [[ -f "$PROJECT_ROOT/.tooling/openspec/package-lock.json" ]]; then
  npm ci --prefix "$PROJECT_ROOT/.tooling/openspec"
else
  # Primeira resolução; save-exact + lockfile fixam a versão efetivamente obtida.
  npm install --prefix "$PROJECT_ROOT/.tooling/openspec" --save-dev --save-exact '@fission-ai/openspec@latest'
fi
bash "$PROJECT_ROOT/scripts/openspec.sh" --version
bash "$PROJECT_ROOT/scripts/openspec.sh" init --tools cursor --profile core
mkdir -p "$PROJECT_ROOT/docs/ai"
bash "$PROJECT_ROOT/scripts/openspec.sh" --version > "$PROJECT_ROOT/docs/ai/openspec-version.txt"
echo "OpenSpec inicializado para Cursor. Revise e versione package.json, lockfile e arquivos gerados."
echo "Abra nova sessão do Cursor e confira os comandos/skills instalados."
