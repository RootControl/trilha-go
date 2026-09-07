// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 06. Olhe depois de tentar, não antes.
package padaria

import (
	"errors"
	"fmt"
	"slices"
	"strings"
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
// Módulo 06.
// ----------------------------------------------------------------------

// Relogio existe para tirar time.Now de dentro da regra de negócio.
type Relogio interface {
	Agora() time.Time
}

// RelogioDoSistema é a implementação que vai para produção.
//
// Repare que a struct não tem campo nenhum. Um tipo vazio assim custa zero
// byte e existe só para pendurar o método e satisfazer a interface.
type RelogioDoSistema struct{}

// Agora devolve a hora de verdade.
//
// Esta é a única linha do pacote que fala com o relógio do sistema. Todo o
// resto do código recebe a hora de fora, e por isso é testável.
func (RelogioDoSistema) Agora() time.Time {
	return time.Now()
}

// RelogioFixo devolve sempre o mesmo instante, escolhido pelo teste.
type RelogioFixo struct {
	Momento time.Time
}

// Agora devolve sempre o mesmo instante.
func (r RelogioFixo) Agora() time.Time {
	return r.Momento
}

// Pedido é um cliente levando uma quantidade de um produto, com hora.
type Pedido struct {
	Cliente    string
	Produto    Produto
	Quantidade int
	Em         time.Time
}

// NovoPedido monta um pedido carimbado com a hora que o relógio disser.
func NovoPedido(relogio Relogio, cliente string, produto Produto, quantidade int) Pedido {
	return Pedido{
		Cliente:    cliente,
		Produto:    produto,
		Quantidade: quantidade,
		Em:         relogio.Agora(),
	}
}

// Registrador anota o que a padaria fez.
type Registrador interface {
	Registrar(linha string)
}

// RegistradorEmMemoria guarda as linhas numa fatia, para o teste olhar depois.
type RegistradorEmMemoria struct {
	linhas []string
}

// Registrar guarda mais uma linha.
//
// O valor zero de RegistradorEmMemoria já funciona: append em fatia nil
// funciona, então não é preciso um NovoRegistrador. Compare com o map do
// módulo 02, que precisava.
func (r *RegistradorEmMemoria) Registrar(linha string) {
	r.linhas = append(r.linhas, linha)
}

// Linhas devolve uma cópia de tudo que foi registrado, na ordem.
//
// A cópia não é preciosismo. Sem ela, quem recebe a fatia enxerga o mesmo
// array de dentro do espião e consegue alterá-lo, exatamente como no módulo 02.
func (r *RegistradorEmMemoria) Linhas() []string {
	return slices.Clone(r.linhas)
}

// VenderComRegistro vende e anota o que aconteceu.
func VenderComRegistro(registrador Registrador, buscador BuscadorDeProdutos, estoque Estoque, nome string, quantidade int) (int, error) {
	total, err := Vender(buscador, estoque, nome, quantidade)
	if err != nil {
		registrador.Registrar("falhou: " + err.Error())
		return 0, err
	}

	registrador.Registrar(fmt.Sprintf("vendido: %d de %q por %s",
		quantidade, nome, FormatarPreco(total)))

	return total, nil
}

// DiferencaEntreProdutos descreve, em texto, o que difere entre dois produtos.
//
// strings.Join sobre uma fatia nil devolve string vazia, então o caso "são
// iguais" sai de graça, sem if nenhum.
func DiferencaEntreProdutos(esperado, recebido Produto) string {
	var linhas []string

	if esperado.Nome != recebido.Nome {
		linhas = append(linhas, fmt.Sprintf("Nome: esperado %q, recebido %q",
			esperado.Nome, recebido.Nome))
	}
	if esperado.PrecoEmCentavos != recebido.PrecoEmCentavos {
		linhas = append(linhas, fmt.Sprintf("PrecoEmCentavos: esperado %d, recebido %d",
			esperado.PrecoEmCentavos, recebido.PrecoEmCentavos))
	}
	if esperado.Disponivel != recebido.Disponivel {
		linhas = append(linhas, fmt.Sprintf("Disponivel: esperado %t, recebido %t",
			esperado.Disponivel, recebido.Disponivel))
	}

	return strings.Join(linhas, "\n")
}
