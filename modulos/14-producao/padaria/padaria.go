// Package padaria guarda o código da Padaria do Seu Zé.
//
// Último módulo. A padaria já funciona; aqui ela fica pronta para rodar em
// algum lugar que não seja a sua máquina: configuração vinda do ambiente, log
// que uma máquina consegue ler, e um binário que cabe num contêiner.
package padaria

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// respostaComStatus embute um http.ResponseWriter e lembra o status enviado,
// como no módulo 10.
type respostaComStatus struct {
	http.ResponseWriter
	status int
}

// WriteHeader guarda o status e repassa.
func (r *respostaComStatus) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Status devolve o status enviado, ou 200 quando o handler não chamou
// WriteHeader.
func (r *respostaComStatus) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

// ErrConfigInvalida é devolvido quando o ambiente traz um valor que a padaria
// não consegue usar.
//
// Configuração errada precisa derrubar o programa NA PARTIDA, com uma
// mensagem clara. O contrário é um serviço que sobe, parece saudável, e falha
// três horas depois na primeira requisição que toca no valor errado.
var ErrConfigInvalida = errors.New("configuração inválida")

// ----------------------------------------------------------------------
// Módulo 14: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Ambiente lê uma variável de ambiente. A assinatura é a mesma de
// os.LookupEnv, de propósito.
//
// Receber isto como parâmetro, em vez de chamar os.Getenv direto, é a mesma
// ideia do relógio do módulo 06: o teste entrega um mapa e decide o que o
// programa enxerga, sem mexer no ambiente da máquina, e sem que dois testes
// rodando juntos briguem por uma variável global.
type Ambiente func(chave string) (string, bool)

// AmbienteDeMapa devolve um Ambiente que lê de um mapa. Serve para teste.
func AmbienteDeMapa(valores map[string]string) Ambiente {
	return func(chave string) (string, bool) {
		valor, existe := valores[chave]
		return valor, existe
	}
}

// Config é tudo que a padaria precisa saber para subir.
type Config struct {
	Endereco            string
	NivelDeLog          slog.Level
	FormatoDeLog        string
	Trabalhadores       int
	PrazoDeDesligamento time.Duration
}

// CarregarConfig monta a Config a partir do ambiente, com padrões.
//
//	PADARIA_ENDERECO             padrão ":8080"
//	PADARIA_LOG_NIVEL            padrão "info"; aceita debug, info, warn, error
//	PADARIA_LOG_FORMATO          padrão "texto"; aceita texto, json
//	PADARIA_TRABALHADORES        padrão 3; precisa ser maior que zero
//	PADARIA_PRAZO_DESLIGAMENTO   padrão "15s"; qualquer coisa que o
//	                             time.ParseDuration entenda
//
// Qualquer valor que não sirva devolve um erro envolvendo ErrConfigInvalida,
// citando o nome da variável e o valor recebido. A mensagem é lida por alguém
// às três da manhã: diga qual variável, o que veio, e o que era esperado.
//
// Use strconv.Atoi para o número e time.ParseDuration para o prazo.
func CarregarConfig(obter Ambiente) (Config, error) {
	// TODO: implemente esta função.
	return Config{}, nil
}

// NovoLogger monta o logger conforme a configuração.
//
// slog é o log estruturado da biblioteca padrão, desde o Go 1.21. Ele escreve
// pares chave-valor em vez de frases, e é isso que permite a uma ferramenta
// de busca filtrar por status ou por rota sem ler texto com expressão regular.
//
// Use slog.NewJSONHandler quando o formato for "json" e slog.NewTextHandler
// nos outros casos, sempre com o nível vindo da Config:
//
//	slog.New(slog.NewJSONHandler(saida, &slog.HandlerOptions{Level: cfg.NivelDeLog}))
func NovoLogger(saida io.Writer, cfg Config) *slog.Logger {
	// TODO: implemente esta função.
	return nil
}

// ComLog embrulha um handler e registra uma linha estruturada por requisição.
//
// É o middleware do módulo 10, agora com slog no lugar da string montada à
// mão. A mensagem é "requisição", e os campos são exatamente estes:
//
//	metodo      o método HTTP
//	caminho     r.URL.Path
//	status      o status realmente enviado, via respostaComStatus
//	duracao_ms  quanto durou, em milissegundos
//
// Prefira log.LogAttrs com slog.String, slog.Int e slog.Int64 à forma
// log.Info("msg", "chave", valor). A primeira é mais verbosa e não aloca; a
// segunda é conveniente e aloca. Num caminho que roda em toda requisição, a
// diferença aparece.
func ComLog(proximo http.Handler, log *slog.Logger) http.Handler {
	// TODO: implemente esta função.
	return proximo
}

// Versao devolve a versão do binário, para aparecer no log da partida e num
// endereço de diagnóstico.
//
// Não é preciso injetar nada com -ldflags: desde o Go 1.18 o compilador grava
// os dados do controle de versão dentro do binário, e runtime/debug.ReadBuildInfo
// lê de volta.
//
// O que fazer:
//
//  1. debug.ReadBuildInfo(); se o segundo retorno for false, devolva
//     "desconhecida";
//  2. se info.Main.Version tiver algo diferente de "" e de "(devel)", devolva
//     ele: quando o binário vem de uma versão publicada, ou de um repositório
//     com histórico, essa string já é a resposta completa;
//  3. só quando não houver versão, procure em info.Settings a chave
//     "vcs.revision", corte para sete caracteres e devolva
//     "desenvolvimento+a1b2c3d".
func Versao() string {
	// TODO: implemente esta função.
	return ""
}
