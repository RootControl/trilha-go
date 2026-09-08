// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria para de repetir o mesmo laço para cada tipo. Repare
// que os genéricos só aparecem agora, no nono módulo: a ordem é essa de
// propósito. Você precisa ter sentido a repetição doer antes de removê-la.
package padaria

import (
	"cmp"
	"fmt"
	"time"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string `json:"nome"`
	PrecoEmCentavos int    `json:"preco_em_centavos"`
	Disponivel      bool   `json:"disponivel"`
}

// Pedido é um cliente levando uma quantidade de um produto, com a hora.
type Pedido struct {
	Cliente    string    `json:"cliente"`
	Produto    Produto   `json:"produto"`
	Quantidade int       `json:"quantidade"`
	Em         time.Time `json:"em"`
}

// Centavos é dinheiro da padaria, com nome próprio.
//
// O tipo subjacente é int, mas Centavos NÃO é int: o compilador trata os dois
// como tipos diferentes. Guarde isso, porque é a razão de existir o til nas
// restrições mais adiante.
type Centavos int

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// TotalEmCentavos devolve quanto o pedido custa.
func TotalEmCentavos(pedido Pedido) int {
	return pedido.Produto.PrecoEmCentavos * pedido.Quantidade
}

// ----------------------------------------------------------------------
// A repetição, para você ver antes de remover.
//
// As três funções abaixo já funcionam. Leia as três seguidas e repare que a
// única coisa diferente entre elas é o tipo e a condição. O corpo é o mesmo
// laço, copiado três vezes. Antes de 2022 não havia jeito melhor em Go.
// ----------------------------------------------------------------------

// ProdutosDisponiveis devolve só os produtos à venda.
func ProdutosDisponiveis(produtos []Produto) []Produto {
	var resultado []Produto
	for _, produto := range produtos {
		if produto.Disponivel {
			resultado = append(resultado, produto)
		}
	}
	return resultado
}

// PedidosDoCliente devolve só os pedidos de um cliente.
func PedidosDoCliente(pedidos []Pedido, cliente string) []Pedido {
	var resultado []Pedido
	for _, pedido := range pedidos {
		if pedido.Cliente == cliente {
			resultado = append(resultado, pedido)
		}
	}
	return resultado
}

// NomesDeProdutos devolve só os nomes.
func NomesDeProdutos(produtos []Produto) []string {
	var resultado []string
	for _, produto := range produtos {
		resultado = append(resultado, produto.Nome)
	}
	return resultado
}

// ----------------------------------------------------------------------
// Módulo 09: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Numero é uma restrição, não uma interface comum.
//
// Uma restrição descreve um CONJUNTO DE TIPOS em vez de um conjunto de
// métodos. A barra vertical é união, e o til quer dizer "qualquer tipo cujo
// tipo subjacente seja este".
//
// Ela já vem pronta porque restrição é conferida em tempo de compilação: uma
// restrição incompleta faria o arquivo de teste nem compilar, e você veria um
// erro de build em vez de um teste vermelho.
//
// Experimente mesmo assim, e desfaça depois: apague os três tis e rode o
// teste. Centavos deixa de ser aceito, porque Centavos não é int, apesar de
// ser feito de int. O til é o que transforma "este tipo" em "qualquer tipo
// construído sobre este".
type Numero interface {
	~int | ~int64 | ~float64
}

// Filtrar devolve os itens para os quais manter devolve true, na ordem
// original.
//
// Esta única função substitui ProdutosDisponiveis e PedidosDoCliente, e
// qualquer outro filtro que a padaria venha a precisar.
//
// O [T any] entre o nome e os parênteses é o parâmetro de tipo. any aqui não
// é o any de "qualquer valor" do módulo 05: como restrição, ele quer dizer
// "qualquer tipo serve", porque esta função não precisa fazer nada com o item
// além de repassá-lo.
func Filtrar[T any](itens []T, manter func(T) bool) []T {
	// TODO: implemente esta função.
	return nil
}

// Mapear transforma cada item, devolvendo uma fatia do tipo novo.
//
// Dois parâmetros de tipo: T é o que entra, R é o que sai. É esta função que
// substitui NomesDeProdutos.
func Mapear[T, R any](itens []T, transformar func(T) R) []R {
	// TODO: implemente esta função.
	return nil
}

// Somar devolve a soma dos valores. Fatia vazia soma zero.
//
// Repare que a restrição é Numero, e não any: para somar é preciso que o
// operador + exista para o tipo, e any não garante isso. Restrição é o que o
// compilador usa para saber o que ele pode fazer com um T.
func Somar[N Numero](valores []N) N {
	// TODO: implemente esta função.
	return 0
}

// MaiorPor devolve o item cujo valor calculado é o maior, e um bool dizendo
// se havia algum item.
//
// Dois parâmetros de tipo de novo, mas agora com restrições diferentes. T é
// qualquer coisa, e O é um tipo ordenável, ou seja, um tipo que aceita o
// operador maior-que. cmp.Ordered é a restrição da biblioteca padrão para
// isso, e cobre todos os inteiros, os floats e string.
//
// Empate fica com o primeiro. Fatia vazia devolve o valor zero de T e false.
// Escrever "o valor zero de T" quando você não sabe qual é T se faz assim:
//
//	var zero T
func MaiorPor[T any, O cmp.Ordered](itens []T, valor func(T) O) (T, bool) {
	// TODO: implemente esta função.
	var zero T
	return zero, false
}

// AgruparPor junta os itens em grupos, indexados pela chave calculada.
//
// A restrição da chave é comparable, e não any, porque chave de map precisa
// aceitar o operador de igualdade. comparable é uma restrição embutida na
// linguagem.
//
// Devolva sempre um map inicializado, mesmo para entrada vazia.
func AgruparPor[T any, K comparable](itens []T, chave func(T) K) map[K][]T {
	// TODO: implemente esta função.
	return nil
}

// Pilha guarda itens de qualquer tipo, e devolve na ordem inversa.
//
// Este é um TIPO genérico, e não uma função genérica. O parâmetro de tipo
// entra na declaração do tipo e reaparece no receptor dos métodos, escrito
// como Pilha[T].
//
// O valor zero precisa funcionar: var p Pilha[int] já está pronta para uso,
// porque append em fatia nil funciona. Não faça uma NovaPilha.
type Pilha[T any] struct {
	itens []T
}

// Empilhar põe um item no topo.
func (p *Pilha[T]) Empilhar(item T) {
	// TODO: implemente este método.
}

// Desempilhar tira o item do topo. O bool diz se havia algum.
func (p *Pilha[T]) Desempilhar() (T, bool) {
	// TODO: implemente este método.
	var zero T
	return zero, false
}

// Tamanho devolve quantos itens estão na pilha.
func (p *Pilha[T]) Tamanho() int {
	// TODO: implemente este método.
	return 0
}
