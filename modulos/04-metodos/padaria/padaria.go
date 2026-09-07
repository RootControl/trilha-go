// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria não ganha comportamento novo. Ela ganha uma forma
// nova: o que era função solta vira método pendurado no tipo.
package padaria

import (
	"errors"
	"fmt"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// Caixa acumula o que a padaria vendeu no dia.
type Caixa struct {
	TotalEmCentavos int
	Vendas          int
}

// Erros sentinela da padaria, do módulo 03.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")
)

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
func NovoEstoque() Estoque {
	return make(Estoque)
}

// ----------------------------------------------------------------------
// Exemplo resolvido: a função Baixar do módulo 03, virada método.
// ----------------------------------------------------------------------

// Baixar tira do estoque a quantidade vendida de um produto.
//
// Compare com a versão do módulo 03. O corpo é idêntico. O que mudou:
//
//	antes:  func Baixar(estoque Estoque, nome string, quantidade int) error
//	agora:  func (e Estoque) Baixar(nome string, quantidade int) error
//
// O primeiro parâmetro saiu da lista e virou o receptor, entre a palavra func
// e o nome do método. É literalmente isso que um método é: uma função cujo
// primeiro parâmetro ganhou um lugar de destaque. A chamada muda de
// Baixar(estoque, "Sonho", 2) para estoque.Baixar("Sonho", 2).
//
// Repare que o receptor é Estoque, não *Estoque, e mesmo assim o método
// consegue escrever no estoque. O README explica por que isso funciona aqui e
// não funcionaria numa struct.
func (e Estoque) Baixar(nome string, quantidade int) error {
	if quantidade <= 0 {
		return fmt.Errorf("baixar %d de %q: %w", quantidade, nome, ErrQuantidadeInvalida)
	}

	quantidadeAtual, existe := e[nome]
	if !existe {
		return fmt.Errorf("baixar %q: %w", nome, ErrProdutoNaoEncontrado)
	}

	if quantidadeAtual < quantidade {
		return fmt.Errorf("baixar %d de %q, só tem %d: %w",
			quantidade, nome, quantidadeAtual, ErrEstoqueInsuficiente)
	}

	e[nome] = quantidadeAtual - quantidade

	return nil
}

// ----------------------------------------------------------------------
// Módulo 04: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Descrever escreve o produto do jeito que aparece na placa da padaria.
//
//	Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}
//	  ->  "Pão de queijo: R$ 4,50"
//	esgotado                       ->  "Pão de queijo: esgotado"
//	sem nome                       ->  "produto sem nome: ..."
//
// Este método só lê. Escolha o receptor de acordo.
func (p Produto) Descrever() string {
	// TODO: implemente este método.
	return ""
}

// Reajustar aumenta o preço deste produto pelo percentual pedido:
//
//	preçoNovo = preço + (preço * percentual / 100)
//
// Este método precisa alterar o produto de quem chamou. A assinatura abaixo
// está incompleta de propósito: decida qual receptor ela precisa ter.
func (p Produto) Reajustar(percentual int) {
	// TODO: implemente este método, e ajuste o receptor.
}

// Esgotar marca o produto como indisponível.
//
// Mesma pergunta do Reajustar sobre o receptor.
func (p Produto) Esgotar() {
	// TODO: implemente este método, e ajuste o receptor.
}

// Quantidade devolve quantas unidades existem do produto no estoque.
//
// Produto ausente, e estoque nil, devolvem zero sem explodir.
func (e Estoque) Quantidade(nome string) int {
	// TODO: implemente este método.
	return 0
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func (e Estoque) Repor(nome string, quantidade int) {
	// TODO: implemente este método.
}

// Registrar anota uma venda no caixa: soma o valor ao total e conta mais uma
// venda.
//
// Decida o receptor.
func (c Caixa) Registrar(valorEmCentavos int) {
	// TODO: implemente este método, e ajuste o receptor.
}

// TicketMedio devolve quanto valeu a venda média do dia, em centavos,
// usando divisão inteira.
//
// Caixa sem nenhuma venda devolve zero, e não pode dividir por zero.
func (c Caixa) TicketMedio() int {
	// TODO: implemente este método.
	return 0
}

// String descreve o caixa em uma linha:
//
//	"caixa: 3 vendas, R$ 27,00"
//
// Um método chamado String, sem parâmetros e devolvendo string, é uma
// convenção que o pacote fmt reconhece sozinho. Depois de implementar,
// fmt.Printf("%v", caixa) passa a imprimir isto em vez dos campos crus.
func (c Caixa) String() string {
	// TODO: implemente este método.
	return ""
}
