// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 04. Olhe depois de tentar, não antes.
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

// Baixar tira do estoque a quantidade vendida de um produto.
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
// Módulo 04.
// ----------------------------------------------------------------------

// Descrever escreve o produto do jeito que aparece na placa da padaria.
//
// Receptor de valor, porque o método só lê. Quem chama sabe, só de olhar a
// assinatura, que este método não vai mexer no produto dele.
func (p Produto) Descrever() string {
	nome := p.Nome
	if nome == "" {
		nome = "produto sem nome"
	}

	if !p.Disponivel {
		return fmt.Sprintf("%s: esgotado", nome)
	}

	return fmt.Sprintf("%s: %s", nome, FormatarPreco(p.PrecoEmCentavos))
}

// Reajustar aumenta o preço deste produto.
//
// Receptor de ponteiro, porque o método escreve. Com *Produto, p é o endereço
// do produto de quem chamou, e p.PrecoEmCentavos alcança o original. Com
// Produto, p seria uma cópia e esta linha não faria absolutamente nada.
//
// Repare que você continua escrevendo p.PrecoEmCentavos, sem asterisco. Go
// desreferencia o ponteiro sozinho no acesso a campo.
func (p *Produto) Reajustar(percentual int) {
	p.PrecoEmCentavos += p.PrecoEmCentavos * percentual / 100
}

// Esgotar marca o produto como indisponível.
func (p *Produto) Esgotar() {
	p.Disponivel = false
}

// Quantidade devolve quantas unidades existem do produto no estoque.
//
// Funciona em estoque nil: ler de um map nil devolve o valor zero.
func (e Estoque) Quantidade(nome string) int {
	return e[nome]
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
//
// Receptor de valor, e mesmo assim o estoque de quem chamou muda. Não é
// exceção à regra: o que foi copiado aqui é a referência à tabela de hash, e
// a tabela é a mesma dos dois lados. Escrever numa struct copiada perde a
// escrita; escrever num map copiado não, porque o map em si não foi copiado.
func (e Estoque) Repor(nome string, quantidade int) {
	e[nome] += quantidade
}

// Registrar anota uma venda no caixa.
//
// Receptor de ponteiro: o método soma nos campos do caixa de quem chamou.
func (c *Caixa) Registrar(valorEmCentavos int) {
	c.TotalEmCentavos += valorEmCentavos
	c.Vendas++
}

// TicketMedio devolve quanto valeu a venda média do dia, em centavos.
func (c Caixa) TicketMedio() int {
	if c.Vendas == 0 {
		return 0
	}

	return c.TotalEmCentavos / c.Vendas
}

// String descreve o caixa em uma linha.
//
// Receptor de valor de propósito, mesmo o Caixa tendo métodos de ponteiro. Se
// String tivesse receptor *Caixa, fmt.Printf("%v", caixa) sobre um Caixa
// comum voltaria a imprimir os campos crus, porque um valor não carrega os
// métodos de ponteiro. Com receptor de valor, tanto Caixa quanto *Caixa
// imprimem bonito.
//
// Cuidado ao escrever este método: usar %v sobre o próprio c aqui dentro
// chamaria String de novo, para sempre, até a pilha estourar.
func (c Caixa) String() string {
	return fmt.Sprintf("caixa: %d vendas, %s", c.Vendas, FormatarPreco(c.TotalEmCentavos))
}
