// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 01. Olhe depois de tentar, não antes.
package padaria

import "fmt"

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// Pedido é um cliente levando uma quantidade de um produto.
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
// A função devolve um Produto, não um *Produto. Devolver o valor é o padrão
// em Go para coisas pequenas como esta. Ponteiro só quando houver motivo, e
// o motivo aparece no módulo 04.
func NovoProduto(nome string, precoEmCentavos int) Produto {
	return Produto{
		Nome:            nome,
		PrecoEmCentavos: precoEmCentavos,
		Disponivel:      true,
	}
}

// Descrever escreve o produto do jeito que aparece na placa da padaria.
func Descrever(p Produto) string {
	nome := p.Nome
	if nome == "" {
		nome = "produto sem nome"
	}

	if !p.Disponivel {
		return fmt.Sprintf("%s: esgotado", nome)
	}

	return fmt.Sprintf("%s: %s", nome, FormatarPreco(p.PrecoEmCentavos))
}

// AplicarDesconto devolve um Produto novo, com o preço reduzido.
//
// Repare que a função mexe direto em p e devolve p. Isso é seguro, e é o
// ponto do módulo: p é uma cópia do produto de quem chamou. Mudar a cópia
// não alcança o original. Em JavaScript ou Python o mesmo código teria
// alterado o objeto de quem chamou.
func AplicarDesconto(p Produto, percentual int) Produto {
	desconto := p.PrecoEmCentavos * percentual / 100
	p.PrecoEmCentavos = p.PrecoEmCentavos - desconto
	return p
}

// TotalEmCentavos devolve quanto o pedido custa no total.
//
// Não precisa tratar o pedido vazio: o valor zero de Pedido tem produto de
// preço 0 e quantidade 0, e 0 vezes 0 já é o resultado certo. Valor zero bem
// escolhido apaga um if.
func TotalEmCentavos(pedido Pedido) int {
	return pedido.Produto.PrecoEmCentavos * pedido.Quantidade
}
