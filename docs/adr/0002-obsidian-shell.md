# ADR 0002 — Shell visual Obsidian do GoTube

- Status: aceito em 14/10/2026, antes de `public-video-catalog`.
- Evidência visual: referência Stitch em `code.html` e o sistema Obsidian Kinetic. A captura mostra barra superior, menu lateral, chips e grade escura.
- Spec relacionada: o shell de `bootstrap-project` ganha esta forma. A listagem continua na mudança `public-video-catalog`.

## Contexto

O bootstrap entregou um cabeçalho simples, claro por padrão. A referência pedida é um shell de vídeo em camadas escuras, com navegação lateral, busca no topo e acento vermelho só no que é ação. O curso pedia navegação superior. A referência muda essa composição.

O HTML de exemplo usa a marca NextTube, fotos externas e um feed inventado (títulos, audiência, canais e ao vivo). Isso não é dado do GoTube.

## Alternativas

- Copiar o HTML com os vídeos e as métricas de exemplo. A página pareceria pronta e mentiria sobre catálogo, sessão e reprodução.
- Manter o cabeçalho claro do bootstrap. A decisão visual nova ficaria só no papel.
- Adotar o shell escuro e preencher a área principal só com o que a etapa realmente tem.

## Decisão

O GoTube usa o shell Obsidian, escuro por padrão, com tema claro ainda disponível.

- Canvas `#131313`, superfície `#1c1b1b`, texto `#e5e2e1`, metadado `#a1a1aa`.
- Ação primária `#ff1a3c` com texto branco. O vermelho não pinta o fundo da página.
- Borda de 1px em baixa opacidade. Sem glow espalhado.
- Geist para interface e JetBrains Mono para duração e atalhos, via `next/font`.
- Desktop: barra fixa de 56px e lateral de 256px. Abaixo de `lg`, a lateral abre pelo botão de menu.
- Chips em pílula. Só “Tudo” fica ativo enquanto não existe filtro.
- Cantos: controles em 8px, cartões em 16px.

A marca é GoTube. Não se usa NextTube, avatar de terceiros nem URL de imagem externa.

Busca, microfone, novo vídeo, notificações, conta e destinos laterais além de Início aparecem desabilitados, com nome acessível, até a etapa que os implementa. A área principal continua dizendo que o catálogo não está implementado. Inscrições mostram ausência real, não canais fictícios.

## Consequências

A home deixa de ser uma coluna central isolada. A mudança `public-video-catalog` preenche esta área com a grade, sem redesenhar o chrome. O tema padrão passa a ser escuro; a escolha `light` no `localStorage` continua válida.

Filtros, busca e atalhos visíveis não executam consulta. Habilitar um deles exige a spec correspondente.

## Critério de revisão

Rever quando a página de assistir, a busca ou o menu mobile precisarem de um padrão diferente do shell. Até lá, cartão, chip e botão novo reutilizam estes tokens.
