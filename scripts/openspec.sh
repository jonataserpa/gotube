#!/usr/bin/env bash
set -euo pipefail
PROJECT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
CLI="$PROJECT_ROOT/.tooling/openspec/node_modules/.bin/openspec"
if [[ ! -x "$CLI" ]]; then
  echo "OpenSpec local ausente. Execute: bash scripts/setup-openspec.sh" >&2
  exit 1
fi
cd "$PROJECT_ROOT"
exec "$CLI" "$@"
