# Tasks

## 1. Schema e seed

- [x] 1.1 Criar as tabelas mínimas de canal e vídeo e o comando explícito de migration; verificar que o comando aplica o schema num PostgreSQL local e não roda no boot da API
- [x] 1.2 Adicionar o seed de desenvolvimento, com cabeçalho que o identifica como fixture, incluindo um vídeo público publicado e ao menos um excluído pela listagem; verificar a marcação no arquivo e que o comando é separado da migration

## 2. Listagem Go

- [x] 2.1 Escrever o teste de integração que exige PostgreSQL e prova a omissão de rascunho e privado, o empate de `published_at` e o 503 sem banco; verificar que o teste falha antes da query existir
- [x] 2.2 Implementar `GET /v1/videos` com DTO público, `limit` e cursor opaco, mantendo `/healthz` sem banco; verificar com o teste de integração e um teste de `limit` inválido e cursor malformado
- [x] 2.3 Documentar `DATABASE_URL` no exemplo de ambiente e o comando de migration/seed no README; verificar que o exemplo não contém senha de produção e que `/healthz` continua descrito como liveness

## 3. Contrato HTTP

- [x] 3.1 Acrescentar somente `GET /v1/videos` em `api/openapi.yaml`, com a página, o item público e os erros 400 e 503; verificar com o teste que impede rota de autenticação ou reprodução e com o lint do OpenAPI

## 4. Página inicial

- [x] 4.1 Escrever testes da home para grade, thumbnail ausente, lista vazia e erro da API; verificar que falham enquanto a página ainda diz que o catálogo não existe
- [x] 4.2 Implementar a leitura de servidor via `GO_API_INTERNAL_URL`, o cursor na URL, `loading.tsx` e os três estados, sem link de reprodução; verificar com o script de teste do `web/` e com typecheck, lint e build
- [x] 4.3 Atualizar `web/.env.example` com a URL interna de desenvolvimento; verificar que a variável não usa prefixo `NEXT_PUBLIC_`

## 5. Integração

- [x] 5.1 Incluir PostgreSQL no job de CI que roda o teste de listagem; verificar que o workflow sobe o banco antes de `go test` e que o teste local contra o Compose passa
- [x] 5.2 Subir API e home e percorrer lista com o seed, página seguinte, banco parado na listagem e `/healthz` ainda 200; registrar os comandos e as respostas reais em `docs/learning/progress.md`

## Workflow follow-up

- Revisar esta proposta antes de implementar. A aplicação começa só com `/opsx-apply` ou pedido explícito.
- Ao arquivar `bootstrap-project`, retirar ou atualizar a exigência de `public-shell` que manda a home dizer que o catálogo não existe.
- Filtros `q`, `category_id` e `channel_id` ficam para uma mudança posterior.
