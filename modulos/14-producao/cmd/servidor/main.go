// Command servidor é a padaria pronta para produção.
//
// Ele junta o que os módulos anteriores construíram: configuração vinda do
// ambiente, log estruturado, prazos no servidor HTTP e desligamento com graça.
//
//	go run ./modulos/14-producao/cmd/servidor
//	PADARIA_LOG_FORMATO=json PADARIA_ENDERECO=:9000 go run ./modulos/14-producao/cmd/servidor
//
// E, em outro terminal:
//
//	curl -s localhost:8080/saude
//	curl -s localhost:8080/versao
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RootControl/trilha-go/modulos/14-producao/padaria"
)

func main() {
	// main faz três coisas e mais nada: liga o mundo de fora, chama executar,
	// e traduz erro em código de saída. Toda a lógica que pode falhar mora numa
	// função que devolve error, e é isso que permite os defers rodarem antes da
	// saída do processo.
	if err := executar(); err != nil {
		fmt.Fprintln(os.Stderr, "padaria:", err)
		os.Exit(1)
	}
}

func executar() error {
	// Configuração primeiro, antes de qualquer outra coisa. Ambiente errado
	// derruba o programa aqui, na partida, com uma mensagem que diz qual
	// variável está errada.
	cfg, err := padaria.CarregarConfig(os.LookupEnv)
	if err != nil {
		return err
	}

	log := padaria.NovoLogger(os.Stderr, cfg)

	// O log vai para o erro padrão, e não para a saída padrão. Assim a saída
	// continua livre para dados, como no módulo 08, e um contêiner que coleta
	// os dois canais separadamente não mistura as coisas.
	log.Info("padaria abrindo",
		"versao", padaria.Versao(),
		"endereco", cfg.Endereco,
		"trabalhadores", cfg.Trabalhadores,
		"nivel_de_log", cfg.NivelDeLog.String(),
	)

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	mux := http.NewServeMux()

	// Endereço de saúde: é o que um orquestrador consulta para saber se pode
	// mandar tráfego. Precisa ser barato e não depender de banco nem de rede.
	mux.HandleFunc("GET /saude", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /versao", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"versao": padaria.Versao()})
	})

	servidor := &http.Server{
		Addr:              cfg.Endereco,
		Handler:           padaria.ComLog(mux, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	erroDoServidor := make(chan error, 1)
	go func() {
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			erroDoServidor <- err
		}
		close(erroDoServidor)
	}()

	select {
	case err := <-erroDoServidor:
		if err != nil {
			return fmt.Errorf("servidor: %w", err)
		}
		return nil

	case <-ctx.Done():
		log.Info("sinal recebido, fechando a padaria")
	}

	ctxDesligamento, cancelar := context.WithTimeout(context.Background(), cfg.PrazoDeDesligamento)
	defer cancelar()

	if err := servidor.Shutdown(ctxDesligamento); err != nil {
		return fmt.Errorf("desligamento forçado: %w", err)
	}

	log.Info("padaria fechada")

	return nil
}
