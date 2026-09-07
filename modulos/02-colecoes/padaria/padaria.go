// Package padaria guarda o código da Padaria do Seu Zé.
//
// Cada módulo é independente: o que você implementou antes chega pronto aqui.
package padaria

import "fmt"

// ----------------------------------------------------------------------
// Vindo dos módulos 00 e 01, já implementado.
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
// Módulo 02: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Estoque diz quantas unidades a padaria tem de cada produto, indexado pelo
// nome do produto.
//
// Isto é um tipo novo com nome próprio, não um apelido. Dar nome a um
// map[string]int faz o código falar português em vez de falar estrutura de
// dados.
type Estoque map[string]int

// Disponiveis devolve só os produtos que estão à venda, na mesma ordem em
// que apareceram no cardápio.
//
// Cardápio vazio ou nil devolve uma fatia vazia. Você não precisa tratar o
// nil separadamente: percorrer uma fatia nil simplesmente não entra no laço.
func Disponiveis(cardapio []Produto) []Produto {
	// TODO: implemente esta função.
	return nil
}

// MaisBarato devolve o produto mais barato do cardápio.
//
// O segundo retorno diz se havia algum produto. Cardápio vazio devolve o
// valor zero de Produto e false. Este par valor-e-encontrou é o mesmo
// formato que você já viu em map, e é o degrau antes dos erros do módulo 03.
//
// Empate fica com o primeiro que apareceu.
func MaisBarato(cardapio []Produto) (Produto, bool) {
	// TODO: implemente esta função.
	return Produto{}, false
}

// Reajustar aumenta o preço de todo o cardápio pelo percentual pedido,
// alterando os produtos no lugar. A função não devolve nada.
//
// Use exatamente esta conta:
//
//	preçoNovo = preço + (preço * percentual / 100)
//
// Preste atenção no que esta função faz com o cardápio de quem chamou. O
// módulo 01 te ensinou que struct é cópia. Aqui a regra é outra, e o teste
// TestReajustarAlteraOCardapioDeQuemChamou está lá para provar.
func Reajustar(cardapio []Produto, percentual int) {
	// TODO: implemente esta função.
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
//
// Não devolva o valor zero de map. O valor zero de map é nil, e escrever num
// map nil derruba o programa em tempo de execução. O README explica.
func NovoEstoque() Estoque {
	// TODO: implemente esta função.
	return nil
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
//
// Produto que ainda não estava no estoque começa do zero.
func Repor(estoque Estoque, nome string, quantidade int) {
	// TODO: implemente esta função.
}

// QuantidadeEm devolve quantas unidades existem do produto.
//
// Produto que não está no estoque tem quantidade zero. Estoque nil também
// devolve zero, sem explodir: ler de um map nil é permitido.
func QuantidadeEm(estoque Estoque, nome string) int {
	// TODO: implemente esta função.
	return 0
}

// NomesEmOrdem devolve os nomes dos produtos do estoque em ordem alfabética.
//
// Ordenar não é enfeite. Percorrer um map em Go devolve as chaves em ordem
// aleatória, de propósito, e ela muda a cada execução. Sem ordenar, esta
// função seria impossível de testar.
func NomesEmOrdem(estoque Estoque) []string {
	// TODO: implemente esta função.
	return nil
}
