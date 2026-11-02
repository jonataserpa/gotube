# refresh-session Specification

## Purpose

Autentica a conta, rotaciona o refresh e encerra só a sessão apresentada, sem misturar contas.

## Requirements

### Requirement: Login com credencial
`POST /v1/auth/login` com e-mail e senha corretos de conta ativa MUST responder 200. O JSON MUST trazer access token de vida curta, identificador e e-mail normalizado. O refresh MUST ir só em cookie `HttpOnly` e MUST NOT aparecer no JSON. A resposta MUST NOT incluir o hash da senha.

#### Scenario: Credencial correta
- **WHEN** a senha confere com a conta ativa
- **THEN** a API devolve access token no corpo e refresh só no cookie `HttpOnly`

### Requirement: Credencial indistinta
E-mail inexistente, senha errada ou conta que não está ativa MUST produzir o mesmo 401, sem indicar qual caso ocorreu e sem cookie de sessão novo.

#### Scenario: Senha errada
- **WHEN** o e-mail existe e a senha não confere
- **THEN** a API responde 401 com o mesmo código usado para e-mail desconhecido

### Requirement: Rotação atômica
`POST /v1/auth/refresh` com o refresh vigente e não expirado MUST revogar esse refresh, gravar outro da mesma família e responder 200 com novo access token, identificador e e-mail. O refresh novo MUST ir só no cookie.

#### Scenario: Cookie vigente
- **WHEN** o cookie corresponde a uma sessão não revogada e não expirada
- **THEN** a sessão anterior fica revogada e a resposta traz access token novo e cookie novo

### Requirement: Reuso revoga a família
Apresentar um refresh que aquela família já rotacionou, depois do prazo curto de corrida, MUST revogar todas as sessões da família e responder 401. Dentro desse prazo, a apresentação perdedora MUST responder 401 sem revogar a sessão nova. Um refresh desconhecido MUST responder 401 sem revogar sessões de outras famílias.

#### Scenario: Token já substituído
- **WHEN** o cookie é de um refresh que esta família já rotacionou e o prazo de corrida já passou
- **THEN** a API responde 401 e nenhuma sessão dessa família permanece ativa

#### Scenario: Cookie desconhecido
- **WHEN** o cookie não corresponde a nenhuma sessão
- **THEN** a API responde 401 e sessões de outras contas permanecem ativas

### Requirement: Corrida não revoga a vencedora
Duas renovações simultâneas do mesmo refresh ainda vigente MUST produzir uma única sessão nova. A requisição que perde a corrida MUST receber 401. Essa perda MUST NOT revogar a sessão vencedora.

#### Scenario: Duas renovações juntas
- **WHEN** duas requisições apresentam o mesmo refresh ainda vigente
- **THEN** só uma recebe o cookie novo e a outra recebe 401 com a sessão vencedora ainda ativa

### Requirement: Logout da sessão apresentada
`POST /v1/auth/logout` com refresh vigente MUST revogar essa sessão, limpar o cookie e responder 204. Outras sessões da mesma conta MUST permanecer. Sem cookie vigente, a API MUST responder 204 e limpar o cookie se ele existir.

#### Scenario: Saída de um dispositivo
- **WHEN** a conta encerra a sessão cujo cookie foi apresentado
- **THEN** essa sessão fica revogada, o cookie é limpo e outra sessão da mesma conta continua válida

### Requirement: Identificador não autoriza
Conhecer o identificador de outra conta MUST NOT renovar a sessão, encerrar a sessão nem trocar a senha dessa conta.

#### Scenario: Só o identificador
- **WHEN** a requisição informa o identificador de outra conta sem o refresh ou o token de reset dessa conta
- **THEN** a API não altera a senha nem as sessões dessa conta

### Requirement: Access token deixa de valer no prazo
O access token MUST ser aceito só com assinatura, emissor, audiência e expiração conferidos. Logout MUST NOT ser obrigado a invalidar um access token já emitido antes do prazo. Decodificar o token sem essas conferências MUST NOT autenticar.

#### Scenario: Saída com access ainda no prazo
- **WHEN** a sessão de refresh é revogada e o access token ainda não expirou
- **THEN** o contrato não exige que esse access token passe a ser recusado antes do prazo

### Requirement: Sessão sem banco
Login, refresh ou logout que precisem do banco e não puderem usá-lo MUST responder 503, sem access token novo. `GET /healthz` MUST continuar 200.

#### Scenario: Login sem banco
- **WHEN** o banco não atende o login
- **THEN** a resposta é 503 e não contém access token
