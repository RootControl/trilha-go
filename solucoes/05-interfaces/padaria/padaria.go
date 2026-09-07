// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 05. Olhe depois de tentar, não antes.
package padaria

import (
	"errors"
	"fmt"
	"slices"
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

// BuscadorDeProdutos é tudo que Vender precisa saber sobre onde os produtos
// moram.
type BuscadorDeProdutos interface {
	Buscar(nome string) (Produto, error)
}

var _ BuscadorDeProdutos = (*RepositorioEmMemoria)(nil)

// ErroDeEstoque é um erro que carrega dados além da mensagem.
type ErroDeEstoque struct {
	Produto    string
	Pedido     int
	Disponivel int
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
		return &ErroDeEstoque{Produto: nome, Pedido: quantidade, Disponivel: quantidadeAtual}
	}

	e[nome] = quantidadeAtual - quantidade

	return nil
}

// ----------------------------------------------------------------------
// Módulo 05.
// ----------------------------------------------------------------------

// Error faz de *ErroDeEstoque um error.
//
// Receptor de ponteiro, e por isso quem satisfaz a interface error é
// *ErroDeEstoque, não ErroDeEstoque. É por esse motivo que Baixar devolve
// &ErroDeEstoque{...} com o e comercial, e que errors.As recebe um **ErroDeEstoque.
func (e *ErroDeEstoque) Error() string {
	return fmt.Sprintf("estoque insuficiente de %q: pediram %d, tem %d",
		e.Produto, e.Pedido, e.Disponivel)
}

// Unwrap liga este erro ao sentinela ErrEstoqueInsuficiente.
//
// Uma linha, e todo o código escrito no módulo 03 continua funcionando sem
// alteração nenhuma. Quem usava errors.Is nem fica sabendo que apareceu um
// tipo novo no caminho.
func (e *ErroDeEstoque) Unwrap() error {
	return ErrEstoqueInsuficiente
}

// RepositorioEmMemoria guarda produtos num map, indexados pelo nome.
type RepositorioEmMemoria struct {
	produtos map[string]Produto
}

// NovoRepositorioEmMemoria devolve um repositório pronto para uso.
//
// A função devolve *RepositorioEmMemoria, o tipo concreto, e não a interface.
// Quem chama fica com acesso a tudo que o tipo sabe fazer, e continua livre
// para guardar o resultado numa variável de interface se quiser. Devolver
// interface aqui só tiraria opções de quem chama.
func NovoRepositorioEmMemoria() *RepositorioEmMemoria {
	return &RepositorioEmMemoria{
		produtos: make(map[string]Produto),
	}
}

// Salvar guarda o produto, sobrescrevendo se já existir um com o mesmo nome.
func (r *RepositorioEmMemoria) Salvar(produto Produto) error {
	if produto.Nome == "" {
		return fmt.Errorf("salvar produto sem nome: %w", ErrProdutoInvalido)
	}

	r.produtos[produto.Nome] = produto

	return nil
}

// Buscar procura um produto pelo nome exato.
func (r *RepositorioEmMemoria) Buscar(nome string) (Produto, error) {
	produto, existe := r.produtos[nome]
	if !existe {
		return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
	}

	return produto, nil
}

// Todos devolve os produtos guardados, em ordem alfabética de nome.
func (r *RepositorioEmMemoria) Todos() []Produto {
	// make com capacidade já reservada evita que o append realoque o array
	// várias vezes. Aqui o tamanho final é conhecido, então vale a pena.
	nomes := make([]string, 0, len(r.produtos))
	for nome := range r.produtos {
		nomes = append(nomes, nome)
	}
	slices.Sort(nomes)

	todos := make([]Produto, 0, len(nomes))
	for _, nome := range nomes {
		todos = append(todos, r.produtos[nome])
	}

	return todos
}

// Vender acha o produto, dá baixa no estoque e devolve o total a pagar.
//
// O corpo é igual ao do módulo 03. O que mudou foi só o primeiro parâmetro,
// que virou interface, e com isso esta função passou a funcionar com qualquer
// coisa que saiba buscar produto.
func Vender(buscador BuscadorDeProdutos, estoque Estoque, nome string, quantidade int) (int, error) {
	produto, err := buscador.Buscar(nome)
	if err != nil {
		return 0, fmt.Errorf("vender %d de %q: %w", quantidade, nome, err)
	}

	if err := estoque.Baixar(nome, quantidade); err != nil {
		return 0, fmt.Errorf("vender %d de %q: %w", quantidade, nome, err)
	}

	return produto.PrecoEmCentavos * quantidade, nil
}
