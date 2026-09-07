// Package padaria guarda o código da Padaria do Seu Zé.
//
// Cada módulo é independente: o que você implementou antes chega pronto aqui.
package padaria

import "fmt"

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores, já implementado.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
func NovoEstoque() Estoque {
	return make(Estoque)
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func Repor(estoque Estoque, nome string, quantidade int) {
	estoque[nome] += quantidade
}

// QuantidadeEm devolve quantas unidades existem do produto.
func QuantidadeEm(estoque Estoque, nome string) int {
	return estoque[nome]
}

// ----------------------------------------------------------------------
// Módulo 03: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Os três erros abaixo são erros sentinela: valores fixos, criados uma vez,
// que quem chama pode reconhecer. Repare na convenção do nome, que começa com
// Err, e na mensagem, que é minúscula e não termina em ponto. O motivo está
// no README.
//
// TODO: crie os três com errors.New.
var (
	ErrProdutoNaoEncontrado error
	ErrEstoqueInsuficiente  error
	ErrQuantidadeInvalida   error
)

// BuscarProduto procura um produto no cardápio pelo nome exato.
//
// Quando não acha, devolve o valor zero de Produto e um erro que envolve
// ErrProdutoNaoEncontrado, acrescentando o nome procurado. Use %w no
// fmt.Errorf para envolver, nunca %v: é o %w que deixa errors.Is enxergar o
// sentinela lá dentro.
//
//	fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
func BuscarProduto(cardapio []Produto, nome string) (Produto, error) {
	// TODO: implemente esta função.
	return Produto{}, nil
}

// Baixar tira do estoque a quantidade vendida de um produto.
//
// Esta é a função que você tentou escrever no desafio do módulo 02, agora do
// jeito idiomático. Verifique nesta ordem, e pare no primeiro problema:
//
//  1. quantidade menor ou igual a zero devolve ErrQuantidadeInvalida
//  2. produto que não existe no estoque devolve ErrProdutoNaoEncontrado
//  3. quantidade maior do que a disponível devolve ErrEstoqueInsuficiente
//
// Envolva o sentinela com o nome do produto, do mesmo jeito que BuscarProduto.
//
// Quando dá tudo certo, o estoque diminui e a função devolve nil. Quando dá
// errado, o estoque não pode ser tocado.
//
// Para distinguir "não existe no estoque" de "existe e está zerado", você
// precisa do par valor-e-existe que viu no módulo 02:
//
//	quantidadeAtual, existe := estoque[nome]
func Baixar(estoque Estoque, nome string, quantidade int) error {
	// TODO: implemente esta função.
	return nil
}

// Vender junta as duas anteriores: acha o produto no cardápio, dá baixa no
// estoque e devolve quanto o cliente tem a pagar, em centavos.
//
// Se qualquer etapa falhar, devolve 0 e um erro que envolve o erro recebido,
// acrescentando o contexto da venda:
//
//	fmt.Errorf("vender %d de %q: %w", quantidade, nome, err)
//
// Repare que aqui você envolve um erro que já vinha envolvido. Isso empilha
// contexto, e errors.Is continua achando o sentinela no fundo da pilha.
func Vender(cardapio []Produto, estoque Estoque, nome string, quantidade int) (int, error) {
	// TODO: implemente esta função.
	return 0, nil
}
