# Frontend Next.js

- App Router e TypeScript estrito; app/ compõe páginas, features/ organiza comportamento.
- Fronteiras server/client explícitas. Não adicionar use client à árvore inteira.
- SSR inicial consulta dados públicos; sessão privada é restaurada no browser.
- Next.js não duplica autenticação de negócio nem acessa tabelas do backend.
- TanStack Query para dados remotos; RHF/Zod para formulário; URL para filtros.
- Zustand somente para estado visual transversal necessário.
- Access token em memória; refresh HttpOnly no Go; não persistir JWT em localStorage.
- Limpe cache privado, cancele requests e descarte respostas antigas ao trocar sessão.
- Não copie dados remotos em várias stores concorrentes.
- Segredos nunca em NEXT_PUBLIC_, props RSC para cliente ou logs.
- Cleanup obrigatório de player, socket, object URLs, listeners e timers.
- Limite concorrência de arquivos e partes no upload direto ao storage.
- Toda tela trata loading, vazio, erro, pending e recuperação aplicável.
- Detalhes e edição em páginas próprias; modais apenas para ações curtas.
- Tema claro/escuro, teclado, foco visível, contraste e movimento reduzido.
- Use scripts reais do package.json; gere lockfile e registre versões adotadas.
- Não crie nova app em web/web nem sobrescreva este arquivo ao usar geradores.

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
