// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria para de depender de onde os produtos estão
// guardados, e o erro de estoque passa a carregar dados em vez de só uma
// mensagem.
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

// Erros sentinela da padaria.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")
	ErrProdutoInvalido      = errors.New("produto inválido")
)

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
func NovoEstoque() Estoque {
	return make(Estoque)
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func (e Estoque) Repor(nome string, quantidade int) {
	e[nome] += quantidade
}

// Quantidade devolve quantas unidades existem do produto no estoque.
func (e Estoque) Quantidade(nome string) int {
	return e[nome]
}

// ----------------------------------------------------------------------
// A interface, e quem depende dela.
// ----------------------------------------------------------------------

// BuscadorDeProdutos é tudo que a função Vender precisa saber sobre onde os
// produtos moram.
//
// Repare em três coisas.
//
// A interface tem um método só. Quanto menor a interface, mais coisas a
// satisfazem, e mais fácil é substituir uma por outra.
//
// A interface está declarada aqui, perto de quem a CONSOME, e não junto de
// quem a implementa. Em Java ou C# você declara a interface ao lado da
// classe e escreve implements. Em Go é o contrário: quem precisa de um
// comportamento descreve o comportamento que precisa, e qualquer tipo que já
// tenha aquele método serve, sem saber que esta interface existe.
//
// Nenhum tipo neste arquivo diz que implementa BuscadorDeProdutos. Não existe
// essa palavra em Go.
type BuscadorDeProdutos interface {
	Buscar(nome string) (Produto, error)
}

// Esta linha não guarda nada e não roda nada. Ela existe só para o
// compilador conferir, agora, que *RepositorioEmMemoria satisfaz
// BuscadorDeProdutos. Sem ela, o erro só apareceria lá na frente, no ponto em
// que alguém tentasse usar um pelo outro.
var _ BuscadorDeProdutos = (*RepositorioEmMemoria)(nil)

// ----------------------------------------------------------------------
// Exemplo resolvido: Baixar devolvendo um erro que carrega dados.
// ----------------------------------------------------------------------

// ErroDeEstoque é o erro prometido lá no fim do módulo 03: um erro que, além
// da mensagem, carrega o que quem chamou precisa para reagir.
//
// Com ele a padaria consegue responder "não tenho 99, mas tenho 10" em vez de
// só recusar a venda.
type ErroDeEstoque struct {
	Produto    string
	Pedido     int
	Disponivel int
}

// Baixar tira do estoque a quantidade vendida de um produto.
//
// Já está pronto. O caso de estoque insuficiente devolve um *ErroDeEstoque em
// vez do sentinela pelado. Para isso funcionar, faltam os dois métodos que
// você vai escrever mais abaixo.
func (e Estoque) Baixar(nome string, quantidade int) error {
	if quantidade <= 0 {
		return fmt.Errorf("baixar %d de %q: %w", quantidade, nome, ErrQuantidadeInvalida)
	}

	quantidadeAtual, existe := e[nome]
	if !existe {
		return fmt.Errorf("baixar %q: %w", nome, ErrProdutoNaoEncontrado)
	}

	if quantidadeAtual < quantidade {
		return &ErroDeEstoque{
			Produto:    nome,
			Pedido:     quantidade,
			Disponivel: quantidadeAtual,
		}
	}

	e[nome] = quantidadeAtual - quantidade

	return nil
}

// ----------------------------------------------------------------------
// Módulo 05: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Error faz de *ErroDeEstoque um error.
//
// O tipo error da biblioteca padrão é, e sempre foi, uma interface de um
// método só:
//
//	type error interface {
//		Error() string
//	}
//
// Ou seja: você vem devolvendo interface desde o módulo 03 sem saber. Escreva
// este método e o seu tipo vira um erro legítimo, sem declarar nada.
//
// A mensagem precisa citar o nome do produto, a quantidade pedida e a
// disponível. Algo como:
//
//	estoque insuficiente de "Pão de queijo": pediram 99, tem 10
func (e *ErroDeEstoque) Error() string {
	// TODO: implemente este método.
	return ""
}

// Unwrap liga este erro ao sentinela ErrEstoqueInsuficiente.
//
// errors.Is chama Unwrap para descer uma camada e continuar procurando. Com
// este método, quem só quer saber "foi falta de estoque?" continua usando
// errors.Is exatamente como no módulo 03, sem saber que existe um tipo novo.
// Quem quiser os números usa errors.As.
func (e *ErroDeEstoque) Unwrap() error {
	// TODO: implemente este método.
	return nil
}

// RepositorioEmMemoria guarda produtos num map, indexados pelo nome.
//
// O campo é minúsculo de propósito: ninguém de fora do pacote mexe no map
// direto, só através dos métodos.
type RepositorioEmMemoria struct {
	produtos map[string]Produto
}

// NovoRepositorioEmMemoria devolve um repositório pronto para uso.
//
// Lembre do módulo 02: um map precisa ser inicializado antes da primeira
// escrita, e o valor zero de RepositorioEmMemoria tem um map nil dentro.
func NovoRepositorioEmMemoria() *RepositorioEmMemoria {
	// TODO: implemente esta função.
	return nil
}

// Salvar guarda o produto, sobrescrevendo se já existir um com o mesmo nome.
//
// Produto sem nome é recusado com um erro que envolve ErrProdutoInvalido.
//
// O método devolve error mesmo não tendo como falhar por outro motivo. A
// assinatura é desenhada para o caso geral: um repositório em banco falharia
// aqui, e quem usa a interface não deveria precisar saber a diferença.
func (r *RepositorioEmMemoria) Salvar(produto Produto) error {
	// TODO: implemente este método.
	return nil
}

// Buscar procura um produto pelo nome exato.
//
// Não achou devolve o valor zero e um erro que envolve
// ErrProdutoNaoEncontrado, citando o nome procurado.
func (r *RepositorioEmMemoria) Buscar(nome string) (Produto, error) {
	// TODO: implemente este método.
	return Produto{}, nil
}

// Todos devolve os produtos guardados, em ordem alfabética de nome.
//
// Ordem alfabética não é enfeite: percorrer map dá ordem aleatória, como você
// viu no módulo 02.
func (r *RepositorioEmMemoria) Todos() []Produto {
	// TODO: implemente este método.
	return nil
}

// Vender acha o produto, dá baixa no estoque e devolve o total a pagar.
//
// Compare a assinatura com a do módulo 03. Onde antes havia []Produto, agora
// há BuscadorDeProdutos. Vender parou de saber onde os produtos moram: pode
// ser um map, um arquivo, um banco de dados ou um dublê de teste, e o código
// aqui dentro não muda.
//
// Envolva o erro recebido com o contexto da venda, como no módulo 03.
func Vender(buscador BuscadorDeProdutos, estoque Estoque, nome string, quantidade int) (int, error) {
	// TODO: implemente esta função.
	return 0, nil
}
