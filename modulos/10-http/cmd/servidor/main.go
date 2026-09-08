// Command servidor sobe a padaria na rede.
//
// Como no módulo 08, este arquivo já está pronto e não decide nada. Ele monta
// as peças, liga o servidor e traduz erro em código de saída.
//
// Rode assim, da raiz do repositório:
//
//	go run ./modulos/10-http/cmd/servidor
//
// E, em outro terminal:
//
//	curl -s localhost:8080/produtos
//	curl -s localhost:8080/produtos/P%C3%A3o%20de%20queijo
//	curl -s -X POST localhost:8080/vendas -d '{"cliente":"Maria","produto":"Broa","quantidade":2}'
//	curl -s -i localhost:8080/produtos/Croissant
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/RootControl/trilha-go/modulos/10-http/padaria"
)

// registradorSlog liga o Registrador da padaria ao log estruturado da
// biblioteca padrão. Quatro linhas, porque a interface tem um método só.
type registradorSlog struct{ log *slog.Logger }

func (r registradorSlog) Registrar(linha string) { r.log.Info(linha) }

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	servidor := padaria.NovoServidor(
		padaria.CardapioPadrao(),
		padaria.EstoquePadrao(),
		padaria.RelogioDoSistema{},
	)

	// http.ListenAndServe, a função de uma linha, sobe um servidor SEM
	// nenhum prazo. Uma conexão lenta ou maliciosa fica pendurada para
	// sempre, e um punhado delas derruba o processo. Em produção, sempre o
	// http.Server com prazos definidos.
	servidorHTTP := &http.Server{
		Addr:              ":8080",
		Handler:           padaria.ComRegistro(servidor, registradorSlog{log: log}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info("padaria aberta", "endereco", servidorHTTP.Addr)

	// ListenAndServe só devolve quando o servidor para, e devolve
	// http.ErrServerClosed num desligamento limpo. Desligamento limpo de
	// verdade é o módulo 13.
	if err := servidorHTTP.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "padaria:", err)
		os.Exit(1)
	}
}
