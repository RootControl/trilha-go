// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 02. Olhe depois de tentar, não antes.
package padaria

import (
	"fmt"
	"slices"
)

// ----------------------------------------------------------------------
// Vindo dos módulos 00 e 01.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoProduto monta um Produto disponível para venda.
func NovoProduto(nome string, precoEmCentavos int) Produto {
	return Produto{Nome: nome, PrecoEmCentavos: precoEmCentavos, Disponivel: true}
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

// ----------------------------------------------------------------------
// Módulo 02.
// ----------------------------------------------------------------------

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// Disponiveis devolve só os produtos que estão à venda.
//
// A fatia começa nil, e isso é de propósito. append funciona em fatia nil, e
// devolver nil quando não há nada é idiomático: para quem recebe, nil e fatia
// vazia se comportam igual em len, em range e em append.
func Disponiveis(cardapio []Produto) []Produto {
	var disponiveis []Produto

	for _, produto := range cardapio {
		if produto.Disponivel {
			disponiveis = append(disponiveis, produto)
		}
	}

	return disponiveis
}

// MaisBarato devolve o produto mais barato do cardápio.
func MaisBarato(cardapio []Produto) (Produto, bool) {
	if len(cardapio) == 0 {
		return Produto{}, false
	}

	maisBarato := cardapio[0]

	// A comparação é < e não <=, e é isso que faz o empate ficar com o
	// primeiro que apareceu.
	for _, produto := range cardapio[1:] {
		if produto.PrecoEmCentavos < maisBarato.PrecoEmCentavos {
			maisBarato = produto
		}
	}

	return maisBarato, true
}

// Reajustar aumenta o preço de todo o cardápio, no lugar.
//
// Repare no laço: ele percorre índices, não valores. Escrever
//
//	for _, produto := range cardapio { produto.PrecoEmCentavos += ... }
//
// compila, roda, não reclama de nada e não muda o cardápio. A variável
// produto é uma cópia do elemento, exatamente como o módulo 01 ensinou. Para
// alcançar o elemento de verdade você precisa do índice.
func Reajustar(cardapio []Produto, percentual int) {
	for i := range cardapio {
		cardapio[i].PrecoEmCentavos += cardapio[i].PrecoEmCentavos * percentual / 100
	}
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
//
// Estoque{} teria o mesmo efeito. make é a forma mais comum quando o map
// nasce vazio.
func NovoEstoque() Estoque {
	return make(Estoque)
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
//
// Não precisa checar se a chave existe. Ler uma chave ausente devolve o valor
// zero do tipo, que aqui é 0, então += já faz a coisa certa nos dois casos.
func Repor(estoque Estoque, nome string, quantidade int) {
	estoque[nome] += quantidade
}

// QuantidadeEm devolve quantas unidades existem do produto.
func QuantidadeEm(estoque Estoque, nome string) int {
	return estoque[nome]
}

// NomesEmOrdem devolve os nomes dos produtos do estoque em ordem alfabética.
//
// Estoque vazio devolve nil, porque nomes nunca chega a receber append.
func NomesEmOrdem(estoque Estoque) []string {
	var nomes []string

	for nome := range estoque {
		nomes = append(nomes, nome)
	}

	slices.Sort(nomes)

	return nomes
}
