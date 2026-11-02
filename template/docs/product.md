# GoTube — produto e limites

Status: especificação inicial; funcionalidades ainda não implementadas pelo starter.

Plataforma educacional inspirada em funcionalidades básicas de vídeo, com identidade própria.
Visitante busca e assiste conteúdo público. Conta autenticada comenta, curte, mantém histórico
e playlists. Criador administra canal, envia/edita/remove vídeos e inicia lives.
Moderador/admin executa operações autorizadas e auditadas.

## Escopo incremental
Cadastro/login/refresh; perfil/canal/categorias; catálogo; upload; processamento HLS;
reprodução; comentários/likes/inscrições; histórico/playlists; Studio; live/chat;
notificações/moderação; observabilidade e deploy.

## Fora da primeira versão
Content ID, DRM, monetização, recomendação ML, escala global, editor de vídeo no browser,
live multiparty e microsserviços por módulo.

## Fontes
Curso: seções 2, 5, 6, 21–23 e 27. A seção 27 refina o contrato inicial da seção 6.
Não copie todos os endpoints para o código antes da etapa correspondente.
Uma spec de capacidade aprovada e o OpenAPI atualizado detalham cada entrega.
