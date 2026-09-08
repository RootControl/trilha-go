// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 09. Olhe depois de tentar, não antes.
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
// Módulo 09.
// ----------------------------------------------------------------------

// Numero é uma restrição: qualquer tipo cujo tipo subjacente seja um destes.
type Numero interface {
	~int | ~int64 | ~float64
}

// Filtrar devolve os itens para os quais manter devolve true.
//
// O corpo é exatamente o mesmo das funções escritas à mão lá em cima. A
// diferença é que este só existe uma vez.
func Filtrar[T any](itens []T, manter func(T) bool) []T {
	var resultado []T

	for _, item := range itens {
		if manter(item) {
			resultado = append(resultado, item)
		}
	}

	return resultado
}

// Mapear transforma cada item, devolvendo uma fatia do tipo novo.
//
// Aqui vale reservar a capacidade: o tamanho final é conhecido desde o
// começo, ao contrário do Filtrar, onde não dá para saber quantos passam.
func Mapear[T, R any](itens []T, transformar func(T) R) []R {
	if len(itens) == 0 {
		return nil
	}

	resultado := make([]R, 0, len(itens))
	for _, item := range itens {
		resultado = append(resultado, transformar(item))
	}

	return resultado
}

// Somar devolve a soma dos valores.
//
// O literal 0 funciona como valor inicial de qualquer N porque a restrição
// Numero só admite tipos numéricos. Num T sem restrição isso não compilaria,
// e seria preciso escrever `var total N`.
func Somar[N Numero](valores []N) N {
	var total N

	for _, valor := range valores {
		total += valor
	}

	return total
}

// MaiorPor devolve o item cujo valor calculado é o maior.
func MaiorPor[T any, O cmp.Ordered](itens []T, valor func(T) O) (T, bool) {
	if len(itens) == 0 {
		// Não dá para escrever T{} nem 0 aqui: T pode ser qualquer coisa. A
		// forma de dizer "o valor zero de T" é declarar uma variável dele.
		var zero T
		return zero, false
	}

	maior := itens[0]
	maiorValor := valor(maior)

	// O operador > está disponível porque a restrição de O é cmp.Ordered.
	// Com O any, esta linha não compilaria.
	for _, item := range itens[1:] {
		if v := valor(item); v > maiorValor {
			maior, maiorValor = item, v
		}
	}

	return maior, true
}

// AgruparPor junta os itens em grupos, indexados pela chave calculada.
//
// make sempre, mesmo para entrada vazia: um map nil aceita leitura mas
// derruba o programa na escrita, como você viu no módulo 02.
func AgruparPor[T any, K comparable](itens []T, chave func(T) K) map[K][]T {
	grupos := make(map[K][]T)

	for _, item := range itens {
		k := chave(item)
		grupos[k] = append(grupos[k], item)
	}

	return grupos
}

// Pilha guarda itens de qualquer tipo, e devolve na ordem inversa.
type Pilha[T any] struct {
	itens []T
}

// Empilhar põe um item no topo.
//
// O receptor é *Pilha[T], com o parâmetro de tipo repetido. O método não pode
// declarar parâmetros de tipo novos: quem declara é o tipo.
func (p *Pilha[T]) Empilhar(item T) {
	p.itens = append(p.itens, item)
}

// Desempilhar tira o item do topo.
func (p *Pilha[T]) Desempilhar() (T, bool) {
	if len(p.itens) == 0 {
		var zero T
		return zero, false
	}

	ultimo := len(p.itens) - 1
	item := p.itens[ultimo]
	p.itens = p.itens[:ultimo]

	return item, true
}

// Tamanho devolve quantos itens estão na pilha.
func (p *Pilha[T]) Tamanho() int {
	return len(p.itens)
}
