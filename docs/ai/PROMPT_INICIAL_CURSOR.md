# Prompt inicial — copiar o texto abaixo no Cursor Agent

Você vai me ajudar a desenvolver e aprender com o GoTube, uma plataforma de vídeos
e lives usando Go no backend e Next.js no frontend. Trabalhe neste repositório.

## Contexto a ler
1. AGENTS.md, internal/AGENTS.md e web/AGENTS.md.
2. docs/product.md, docs/architecture.md e docs/roadmap.md.
3. openspec/config.yaml e as specs/changes que já existirem.
4. docs/course/GoTube_Curso_Completo.md: primeiro o mapa, seções 1–3 e 21–23.
   Consulte seções específicas conforme necessário. A seção 27 refina a API da seção 6.
5. docs/ai/workflow.md.

O curso é referência pedagógica, não código pronto. Não leia as 56 aulas como contexto
obrigatório de toda operação e não implemente todo o roadmap de uma vez.

## Tarefa autorizada desta rodada: bootstrap-project
Prepare a proposta OpenSpec e implemente apenas a base executável abaixo.
O escopo e as decisões de partida estão autorizados; avance até o aceite sem pedir
confirmação a cada arquivo. Se surgir uma decisão bloqueante fora desse escopo,
explique as alternativas e peça esclarecimento. Não avance para a próxima feature.

### 1. Diagnóstico e plano
- Inspecione arquivos e Git; preserve alterações existentes.
- Verifique Go, Node, npm, Docker/Compose e OpenSpec e registre versões reais.
- Se faltar OpenSpec, execute bash scripts/setup-openspec.sh, se rede/npm permitirem.
- Se infraestrutura/ferramenta impedir a execução, relate a limitação com precisão.
- Use o workflow OpenSpec instalado para criar bootstrap-project com proposal, specs,
  design e tasks. Não invente um comando que a versão instalada não possui.

### 2. Backend mínimo
- Inicie um módulo Go na raiz com module path local example.com/gotube; ele é provisório,
  não representa um repositório publicado. Documente como trocar quando houver remoto.
- Crie cmd/api e composição mínima em internal/platform.
- Disponibilize GET /healthz retornando JSON e HTTP 200.
- Defina timeout HTTP e graceful shutdown adequados ao serviço inicial.
- Escreva teste relevante do handler e tratamento de configuração inválida.
- Não crie login, repos genéricos, fila ou processamento nesta etapa.

### 3. Frontend mínimo
- Crie aplicação Next.js App Router e TS estrito diretamente em web/.
- Preserve web/AGENTS.md e arquivos existentes. Se create-next-app recusar pasta não vazia,
  gere em diretório temporário e incorpore arquivos sem sobrescrever instruções.
- Use versão estável compatível com o Node instalado e registre versões/lockfile.
- Configure Tailwind conforme a versão escolhida, sem copiar config obsoleta.
- Crie shell GoTube com navegação superior, tema claro/escuro e layout responsivo.
- Exiba estado vazio honesto de catálogo ainda não implementado, sem funcionalidades falsas.
- Configure teste de componente significativo e teste de fumaça da página.
- Não instale todas as libs do roadmap antecipadamente; adicione quando usadas.

### 4. Infraestrutura e contrato
- PostgreSQL em Docker Compose para desenvolvimento, porta local e volume persistente.
- Credenciais locais via .env a partir de .env.example, sem versionar segredos reais.
- Sem Kubernetes, Redis, RabbitMQ, storage ou MediaMTX neste bootstrap.
- Crie api/openapi.yaml válido apenas para o endpoint implementado; planos futuros
  permanecem nas specs/roadmap, não como endpoints falsamente disponíveis.
- Faça teste de integração simples da página/shell e API quando ambiente permitir.

### 5. Operação e qualidade
- README com comandos reais para subir API, web e PostgreSQL, testar e encerrar.
- Scripts reais de typecheck, lint, teste e build no package.json do frontend.
- CI em .github/workflows/ para Go e Next, com versões compatíveis e lockfile.
- Documente endpoint de health e diferença de liveness versus readiness futura.
- Execute go test ./..., go vet ./... e verificação de formatação.
- Execute typecheck, lint, testes e build do frontend e teste de fumaça aplicável.
- Teste Compose com validação de config; se Docker estiver disponível, suba e verifique saúde.
- Não diga que um teste passou se ele não foi executado. Não faça push/merge/deploy.

## Aceite
- API sobe localmente e /healthz responde conforme OpenAPI.
- Next abre com layout acessível e sem erro de hidratação.
- PostgreSQL local possui instruções reproduzíveis de execução.
- Verificações reais e limitações registradas.
- Tasks OpenSpec refletem evidência; nada é marcado pronto só porque o arquivo existe.
- Atualize docs/learning/progress.md e relate próximo passo: public-video-catalog.

## Forma de ensinar
Explique em português o problema, conceito, local de execução e uma falha relevante.
Prefira uma fatia funcional pequena e explique decisões, sem narrar cada linha.
Ao concluir, informe comandos de execução, resultados reais e um exercício curto.
Execute agora o diagnóstico e conduza bootstrap-project até esse aceite.
