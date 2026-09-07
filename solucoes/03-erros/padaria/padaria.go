// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 03. Olhe depois de tentar, não antes.
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
// Módulo 03.
// ----------------------------------------------------------------------

// Erros sentinela da padaria.
//
// Cada errors.New devolve um valor novo e único. É por isso que dois
// sentinelas com a mesma mensagem ainda seriam diferentes um do outro, e é
// por isso que comparar erro por texto da mensagem é sempre errado.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")
)

// BuscarProduto procura um produto no cardápio pelo nome exato.
func BuscarProduto(cardapio []Produto, nome string) (Produto, error) {
	for _, produto := range cardapio {
		if produto.Nome == nome {
			return produto, nil
		}
	}

	// O valor zero de Produto junto com o erro. Quem chamou não deve olhar
	// para o primeiro retorno quando o segundo não é nil, mas devolver o
	// valor zero garante que, se olhar, não encontra lixo.
	return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
}

// Baixar tira do estoque a quantidade vendida de um produto.
func Baixar(estoque Estoque, nome string, quantidade int) error {
	if quantidade <= 0 {
		return fmt.Errorf("baixar %d de %q: %w", quantidade, nome, ErrQuantidadeInvalida)
	}

	// O par valor-e-existe do módulo 02 é o que separa "não tem no estoque"
	// de "tem no estoque e está zerado". Sem ele, os dois casos leriam 0 e
	// receberiam o mesmo erro errado.
	quantidadeAtual, existe := estoque[nome]
	if !existe {
		return fmt.Errorf("baixar %q: %w", nome, ErrProdutoNaoEncontrado)
	}

	if quantidadeAtual < quantidade {
		return fmt.Errorf("baixar %d de %q, só tem %d: %w",
			quantidade, nome, quantidadeAtual, ErrEstoqueInsuficiente)
	}

	// A escrita acontece depois de todas as verificações. Enquanto qualquer
	// coisa pode dar errado, o estoque continua intocado.
	estoque[nome] = quantidadeAtual - quantidade

	return nil
}

// Vender acha o produto, dá baixa no estoque e devolve o total a pagar.
func Vender(cardapio []Produto, estoque Estoque, nome string, quantidade int) (int, error) {
	produto, err := BuscarProduto(cardapio, nome)
	if err != nil {
		return 0, fmt.Errorf("vender %d de %q: %w", quantidade, nome, err)
	}

	// Buscar antes de baixar não é detalhe: se o produto nem está no
	// cardápio, o estoque não chega a ser tocado.
	if err := Baixar(estoque, nome, quantidade); err != nil {
		return 0, fmt.Errorf("vender %d de %q: %w", quantidade, nome, err)
	}

	return produto.PrecoEmCentavos * quantidade, nil
}
