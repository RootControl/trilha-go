// Package padaria guarda o código da Padaria do Seu Zé.
//
// Este módulo é diferente dos outros. A padaria dos módulos anteriores chega
// aqui inteira e funcionando: repositório, estoque, erros e venda. O que
// falta são as peças que tornam esse código testável de verdade, e é isso que
// você vai escrever.
package padaria

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// ----------------------------------------------------------------------
// A padaria dos módulos 00 a 05, completa.
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

// ErroDeEstoque é um erro que carrega dados além da mensagem.
type ErroDeEstoque struct {
	Produto    string
	Pedido     int
	Disponivel int
}

// Error faz de *ErroDeEstoque um error.
func (e *ErroDeEstoque) Error() string {
	return fmt.Sprintf("estoque insuficiente de %q: pediram %d, tem %d",
		e.Produto, e.Pedido, e.Disponivel)
}

// Unwrap liga este erro ao sentinela ErrEstoqueInsuficiente.
func (e *ErroDeEstoque) Unwrap() error { return ErrEstoqueInsuficiente }

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// NovoEstoque devolve um Estoque pronto para receber escrita.
func NovoEstoque() Estoque { return make(Estoque) }

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func (e Estoque) Repor(nome string, quantidade int) { e[nome] += quantidade }

// Quantidade devolve quantas unidades existem do produto no estoque.
func (e Estoque) Quantidade(nome string) int { return e[nome] }

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

// BuscadorDeProdutos é tudo que Vender precisa saber sobre onde os produtos
// moram.
type BuscadorDeProdutos interface {
	Buscar(nome string) (Produto, error)
}

var _ BuscadorDeProdutos = (*RepositorioEmMemoria)(nil)

// RepositorioEmMemoria guarda produtos num map, indexados pelo nome.
type RepositorioEmMemoria struct {
	produtos map[string]Produto
}

// NovoRepositorioEmMemoria devolve um repositório pronto para uso.
func NovoRepositorioEmMemoria() *RepositorioEmMemoria {
	return &RepositorioEmMemoria{produtos: make(map[string]Produto)}
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

// ----------------------------------------------------------------------
// Módulo 06: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Relogio existe por um motivo só: tirar time.Now de dentro da regra de
// negócio.
//
// Código que chama time.Now direto é impossível de testar sem trapaça. O
// resultado muda a cada execução, e você acaba escrevendo teste que compara
// com margem de tolerância, ou pior, teste que dorme. Com esta interface, o
// teste decide que horas são.
type Relogio interface {
	Agora() time.Time
}

// RelogioDoSistema é a implementação que vai para produção.
type RelogioDoSistema struct{}

// Agora devolve a hora de verdade.
func (RelogioDoSistema) Agora() time.Time {
	// TODO: implemente este método.
	return time.Time{}
}

// RelogioFixo é a implementação que vai para os testes: ela devolve sempre o
// mesmo instante, escolhido por quem escreveu o teste.
//
// Escrever um dublê como este custa cinco linhas e não precisa de biblioteca
// nenhuma. É o que o módulo 05 comprou quando a interface ficou pequena.
type RelogioFixo struct {
	Momento time.Time
}

// Agora devolve sempre o mesmo instante.
func (r RelogioFixo) Agora() time.Time {
	// TODO: implemente este método.
	return time.Time{}
}

// Pedido é um cliente levando uma quantidade de um produto, com a hora em que
// isso aconteceu.
type Pedido struct {
	Cliente    string
	Produto    Produto
	Quantidade int
	Em         time.Time
}

// NovoPedido monta um pedido carimbado com a hora que o relógio disser.
//
// Repare que o relógio entra como parâmetro. Essa é a diferença entre um
// código testável e um código que você só consegue verificar rodando e
// olhando.
func NovoPedido(relogio Relogio, cliente string, produto Produto, quantidade int) Pedido {
	// TODO: implemente esta função.
	return Pedido{}
}

// Registrador anota o que a padaria fez. Um método só, de novo.
type Registrador interface {
	Registrar(linha string)
}

// RegistradorEmMemoria guarda as linhas numa fatia, para o teste conseguir
// olhar depois. Um dublê que guarda o que recebeu costuma ser chamado de
// espião.
type RegistradorEmMemoria struct {
	linhas []string
}

// Registrar guarda mais uma linha.
func (r *RegistradorEmMemoria) Registrar(linha string) {
	// TODO: implemente este método.
}

// Linhas devolve tudo que foi registrado até agora, na ordem.
//
// Devolva uma cópia. Se você devolver a fatia interna, quem chamou consegue
// alterar o estado do espião por acidente, e você já sabe disso desde o
// módulo 02. slices.Clone resolve.
func (r *RegistradorEmMemoria) Linhas() []string {
	// TODO: implemente este método.
	return nil
}

// VenderComRegistro vende e anota o que aconteceu.
//
// Em caso de sucesso, registra exatamente:
//
//	vendido: 3 de "Pão de queijo" por R$ 13,50
//
// Em caso de falha, registra exatamente:
//
//	falhou: vender 99 de "Pão de queijo": estoque insuficiente de "Pão de queijo": pediram 99, tem 10
//
// ou seja, a palavra "falhou: " seguida da mensagem do erro. Depois de
// registrar, devolve o mesmo total e o mesmo erro que Vender devolveria.
func VenderComRegistro(registrador Registrador, buscador BuscadorDeProdutos, estoque Estoque, nome string, quantidade int) (int, error) {
	// TODO: implemente esta função.
	return 0, nil
}

// DiferencaEntreProdutos descreve, em texto, o que difere entre dois
// produtos. Devolve string vazia quando são iguais.
//
// Uma linha por campo diferente, nesta ordem de campos, separadas por \n:
//
//	Nome: esperado "Broa", recebido "Sonho"
//	PrecoEmCentavos: esperado 300, recebido 700
//	Disponivel: esperado true, recebido false
//
// Isto é, em miniatura, o que uma biblioteca de asserção faz por você. Depois
// de escrever, você vai entender por que muita gente em Go acha que não
// precisa de uma.
func DiferencaEntreProdutos(esperado, recebido Produto) string {
	// TODO: implemente esta função.
	return ""
}
