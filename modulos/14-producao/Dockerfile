# Imagem em duas etapas: uma para compilar, outra para rodar.
#
# A etapa de compilação carrega o Go inteiro, com compilador, biblioteca padrão
# e cache de módulos, e passa dos setecentos megabytes. Nada disso precisa ir
# junto para produção: um binário Go é estático, então a imagem final leva o
# arquivo e mais nada.

# ---------- etapa 1: compilar ----------
FROM golang:1.24-alpine AS construcao

WORKDIR /origem

# go.mod e go.sum primeiro, sozinhos. O Docker guarda em cache o resultado de
# cada instrução, e essa camada só é refeita quando as dependências mudam. Se
# você copiasse o projeto inteiro antes, cada edição de uma linha de código
# baixaria tudo de novo.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Qual programa compilar. O padrão é o do módulo, que é o seu; o CI deste
# repositório aponta para a solução, porque precisa de um binário que funcione
# para conseguir testar a imagem de ponta a ponta.
ARG CAMINHO=./modulos/14-producao/cmd/servidor

# CGO_ENABLED=0 é o que produz um binário verdadeiramente estático, sem
# depender da libc da imagem base. É por isso que o driver de banco do módulo
# 11 é escrito em Go puro: um driver com cgo obrigaria a outra imagem base e a
# outra conversa inteira.
#
# -trimpath tira os caminhos da sua máquina do binário.
# -ldflags "-s -w" descarta a tabela de símbolos e o DWARF, e tira alguns
# megabytes. Em troca, um perfil de execução fica menos legível.
RUN CGO_ENABLED=0 GOOS=linux go build \
	-trimpath \
	-ldflags="-s -w" \
	-o /padaria \
	"$CAMINHO"

# ---------- etapa 2: rodar ----------
# scratch é a imagem vazia: zero byte, zero pacote, zero vulnerabilidade
# herdada. Não há shell, não há gerenciador de pacotes e não há nada para um
# invasor usar depois de entrar.
#
# O preço é que também não há como depurar de dentro. Quando isso incomodar,
# gcr.io/distroless/static é o meio-termo: continua sem shell, mas traz
# certificados e fuso horário.
FROM scratch

# Certificados raiz, para o binário conseguir falar HTTPS com outros serviços.
# Sem esta linha, qualquer chamada TLS falha com erro de autoridade
# desconhecida, e a mensagem não deixa nada óbvio.
COPY --from=construcao /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=construcao /padaria /padaria

# Usuário sem privilégio. Em scratch não há /etc/passwd, então o número é
# usado direto. Rodar como root dentro do contêiner é hábito, não necessidade.
USER 65534:65534

EXPOSE 8080

ENV PADARIA_ENDERECO=":8080" \
	PADARIA_LOG_FORMATO="json" \
	PADARIA_LOG_NIVEL="info"

# ENTRYPOINT na forma de lista, e não de texto. Na forma de texto o Docker
# embrulha o comando num shell, e o shell fica sendo o processo 1: os sinais
# vão para ele, não para o seu programa, e o desligamento com graça do módulo
# 13 simplesmente nunca acontece.
ENTRYPOINT ["/padaria"]
