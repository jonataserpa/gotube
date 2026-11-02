# Verificação do pacote — 14/10/2026

Executado nesta entrega:

- Parsing do Python de bootstrap e bash -n em todos os scripts shell.
- Geração offline completa em pasta temporária com espaços no nome.
- Invocação do script a partir de outro diretório.
- Comparação SHA-256 do curso copiado com a versão full stack recebida.
- Contagem de 32 módulos Go e 24 módulos frontend no curso incluído.
- Recusa de segunda execução sobre destino preenchido, preservando edição do usuário.
- Recusa de destino dentro do próprio pacote.
- Inicialização Git local na branch main sem commit ou remoto.
- Conferência estrutural de frontmatter nas três regras MDC.

Não executado nesta entrega:

- Download/instalação npm e execução real do OpenSpec.
- Teste de reconhecimento de regras/comandos no aplicativo Cursor.
- Criação/compilação da aplicação Go ou Next.js.
- Inicialização de PostgreSQL/Docker ou CI remoto.

Esses itens dependem do ambiente do usuário. O prompt inicial orienta sua implementação
e validação sem alegar resultados inexistentes. O pacote fornece contexto e inicialização
documental; a primeira sessão do Cursor implementará bootstrap-project.
