// Command servidor mostra um desligamento com graça.
//
// Este arquivo já está pronto, e vale mais rodá-lo do que lê-lo. Suba o
// servidor, faça uma requisição lenta e, enquanto ela está no ar, aperte
// Ctrl+C. O servidor para de aceitar conexões novas, espera a requisição em
// andamento terminar, e só então sai.
//
//	go run ./modulos/13-fila/cmd/servidor
//
// Em outro terminal:
//
//	curl "localhost:8080/assar?produto=Broa&quantidade=5"
//
// E, com a requisição no ar, Ctrl+C no primeiro terminal.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RootControl/trilha-go/modulos/13-fila/padaria"
)

const prazoDeDesligamento = 15 * time.Second

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// signal.NotifyContext devolve um contexto que é cancelado quando um
	// destes sinais chega. Ctrl+C manda SIGINT; um orquestrador de contêiner
	// manda SIGTERM antes de matar o processo. Tratar os dois é o mínimo.
	//
	// O parar adiado devolve o tratamento padrão dos sinais ao sistema, o que
	// importa se o programa continuasse rodando depois.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	forno := padaria.FornoLento{Demora: 5 * time.Second}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /assar", func(w http.ResponseWriter, r *http.Request) {
		quantidade, _ := strconv.Atoi(r.URL.Query().Get("quantidade"))
		pedido := padaria.PedidoDeFornada{
			Produto:    r.URL.Query().Get("produto"),
			Quantidade: quantidade,
		}

		// r.Context() é cancelado quando o CLIENTE desiste, por exemplo se
		// alguém fecha o curl com Ctrl+C. Repassar esse contexto é o que faz o
		// servidor parar de trabalhar por alguém que já foi embora.
		fornada, err := forno.Assar(r.Context(), pedido)
		if err != nil {
			log.Warn("fornada abortada", "erro", err, "produto", pedido.Produto)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(fornada)
	})

	servidor := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// O servidor sobe numa goroutine para a main poder ficar esperando o
	// sinal. O erro dele viaja de volta por um canal, porque uma goroutine não
	// tem como devolver valor.
	erroDoServidor := make(chan error, 1)
	go func() {
		log.Info("padaria aberta", "endereco", servidor.Addr)

		// ListenAndServe devolve http.ErrServerClosed num desligamento
		// pedido, e isso não é falha.
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			erroDoServidor <- err
		}
		close(erroDoServidor)
	}()

	select {
	case err := <-erroDoServidor:
		if err != nil {
			fmt.Fprintln(os.Stderr, "padaria:", err)
			os.Exit(1)
		}

	case <-ctx.Done():
		log.Info("sinal recebido, fechando a padaria sem jogar pão fora")
	}

	// Prazo próprio para o desligamento, e ele NÃO pode descender do ctx, que
	// já está cancelado. context.Background é a raiz certa aqui.
	ctxDesligamento, cancelar := context.WithTimeout(context.Background(), prazoDeDesligamento)
	defer cancelar()

	// Shutdown para de aceitar conexões novas e espera as requisições em
	// andamento terminarem. Se o prazo estourar antes, ele devolve o erro do
	// contexto e as conexões restantes são cortadas.
	//
	// A diferença para servidor.Close() é exatamente essa: Close corta tudo na
	// hora, no meio da resposta de quem estava sendo atendido.
	if err := servidor.Shutdown(ctxDesligamento); err != nil {
		log.Error("desligamento forçado", "erro", err)
		os.Exit(1)
	}

	log.Info("padaria fechada")
}
