// Package padaria guarda o código da Padaria do Seu Zé.
//
// Cada módulo é independente: o código que você já implementou nos módulos
// anteriores vem pronto aqui, para você poder rodar este módulo sozinho.
package padaria

import "fmt"

// Produto é uma coisa que a padaria vende.
//
// Repare que não existe construtor obrigatório, não existe herança e não
// existe null. Um Produto declarado sem nada dentro já é um Produto válido.
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// Pedido é um cliente levando uma quantidade de um produto.
//
// Por enquanto o pedido tem um item só. No módulo 02, quando você conhecer
// slices, ele passa a aceitar vários.
type Pedido struct {
	Cliente    string
	Produto    Produto
	Quantidade int
}

// FormatarPreco veio pronta do módulo 00.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoProduto monta um Produto disponível para venda.
//
// Go não tem construtor. O que existe é a convenção de escrever uma função
// NovoAlgumaCoisa que devolve o valor já montado do jeito certo.
//
// Um produto recém-criado está sempre disponível.
func NovoProduto(nome string, precoEmCentavos int) Produto {
	// TODO: implemente esta função.
	return Produto{}
}

// Descrever escreve o produto do jeito que aparece na placa da padaria:
//
//	Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}
//	  ->  "Pão de queijo: R$ 4,50"
//
//	Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: false}
//	  ->  "Pão de queijo: esgotado"
//
// Produto sem nome vira "produto sem nome". Isso vale inclusive para o valor
// zero de Produto, que é um Produto legítimo e precisa funcionar aqui.
func Descrever(p Produto) string {
	// TODO: implemente esta função.
	return ""
}

// AplicarDesconto devolve um Produto NOVO, com o preço reduzido pelo
// percentual pedido. O produto que você recebeu não pode mudar.
//
// Use exatamente esta conta, nesta ordem:
//
//	preçoNovo = preço - (preço * percentual / 100)
//
// A ordem importa porque divisão de inteiro em Go corta a parte quebrada.
// O README explica por que isso é decisão, não descuido.
func AplicarDesconto(p Produto, percentual int) Produto {
	// TODO: implemente esta função.
	return Produto{}
}

// TotalEmCentavos devolve quanto o pedido custa no total.
func TotalEmCentavos(pedido Pedido) int {
	// TODO: implemente esta função.
	return 0
}
