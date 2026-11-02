# Proposal

## Why

A conta autenticada ainda não tem perfil nem canal próprio. A etapa 4 do roadmap pede esses dados, e a página pública do canal, antes de upload. O catálogo já mostra o canal de laboratório, que não tem dono.

## What Changes

- `GET /v1/me` e `PATCH /v1/me` leem e editam o nome de exibição da conta autenticada. O e-mail continua o da conta e não muda aqui.
- Cada conta pode criar no máximo um canal, com handle único, nome e descrição. `GET /v1/me/channel` devolve esse canal ou ausência explícita. `PATCH` só altera o canal da própria conta.
- `GET /v1/channels/{handle}` devolve só metadados públicos. O canal de laboratório do seed permanece sem dono e continua legível.
- `GET /v1/categories` lista só categorias ativas. O dono grava no próprio canal um conjunto dessas categorias. Identificador inativo ou desconhecido não grava nada.
- `GET /v1/videos` aceita `channel_id` e continua devolvendo apenas vídeos públicos publicados. A home não ganha busca nem filtro de categoria.
- A página do canal mostra os metadados públicos e a grade daquele canal. O card da home passa a apontar para essa página, sem página de reprodução.
- Com sessão, o shell oferece o perfil real. Busca e Novo vídeo continuam indisponíveis.

Suposições: um canal por conta; sem upload de avatar, porque ainda não há object storage; o CRUD administrativo de categorias fica na etapa de administração; categoria pertence ao canal, não ao vídeo.

Fora deste escopo: alterar e-mail ou senha, papel administrativo, inscrição, Studio, rascunho, filtros `q` e `category_id`, e qualquer URL de mídia informada pelo cliente.

## Capabilities

### New Capabilities

- `account-profile`: consultar e editar o nome de exibição da conta autenticada.
- `channel-profile`: criar e editar o único canal da conta e expor a leitura pública por handle.
- `active-categories`: listar categorias ativas e permitir que o dono as associe só ao próprio canal.

### Modified Capabilities

- `public-video-list`: a listagem pública aceita `channel_id` sem alargar o que pode aparecer.
- `catalog-home`: o card aponta para a página pública do canal que já existe.
- `public-shell`: a sessão abre o perfil real; busca e Novo vídeo permanecem desligados.

## Impact

- Go: rotas autenticadas com o access token já emitido, no `ServeMux` existente. Sem token válido, a mutação e a leitura privada respondem 401. Banco indisponível continua 503, e `GET /healthz` continua 200.
- Next.js: páginas de perfil e de canal público. A consulta pública do canal segue no servidor, sem credencial. O formulário da conta usa a sessão já restaurada no browser.
- Dados: migration com nome de exibição, dono opcional do canal, descrição, categorias e o vínculo canal-categoria. O seed de desenvolvimento ganha categorias ativas e não passa a ter dono no canal de laboratório.
- Operação: sem fila, worker, storage nem e-mail. `api/openapi.yaml` ganha só as rotas desta etapa.
