// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 14. Olhe depois de tentar, não antes.
package padaria

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
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
// Módulo 14.
// ----------------------------------------------------------------------

// Ambiente lê uma variável de ambiente, com a assinatura de os.LookupEnv.
type Ambiente func(chave string) (string, bool)

// AmbienteDeMapa devolve um Ambiente que lê de um mapa.
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

// comPadrao devolve o valor do ambiente, ou o padrão quando ele não existe ou
// está vazio.
//
// Tratar string vazia como ausente é decisão consciente: num arquivo de
// ambiente, PADARIA_ENDERECO= é quase sempre esquecimento, não intenção de
// usar endereço vazio.
func comPadrao(obter Ambiente, chave, padrao string) string {
	if valor, existe := obter(chave); existe && valor != "" {
		return valor
	}
	return padrao
}

// CarregarConfig monta a Config a partir do ambiente, com padrões.
func CarregarConfig(obter Ambiente) (Config, error) {
	cfg := Config{
		Endereco:     comPadrao(obter, "PADARIA_ENDERECO", ":8080"),
		FormatoDeLog: comPadrao(obter, "PADARIA_LOG_FORMATO", "texto"),
	}

	nivelTexto := comPadrao(obter, "PADARIA_LOG_NIVEL", "info")
	switch strings.ToLower(nivelTexto) {
	case "debug":
		cfg.NivelDeLog = slog.LevelDebug
	case "info":
		cfg.NivelDeLog = slog.LevelInfo
	case "warn":
		cfg.NivelDeLog = slog.LevelWarn
	case "error":
		cfg.NivelDeLog = slog.LevelError
	default:
		// A mensagem diz a variável, o valor recebido e o que era esperado.
		// Quem lê isto está com o serviço fora do ar.
		return Config{}, fmt.Errorf(
			"PADARIA_LOG_NIVEL=%q, esperado debug, info, warn ou error: %w",
			nivelTexto, ErrConfigInvalida)
	}

	if formato := strings.ToLower(cfg.FormatoDeLog); formato != "texto" && formato != "json" {
		return Config{}, fmt.Errorf(
			"PADARIA_LOG_FORMATO=%q, esperado texto ou json: %w",
			cfg.FormatoDeLog, ErrConfigInvalida)
	}

	trabalhadoresTexto := comPadrao(obter, "PADARIA_TRABALHADORES", "3")
	trabalhadores, err := strconv.Atoi(trabalhadoresTexto)
	if err != nil {
		return Config{}, fmt.Errorf(
			"PADARIA_TRABALHADORES=%q não é um número: %w",
			trabalhadoresTexto, ErrConfigInvalida)
	}
	if trabalhadores <= 0 {
		return Config{}, fmt.Errorf(
			"PADARIA_TRABALHADORES=%d, esperado maior que zero: %w",
			trabalhadores, ErrConfigInvalida)
	}
	cfg.Trabalhadores = trabalhadores

	prazoTexto := comPadrao(obter, "PADARIA_PRAZO_DESLIGAMENTO", "15s")
	prazo, err := time.ParseDuration(prazoTexto)
	if err != nil {
		return Config{}, fmt.Errorf(
			"PADARIA_PRAZO_DESLIGAMENTO=%q não é uma duração, tente algo como 15s: %w",
			prazoTexto, ErrConfigInvalida)
	}
	if prazo <= 0 {
		return Config{}, fmt.Errorf(
			"PADARIA_PRAZO_DESLIGAMENTO=%v, esperado maior que zero: %w",
			prazo, ErrConfigInvalida)
	}
	cfg.PrazoDeDesligamento = prazo

	return cfg, nil
}

// NovoLogger monta o logger conforme a configuração.
func NovoLogger(saida io.Writer, cfg Config) *slog.Logger {
	opcoes := &slog.HandlerOptions{Level: cfg.NivelDeLog}

	if strings.EqualFold(cfg.FormatoDeLog, "json") {
		return slog.New(slog.NewJSONHandler(saida, opcoes))
	}

	return slog.New(slog.NewTextHandler(saida, opcoes))
}

// ComLog embrulha um handler e registra uma linha estruturada por requisição.
func ComLog(proximo http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		resposta := &respostaComStatus{ResponseWriter: w}

		proximo.ServeHTTP(resposta, r)

		// LogAttrs é a forma sem alocação. A forma curta, log.Info com pares
		// soltos, precisa embrulhar cada valor num any. Aqui, que roda em toda
		// requisição, a diferença vale a verbosidade.
		log.LogAttrs(r.Context(), slog.LevelInfo, "requisição",
			slog.String("metodo", r.Method),
			slog.String("caminho", r.URL.Path),
			slog.Int("status", resposta.Status()),
			slog.Int64("duracao_ms", time.Since(inicio).Milliseconds()),
		)
	})
}

// Versao devolve a versão do binário.
//
// Não é preciso injetar nada com -ldflags: desde o Go 1.18 o compilador grava
// os dados do controle de versão dentro do binário.
func Versao() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "desconhecida"
	}

	// Num binário publicado isto é a etiqueta da versão. Num compilado a
	// partir do repositório, o Go monta uma pseudoversão que já traz data e
	// revisão, e um sufixo +dirty quando havia alteração não commitada. Nos
	// dois casos, essa string sozinha já responde a pergunta.
	if versao := info.Main.Version; versao != "" && versao != "(devel)" {
		return versao
	}

	for _, ajuste := range info.Settings {
		if ajuste.Key == "vcs.revision" && ajuste.Value != "" {
			revisao := ajuste.Value
			if len(revisao) > 7 {
				revisao = revisao[:7]
			}
			return "desenvolvimento+" + revisao
		}
	}

	return "desenvolvimento"
}
