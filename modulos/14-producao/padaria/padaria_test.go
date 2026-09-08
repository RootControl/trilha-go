package padaria

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// loggerDeTeste monta o logger e para o teste se ele vier nil, para uma
// implementação ainda por escrever produzir falha legível em vez de pânico.
func loggerDeTeste(t *testing.T, saida io.Writer, cfg Config) *slog.Logger {
	t.Helper()

	log := NovoLogger(saida, cfg)
	if log == nil {
		t.Fatal("NovoLogger devolveu nil")
	}

	return log
}

func TestCarregarConfigComOsPadroes(t *testing.T) {
	// Ambiente completamente vazio precisa produzir uma configuração
	// utilizável. Serviço que não sobe sem seis variáveis definidas é serviço
	// que ninguém consegue rodar na própria máquina.
	cfg, err := CarregarConfig(AmbienteDeMapa(nil))

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Endereco != ":8080" {
		t.Errorf("Endereco = %q, esperado :8080", cfg.Endereco)
	}
	if cfg.NivelDeLog != slog.LevelInfo {
		t.Errorf("NivelDeLog = %v, esperado info", cfg.NivelDeLog)
	}
	if cfg.FormatoDeLog != "texto" {
		t.Errorf("FormatoDeLog = %q, esperado texto", cfg.FormatoDeLog)
	}
	if cfg.Trabalhadores != 3 {
		t.Errorf("Trabalhadores = %d, esperado 3", cfg.Trabalhadores)
	}
	if cfg.PrazoDeDesligamento != 15*time.Second {
		t.Errorf("PrazoDeDesligamento = %v, esperado 15s", cfg.PrazoDeDesligamento)
	}
}

func TestCarregarConfigDoAmbiente(t *testing.T) {
	cfg, err := CarregarConfig(AmbienteDeMapa(map[string]string{
		"PADARIA_ENDERECO":           ":9000",
		"PADARIA_LOG_NIVEL":          "debug",
		"PADARIA_LOG_FORMATO":        "json",
		"PADARIA_TRABALHADORES":      "8",
		"PADARIA_PRAZO_DESLIGAMENTO": "30s",
	}))

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Endereco != ":9000" {
		t.Errorf("Endereco = %q, esperado :9000", cfg.Endereco)
	}
	if cfg.NivelDeLog != slog.LevelDebug {
		t.Errorf("NivelDeLog = %v, esperado debug", cfg.NivelDeLog)
	}
	if cfg.FormatoDeLog != "json" {
		t.Errorf("FormatoDeLog = %q, esperado json", cfg.FormatoDeLog)
	}
	if cfg.Trabalhadores != 8 {
		t.Errorf("Trabalhadores = %d, esperado 8", cfg.Trabalhadores)
	}
	if cfg.PrazoDeDesligamento != 30*time.Second {
		t.Errorf("PrazoDeDesligamento = %v, esperado 30s", cfg.PrazoDeDesligamento)
	}
}

func TestCarregarConfigTrataVazioComoAusente(t *testing.T) {
	// PADARIA_ENDERECO= num arquivo de ambiente é quase sempre esquecimento.
	cfg, err := CarregarConfig(AmbienteDeMapa(map[string]string{"PADARIA_ENDERECO": ""}))

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Endereco != ":8080" {
		t.Errorf("Endereco = %q, esperado o padrão :8080", cfg.Endereco)
	}
}

func TestCarregarConfigRecusa(t *testing.T) {
	casos := []struct {
		nome             string
		ambiente         map[string]string
		trechoNaMensagem string
	}{
		{
			nome:             "nível de log desconhecido",
			ambiente:         map[string]string{"PADARIA_LOG_NIVEL": "verboso"},
			trechoNaMensagem: "PADARIA_LOG_NIVEL",
		},
		{
			nome:             "formato de log desconhecido",
			ambiente:         map[string]string{"PADARIA_LOG_FORMATO": "xml"},
			trechoNaMensagem: "PADARIA_LOG_FORMATO",
		},
		{
			nome:             "trabalhadores que não é número",
			ambiente:         map[string]string{"PADARIA_TRABALHADORES": "muitos"},
			trechoNaMensagem: "PADARIA_TRABALHADORES",
		},
		{
			nome:             "trabalhadores igual a zero",
			ambiente:         map[string]string{"PADARIA_TRABALHADORES": "0"},
			trechoNaMensagem: "PADARIA_TRABALHADORES",
		},
		{
			nome:             "trabalhadores negativo",
			ambiente:         map[string]string{"PADARIA_TRABALHADORES": "-2"},
			trechoNaMensagem: "PADARIA_TRABALHADORES",
		},
		{
			nome:             "prazo que não é duração",
			ambiente:         map[string]string{"PADARIA_PRAZO_DESLIGAMENTO": "quinze"},
			trechoNaMensagem: "PADARIA_PRAZO_DESLIGAMENTO",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			_, err := CarregarConfig(AmbienteDeMapa(caso.ambiente))

			if err == nil {
				t.Fatal("esperava erro: configuração errada precisa derrubar o programa na partida")
			}
			if !errors.Is(err, ErrConfigInvalida) {
				t.Errorf("errors.Is não achou ErrConfigInvalida em %v", err)
			}
			// A mensagem é lida às três da manhã. Ela precisa dizer QUAL
			// variável está errada.
			if !strings.Contains(err.Error(), caso.trechoNaMensagem) {
				t.Errorf("a mensagem deveria citar %q, veio %q", caso.trechoNaMensagem, err.Error())
			}
		})
	}
}

func TestNovoLoggerEmJSON(t *testing.T) {
	var saida bytes.Buffer
	log := loggerDeTeste(t, &saida, Config{FormatoDeLog: "json", NivelDeLog: slog.LevelInfo})

	log.Info("padaria aberta", "endereco", ":8080")

	var linha map[string]any
	if err := json.Unmarshal(saida.Bytes(), &linha); err != nil {
		t.Fatalf("a saída não é JSON: %v\n%s", err, saida.String())
	}
	if linha["msg"] != "padaria aberta" {
		t.Errorf("msg = %v, esperado %q", linha["msg"], "padaria aberta")
	}
	if linha["endereco"] != ":8080" {
		t.Errorf("endereco = %v, esperado :8080", linha["endereco"])
	}
}

func TestNovoLoggerEmTexto(t *testing.T) {
	var saida bytes.Buffer
	log := loggerDeTeste(t, &saida, Config{FormatoDeLog: "texto", NivelDeLog: slog.LevelInfo})

	log.Info("padaria aberta", "endereco", ":8080")

	texto := saida.String()
	if json.Valid(saida.Bytes()) {
		t.Errorf("esperava formato texto, veio JSON: %s", texto)
	}
	if !strings.Contains(texto, "endereco=:8080") {
		t.Errorf("esperava o par chave=valor na saída, veio %q", texto)
	}
}

func TestNovoLoggerRespeitaONivel(t *testing.T) {
	var saida bytes.Buffer
	log := loggerDeTeste(t, &saida, Config{FormatoDeLog: "json", NivelDeLog: slog.LevelWarn})

	log.Debug("isto não deveria aparecer")
	log.Info("isto também não")
	log.Warn("esta sim")

	linhas := strings.Count(strings.TrimSpace(saida.String()), "\n") + 1
	if linhas != 1 {
		t.Errorf("saíram %d linhas, esperado 1:\n%s", linhas, saida.String())
	}
	if !strings.Contains(saida.String(), "esta sim") {
		t.Errorf("a linha de aviso não saiu: %s", saida.String())
	}
}

func TestComLog(t *testing.T) {
	var saida bytes.Buffer
	log := loggerDeTeste(t, &saida, Config{FormatoDeLog: "json", NivelDeLog: slog.LevelInfo})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	gravador := httptest.NewRecorder()
	ComLog(handler, log).ServeHTTP(gravador, httptest.NewRequest(http.MethodPost, "/vendas", nil))

	if gravador.Code != http.StatusCreated {
		t.Errorf("status = %d, esperado 201: o middleware não pode mudar a resposta", gravador.Code)
	}

	var linha map[string]any
	if err := json.Unmarshal(saida.Bytes(), &linha); err != nil {
		t.Fatalf("a saída não é JSON: %v\n%s", err, saida.String())
	}

	if linha["msg"] != "requisição" {
		t.Errorf("msg = %v, esperado %q", linha["msg"], "requisição")
	}
	if linha["metodo"] != "POST" {
		t.Errorf("metodo = %v, esperado POST", linha["metodo"])
	}
	if linha["caminho"] != "/vendas" {
		t.Errorf("caminho = %v, esperado /vendas", linha["caminho"])
	}
	// JSON não tem inteiro: todo número vira float64 ao decodificar em any.
	if status, ok := linha["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("status = %v, esperado 201", linha["status"])
	}
	if _, existe := linha["duracao_ms"]; !existe {
		t.Errorf("faltou o campo duracao_ms na linha: %s", saida.String())
	}
}

func TestComLogRegistraOStatusPadrao(t *testing.T) {
	// Handler que só escreve o corpo nunca chama WriteHeader, e o net/http
	// manda 200 sozinho. O middleware precisa registrar 200, não zero.
	var saida bytes.Buffer
	log := loggerDeTeste(t, &saida, Config{FormatoDeLog: "json", NivelDeLog: slog.LevelInfo})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	gravador := httptest.NewRecorder()
	ComLog(handler, log).ServeHTTP(gravador, httptest.NewRequest(http.MethodGet, "/saude", nil))

	var linha map[string]any
	if err := json.Unmarshal(saida.Bytes(), &linha); err != nil {
		t.Fatalf("a saída não é JSON: %v", err)
	}
	if status, ok := linha["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("status = %v, esperado 200", linha["status"])
	}
}

func TestVersao(t *testing.T) {
	versao := Versao()

	if versao == "" {
		t.Fatal("Versao() devolveu string vazia")
	}
	// Rodando por go test, o módulo principal é o pacote de teste e a versão
	// costuma vir como desenvolvimento. O que importa é nunca devolver vazio.
	t.Logf("versão do binário de teste: %q", versao)
}
