#!/usr/bin/env bash
set -euo pipefail
STARTER_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
command -v python3 >/dev/null || { echo 'Python 3 é necessário para preparar os arquivos.' >&2; exit 1; }
exec python3 "$STARTER_ROOT/bootstrap.py" "$@"
