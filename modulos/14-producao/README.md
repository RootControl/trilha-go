# Módulo 14 — Produção

Último módulo. A padaria já funciona; aqui ela fica pronta para rodar em algum
lugar que não seja a sua máquina.

```bash
go test ./modulos/14-producao/...
go run ./modulos/14-producao/cmd/servidor
```

```bash
curl -s localhost:8080/saude
curl -s localhost:8080/versao
PADARIA_LOG_FORMATO=json PADARIA_LOG_NIVEL=debug go run ./modulos/14-producao/cmd/servidor
```

## Por que Go faz assim

**Um arquivo, e mais nada.**
`go build` produz um binário estático. Não há interpretador, não há
`node_modules`, não há ambiente virtual, não há runtime para instalar no
servidor. Você copia o arquivo e ele roda.

Compilar para outro sistema são duas variáveis:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o padaria ./modulos/14-producao/cmd/servidor
```

Isso é o que torna a imagem de contêiner deste módulo tão pequena, e é uma das
razões práticas de Go ter virado a linguagem padrão de ferramenta de
infraestrutura.

**Configuração vem do ambiente, e é validada na partida.**
`CarregarConfig` lê cinco variáveis, aplica padrões e **recusa** qualquer valor
que não sirva. O programa morre ali, com uma mensagem que diz qual variável
está errada:

```
padaria: PADARIA_TRABALHADORES="zero" não é um número: configuração inválida
```

A alternativa é um serviço que sobe, parece saudável, e falha três horas
depois na primeira requisição que toca no valor errado. Falhar cedo e alto é
sempre melhor.

E repare que o ambiente **vazio** produz uma configuração utilizável. Serviço
que não sobe sem seis variáveis definidas é serviço que ninguém consegue rodar
na própria máquina.

**O ambiente entra como parâmetro.**
`CarregarConfig` recebe uma `Ambiente`, que tem a mesma assinatura de
`os.LookupEnv`. É a mesma ideia do relógio do módulo 06 e do banco do módulo
11: o teste entrega um mapa e decide o que o programa enxerga, sem mexer no
ambiente da máquina e sem que dois testes rodando juntos briguem por uma
variável global. O `main` passa o `os.LookupEnv` de verdade.

**Log estruturado.**
`log/slog` está na biblioteca padrão desde o Go 1.21, e escreve pares
chave-valor em vez de frases:

```json
{"time":"...","level":"INFO","msg":"requisição","metodo":"GET","caminho":"/saude","status":200,"duracao_ms":0}
```

A diferença aparece quando alguém precisa achar algo. Com texto solto, você
escreve expressão regular. Com campos, você filtra por `status >= 500` e
pronto. Em desenvolvimento o formato de texto lê melhor; em produção, JSON,
porque quem lê é uma máquina. É por isso que o formato é configurável.

**`LogAttrs` em vez de `Info` no caminho quente.**
As duas formas funcionam. `log.Info("msg", "chave", valor)` é conveniente e
embrulha cada valor num `any`, o que aloca. `log.LogAttrs` com `slog.String` e
`slog.Int` é mais verboso e não aloca. Num middleware que roda em toda
requisição, vale a verbosidade; num log de partida, não faz diferença.

**Log vai para o erro padrão.**
A saída padrão continua livre para dados, como no módulo 08. Um coletor de
logs que trata os dois canais separadamente não mistura as coisas.

**Versão sem `-ldflags`.**
O truque de injetar a versão na compilação com `-ldflags "-X main.versao=..."`
ainda funciona e ainda aparece em todo tutorial. Desde o Go 1.18 ele é
desnecessário para o caso comum: o compilador grava os dados do controle de
versão dentro do binário, e `runtime/debug.ReadBuildInfo` lê de volta. O
binário deste módulo se identifica assim:

```
{"versao":"v0.0.0-20260908020549-ca9f60e5a21d+dirty"}
```

O sufixo `+dirty` quer dizer que havia alteração não commitada na hora de
compilar, e é uma informação que vale ouro quando alguém pergunta qual código
exatamente está rodando.

Com uma ressalva importante, e ela está nas pegadinhas: isso só funciona
quando o compilador enxerga o repositório.

## A imagem de contêiner

O `Dockerfile` está em `modulos/14-producao/Dockerfile`, e as instruções de uso
em [README-imagem.md](README-imagem.md).

Ele tem duas etapas. A primeira carrega o Go inteiro, compilador e cache de
módulos, e passa dos setecentos megabytes. A segunda parte de `scratch`, a
imagem **vazia**, e recebe só o binário e os certificados raiz.

Quatro decisões que valem explicação.

**`CGO_ENABLED=0`** é o que produz um binário verdadeiramente estático, sem
depender da biblioteca C da imagem base. Sem isso, o binário não roda em
`scratch`. Este curso pôde tomar essa decisão porque o driver de banco do
módulo 11 é escrito em Go puro: um driver com cgo forçaria outra imagem base e
outra conversa inteira.

**`scratch`** não tem shell, não tem gerenciador de pacotes e não tem nada que
um invasor possa usar depois de entrar. O preço é que também não há como
depurar de dentro. Quando isso incomodar, `gcr.io/distroless/static` é o
meio-termo.

**`USER 65534:65534`** faz o processo rodar sem privilégio. Root dentro de
contêiner é hábito, não necessidade.

**`ENTRYPOINT` em forma de lista**, e essa é a que mais morde. Escrito como
texto, o Docker embrulha o comando num shell, e o shell vira o processo 1: os
sinais vão para ele e não para o seu programa. O desligamento com graça do
módulo 13 simplesmente nunca acontece, e cada implantação corta as requisições
em andamento.

O CI deste repositório constrói a imagem, sobe o contêiner, consulta a saúde e
o desliga com `docker stop`. Um Dockerfile que ninguém constrói apodrece igual
a código que ninguém compila.

## Pegadinhas

**`ENTRYPOINT` como texto mata o desligamento com graça.**
Já explicado acima, e repetido aqui porque é o erro mais caro do módulo: ele
não quebra nada visivelmente, só faz o serviço perder requisições a cada
implantação.

**Faltar `ca-certificates.crt` em `scratch`.**
Qualquer chamada HTTPS falha com erro de autoridade desconhecida, e a mensagem
não deixa a causa óbvia. A imagem base cheia trazia isso de graça; a vazia,
não.

**Copiar o projeto inteiro antes do `go mod download`.**
O Docker guarda em cache o resultado de cada instrução. Com `go.mod` e
`go.sum` copiados sozinhos primeiro, a camada de dependências só é refeita
quando elas mudam. Na ordem errada, cada edição de uma linha baixa tudo de
novo.

**Ler configuração no meio do programa.**
`os.Getenv` espalhado pelo código dá um serviço que sobe bem e falha na
primeira requisição que toca no valor errado. Leia tudo na partida, valide
tudo na partida, e depois passe a `Config` adiante.

**Chamar `os.Getenv` direto numa função que você quer testar.**
Testar isso obriga a mexer no ambiente do processo, o que faz dois testes em
paralelo brigarem. Receba a leitura como parâmetro.

**Log no `os.Stdout`.**
Mistura recado com dado e quebra qualquer cano, exatamente como no módulo 08.

**Sem `.git` no contexto, a versão some.**
Esta apareceu construindo a imagem deste próprio módulo. O `.dockerignore`
exclui `.git`, porque copiar o histórico inteiro para dentro do contexto de
construção é desperdício. Só que é dele que o compilador tira a revisão, então
o binário de dentro do contêiner se identifica como `desenvolvimento`, e não
com a pseudoversão que o mesmo código produz na sua máquina.

Confira você mesmo:

```bash
go run ./modulos/14-producao/cmd/servidor    # versão com revisão
docker run --rm padaria:local                # versão "desenvolvimento"
```

Há duas saídas. Tirar `.git` do `.dockerignore`, e pagar em contexto de
construção maior. Ou passar a versão explicitamente na compilação, que é
justamente o caso em que o velho `-ldflags` continua sendo a resposta certa:

```dockerfile
ARG VERSAO=desenvolvimento
RUN go build -ldflags="-s -w -X main.versao=$VERSAO" ...
```

```bash
docker build --build-arg VERSAO="$(git describe --tags --always --dirty)" ...
```

O que fica de lição não é qual das duas escolher: é que a informação de
versão depende de o compilador enxergar o repositório, e um contexto de
construção enxuto é exatamente o lugar onde ele não enxerga.

## O exercício

Quatro implementações em `padaria/padaria.go`:

1. `CarregarConfig`, com padrões, validação e mensagens que dizem qual
   variável está errada.
2. `NovoLogger`, escolhendo entre texto e JSON, com o nível da configuração.
3. `ComLog`, o middleware do módulo 10 com campos estruturados.
4. `Versao`, lendo o que o compilador já gravou no binário.

Comece por `CarregarConfig`: ela é a maior, é pura, e os testes dela cobrem
seis formas de errar.

Depois de tudo passando:

```bash
make binario
PADARIA_LOG_FORMATO=json ./bin/padaria
```

## Desafio, sem teste pronto

**Separe saúde de prontidão.** `/saude` responde se o processo está vivo;
`/pronto` deveria responder se ele consegue **atender**, o que inclui o banco
estar acessível. Um orquestrador usa os dois de formas diferentes: reiniciar
um processo morto, e tirar do balanceamento um processo vivo mas sem banco.

**Publique de verdade.** O CI deste repositório compila para quatro
combinações de sistema e arquitetura a cada envio. Transforme isso num fluxo
que dispara em etiqueta de versão e anexa os binários a uma versão do GitHub.
Repare que, com uma etiqueta `v1.0.0`, a `Versao()` que você escreveu passa a
devolver `v1.0.0` sozinha, sem nenhuma mudança de código.

**Conserte a versão no contêiner.** Escolha uma das duas saídas descritas nas
pegadinhas e faça `curl localhost:8080/versao` devolver a versão de verdade
dentro do contêiner.

**Meça a imagem.** A imagem final deste módulo, medida no CI, tem 6,09 MB.
Compare com o tamanho da etapa de construção, e depois compare com `golang:1.24-alpine` como base final em vez de
`scratch`. A diferença é a razão de existir a construção em duas etapas.

**Ligue tudo.** Junte o servidor do módulo 10, o banco do 11 e a fila do 13
neste `main`. É o exercício mais longo do curso e é o único que produz algo que
você poderia colocar no ar.

## Fim da trilha

Quinze módulos, do 00 ao 14, e a padaria saiu de uma função que cumprimenta cliente até um
serviço em contêiner com log estruturado e desligamento ordenado.

Se você chegou até aqui implementando, e não lendo, você escreve Go. O que
falta agora não se aprende em curso: é ler código dos outros, é abrir a
biblioteca padrão para ver como ela resolve as mesmas coisas, e é manter um
programa seu em produção tempo suficiente para ele te ensinar o que faltava.

## Se travar

A solução está em `solucoes/14-producao/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
