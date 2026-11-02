# Backend Go

- Go idiomático, composição e interfaces pequenas no consumidor quando necessárias.
- Domínio não depende de HTTP, ORM ou broker.
- Context nas fronteiras de aplicação/I/O; não o use como recipiente de dependências.
- Erros explícitos e wrapping; não usar panic para entrada inválida do cliente.
- Toda goroutine tem responsável, limite e término previsível.
- Trabalho que sobrevive a restart exige persistência; channel não é fila durável.
- Constraints e transações protegem invariantes; mutex local não coordena réplicas.
- Eventos duráveis usam outbox; consumers toleram duplicação e redelivery.
- Nunca segure transação durante FFmpeg, upload ou chamada remota demorada.
- Reavalie autorização por recurso no caso de uso, mesmo que a UI esconda a ação.
- Use argumentos separados para processos externos, sem shell interpolado.
- Teste lógica de domínio, falhas, SQL real nas integrações e concorrência relevante.
- Use race detector em alterações concorrentes quando a plataforma suportar.
- gofmt e go vet fazem parte dos gates após existir módulo compilável.
