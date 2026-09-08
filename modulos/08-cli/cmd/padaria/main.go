// Command padaria é o balcão da Padaria do Seu Zé, na linha de comando.
//
// Este arquivo já está pronto, e é de propósito que ele seja tão curto. Um
// pacote main bem escrito não decide nada: ele liga o mundo de fora ao seu
// código e traduz um erro em código de saída. Toda a lógica está na
// biblioteca, onde dá para testar.
//
// Rode assim, da raiz do repositório:
//
//	go run ./modulos/08-cli/cmd/padaria cardapio
//	go run ./modulos/08-cli/cmd/padaria -arquivo /tmp/pedidos.json vender "Pão de queijo" 3
//	go run ./modulos/08-cli/cmd/padaria -arquivo /tmp/pedidos.json pedidos
//	go run ./modulos/08-cli/cmd/padaria -h
package main

import (
	"fmt"
	"os"

	"github.com/RootControl/trilha-go/modulos/08-cli/padaria"
)

func main() {
	// os.Args[0] é o nome do programa, e a biblioteca de flags não quer ele.
	err := padaria.Executar(os.Args[1:], os.Stdout, os.Stderr, padaria.RelogioDoSistema{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "padaria:", err)

		// os.Exit encerra na hora e NÃO roda os defers pendentes. Por isso
		// ele fica aqui, sozinho, na última linha de main, e nunca no meio da
		// lógica. Código de saída diferente de zero é como um script em volta
		// descobre que deu errado.
		os.Exit(1)
	}
}
