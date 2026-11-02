# Ordem de implementação

Cada item é uma mudança separada. O prompt inicial autoriza somente bootstrap-project.

1. bootstrap-project: Go health, Next shell, PostgreSQL local, contrato mínimo, scripts e CI.
2. public-video-catalog: dados públicos, paginação, busca e página Next com estados.
3. identity-session: cadastro/login/refresh/logout/recuperação e isolamento de sessão.
4. channel-management: perfil, canal e categorias autorizadas.
5. video-upload: rascunho, autorização de storage, progresso e confirmação idempotente.
6. video-processing: outbox, fila, worker, FFmpeg, thumbnails e HLS.
7. video-playback: autorização, player, privacidade e edição/publicação/exclusão.
8. social-interactions: comentários, likes e inscrições.
9. personal-library: histórico, progresso e playlists.
10. live-streaming: controle Go, OBS, MediaMTX e player.
11. live-chat: ticket, WebSocket, reconexão, moderação e limites.
12. administration: notificações, denúncias, categorias e auditoria.
13. operations-hardening: profiling, carga, recuperação e deploy.

Testes, acessibilidade e documentação acompanham cada entrega; não ficam só no último item.
O curso completo orienta os conceitos de cada etapa. Não marque capacidade como pronta por
existir uma pasta, fixture, mock ou tela demonstrativa.
