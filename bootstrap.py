#!/usr/bin/env python3
"""Prepara contexto e instruções GoTube. Não gera uma aplicação ficticiamente pronta."""
import argparse
import json
import shutil
import subprocess
import sys
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description='Inicia scaffold GoTube + curso + regras Cursor.')
    parser.add_argument('destination', nargs='?', default='gotube', help='Pasta nova ou vazia (padrão: gotube)')
    parser.add_argument('--openspec', action='store_true', help='Instalar CLI local via npm e inicializar integração Cursor (rede necessária)')
    parser.add_argument('--openspec-version', help='Release específica; implica --openspec. Ausente: primeira instalação resolve latest e fixa no lockfile.')
    parser.add_argument('--no-git', action='store_true', help='Não executar git init')
    args = parser.parse_args()
    bundle = Path(__file__).resolve().parent
    source = bundle / 'template'
    raw_target = Path(args.destination).expanduser()
    if raw_target.is_symlink():
        parser.error('Destino não pode ser um link simbólico.')
    target = raw_target.resolve()
    if target == bundle or bundle in target.parents or target in bundle.parents:
        parser.error('Escolha uma pasta fora do pacote, por exemplo ../gotube.')
    if target.exists() and (not target.is_dir() or any(target.iterdir())):
        parser.error('Destino deve ser novo ou vazio. Nenhum arquivo existente será sobrescrito.')
    if not args.no_git and not shutil.which('git'):
        parser.error('Git não encontrado. Instale Git ou use --no-git.')
    install = args.openspec or args.openspec_version is not None
    if install and not all(shutil.which(x) for x in ('node', 'npm', 'bash')):
        parser.error('--openspec requer Node.js, npm e bash. Sem a opção, o scaffold é offline.')

    required = ['AGENTS.md', 'internal/AGENTS.md', 'web/AGENTS.md',
                'docs/course/GoTube_Curso_Completo.md', 'docs/ai/PROMPT_INICIAL_CURSOR.md',
                'openspec/config.yaml', 'scripts/setup-openspec.sh']
    for name in required:
        if not (source / name).is_file():
            parser.error('Pacote incompleto: falta template/' + name)
    templates = sorted(source.rglob('*'))
    if any(p.is_symlink() for p in templates):
        parser.error('O template deve conter apenas arquivos e diretórios, sem symlinks.')
    directories = json.loads((bundle / 'directories.json').read_text(encoding='utf-8'))
    for name in directories:
        rel = Path(name)
        if rel.is_absolute() or '..' in rel.parts:
            parser.error('Caminho inválido no manifesto de diretórios.')

    target.mkdir(parents=True, exist_ok=True)
    for src in templates:
        dst = target / src.relative_to(source)
        if src.is_dir():
            dst.mkdir(parents=True, exist_ok=True)
        elif src.is_file():
            dst.parent.mkdir(parents=True, exist_ok=True)
            # Exclusive create: nunca sobrescrever arquivo criado concorrentemente.
            with src.open('rb') as inp, dst.open('xb') as out:
                shutil.copyfileobj(inp, out)
            dst.chmod(0o755 if src.suffix == '.sh' else 0o644)
    for name in directories:
        (target / name).mkdir(parents=True, exist_ok=True)

    if not args.no_git:
        subprocess.run(['git', 'init', '-b', 'main', str(target)], check=True)
    if install:
        command = ['bash', str(target / 'scripts/setup-openspec.sh')]
        if args.openspec_version:
            command.append(args.openspec_version)
        try:
            subprocess.run(command, cwd=target, check=True)
        except subprocess.CalledProcessError as error:
            print('Scaffold preservado em: ' + str(target), file=sys.stderr)
            print('OpenSpec não terminou. Corrija a causa e execute scripts/setup-openspec.sh nessa pasta.', file=sys.stderr)
            return error.returncode or 1

    print('\nGoTube preparado em: ' + str(target))
    print('Abra ESSA pasta no Cursor (File > Open Folder).')
    if not install:
        print('Para gerar integração OpenSpec: bash scripts/setup-openspec.sh')
    print('Cole no Agent: docs/ai/PROMPT_INICIAL_CURSOR.md')
    print('Curso incluído: docs/course/GoTube_Curso_Completo.md')
    print('Aplicação Go/Next será implementada pelo Cursor na etapa bootstrap-project.')
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (OSError, subprocess.CalledProcessError, ValueError) as error:
        print('Falha ao preparar o projeto: ' + str(error), file=sys.stderr)
        print('Revise o destino; arquivos eventualmente criados foram preservados.', file=sys.stderr)
        sys.exit(1)
