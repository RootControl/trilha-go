// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria vira um programa que alguém consegue rodar. Toda a
// lógica continua aqui, na biblioteca; o pacote main lá em cmd/padaria tem
// nove linhas e não decide nada.
package padaria

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores, tudo pronto.
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
	Observacao string    `json:"observacao,omitempty"`
}

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// Relogio tira time.Now de dentro da regra de negócio, como no módulo 06.
type Relogio interface {
	Agora() time.Time
}

// RelogioDoSistema é a implementação que vai para produção.
type RelogioDoSistema struct{}

// Agora devolve a hora de verdade.
func (RelogioDoSistema) Agora() time.Time { return time.Now() }

// RelogioFixo devolve sempre o mesmo instante, escolhido pelo teste.
type RelogioFixo struct{ Momento time.Time }

// Agora devolve sempre o mesmo instante.
func (r RelogioFixo) Agora() time.Time { return r.Momento }

// Erros sentinela da padaria.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")

	// ErrUsoInvalido é novo neste módulo: significa que quem chamou o
	// programa errou a linha de comando, e não que a padaria recusou a venda.
	// A diferença importa para quem escreve script em volta do programa.
	ErrUsoInvalido = errors.New("uso inválido")
)

// FormatarPreco escreve centavos no formato brasileiro.
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}

// Descrever escreve o produto do jeito que aparece na placa da padaria.
func Descrever(p Produto) string {
	if !p.Disponivel {
		return fmt.Sprintf("%s: esgotado", p.Nome)
	}
	return fmt.Sprintf("%s: %s", p.Nome, FormatarPreco(p.PrecoEmCentavos))
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
		return fmt.Errorf("baixar %d de %q, só tem %d: %w",
			quantidade, nome, quantidadeAtual, ErrEstoqueInsuficiente)
	}

	e[nome] = quantidadeAtual - quantidade

	return nil
}

// CardapioPadrao é o cardápio fixo da padaria, em ordem alfabética.
func CardapioPadrao() []Produto {
	return []Produto{
		{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
	}
}

// EstoquePadrao é o estoque com que a padaria abre o dia.
func EstoquePadrao() Estoque {
	estoque := NovoEstoque()
	estoque.Repor("Broa", 5)
	estoque.Repor("Pão de queijo", 10)
	estoque.Repor("Pão francês", 50)
	return estoque
}

// BuscarNoCardapio procura um produto pelo nome exato.
func BuscarNoCardapio(cardapio []Produto, nome string) (Produto, error) {
	for _, produto := range cardapio {
		if produto.Nome == nome {
			return produto, nil
		}
	}
	return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
}

// SalvarPedidosEmArquivo grava os pedidos no caminho indicado.
func SalvarPedidosEmArquivo(caminho string, pedidos []Pedido) (err error) {
	arquivo, err := os.Create(caminho)
	if err != nil {
		return fmt.Errorf("criar %q: %w", caminho, err)
	}

	defer func() {
		if erroAoFechar := arquivo.Close(); erroAoFechar != nil && err == nil {
			err = fmt.Errorf("fechar %q: %w", caminho, erroAoFechar)
		}
	}()

	codificador := json.NewEncoder(arquivo)
	codificador.SetIndent("", "  ")
	if err := codificador.Encode(pedidos); err != nil {
		return fmt.Errorf("codificar pedidos em %q: %w", caminho, err)
	}

	return nil
}

// CarregarPedidosDeArquivo lê os pedidos gravados no caminho indicado.
// Arquivo inexistente devolve fatia vazia e nil.
func CarregarPedidosDeArquivo(caminho string) ([]Pedido, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Pedido{}, nil
		}
		return nil, fmt.Errorf("abrir %q: %w", caminho, err)
	}
	defer arquivo.Close()

	var pedidos []Pedido
	if err := json.NewDecoder(arquivo).Decode(&pedidos); err != nil {
		return nil, fmt.Errorf("decodificar %q: %w", caminho, err)
	}

	return pedidos, nil
}

// escreverJSON é um auxiliar para as duas funções de escrita.
func escreverJSON(saida io.Writer, valor any) error {
	codificador := json.NewEncoder(saida)
	codificador.SetIndent("", "  ")

	if err := codificador.Encode(valor); err != nil {
		return fmt.Errorf("codificar saída: %w", err)
	}

	return nil
}

// ----------------------------------------------------------------------
// Módulo 08: é aqui que você trabalha.
// ----------------------------------------------------------------------

// UsoDoPrograma é o texto de ajuda do comando.
const UsoDoPrograma = `padaria: o balcão da Padaria do Seu Zé

uso:
  padaria [opções] cardapio
  padaria [opções] vender <produto> <quantidade>
  padaria [opções] pedidos

opções:
`

// EscreverCardapio escreve o cardápio na saída.
//
// Em texto, uma linha por produto, na ordem em que vieram, usando Descrever:
//
//	Broa: R$ 3,00
//	Pão de queijo: R$ 4,50
//	Pão francês: R$ 1,00
//	Sonho: esgotado
//
// Em JSON, use o auxiliar escreverJSON com a fatia inteira.
func EscreverCardapio(saida io.Writer, cardapio []Produto, comoJSON bool) error {
	// TODO: implemente esta função.
	return nil
}

// EscreverPedidos escreve os pedidos na saída.
//
// Sem nenhum pedido, em texto, escreva exatamente uma linha:
//
//	nenhum pedido registrado
//
// Com pedidos, uma linha por pedido, neste formato:
//
//	2026-03-15 07:30 | balcão | 3 x Pão de queijo | R$ 13,50
//
// A data usa o layout "2006-01-02 15:04". Esse número não é aleatório e o
// README explica de onde ele vem. O valor no fim é preço vezes quantidade.
//
// Em JSON, escreva a fatia inteira com escreverJSON, inclusive quando vazia.
func EscreverPedidos(saida io.Writer, pedidos []Pedido, comoJSON bool) error {
	// TODO: implemente esta função.
	return nil
}

// Executar é o programa inteiro, sem depender de nada global.
//
// Repare no que ela recebe: os argumentos, para onde escrever a saída, para
// onde escrever os erros, e o relógio. Nada de os.Args, nada de os.Stdout,
// nada de time.Now. É isso que permite ao teste rodar o programa completo
// dentro da memória, e é o assunto do módulo.
//
// O que ela precisa fazer, em ordem:
//
//  1. Montar um flag.NewFlagSet chamado "padaria", com flag.ContinueOnError,
//     mandando a saída dele para erros.
//  2. Declarar três opções:
//     -arquivo string, padrão "pedidos.json", onde os pedidos são gravados
//     -cliente string, padrão "balcão", nome de quem está comprando
//     -json bool, padrão false, escreve a saída em JSON
//  3. Definir conjunto.Usage para imprimir UsoDoPrograma seguido das opções,
//     com conjunto.PrintDefaults.
//  4. Chamar conjunto.Parse(args). Se der erro, devolver nil quando for
//     flag.ErrHelp, porque pedir ajuda não é falhar, e o erro envolvido com
//     ErrUsoInvalido nos outros casos.
//  5. Olhar conjunto.Args(). Vazio devolve ErrUsoInvalido dizendo que nenhum
//     comando foi informado.
//  6. Despachar pelo primeiro argumento:
//     "cardapio" chama EscreverCardapio com CardapioPadrao()
//     "pedidos"  carrega do arquivo e chama EscreverPedidos
//     "vender"   exige exatamente mais dois argumentos, produto e quantidade
//     qualquer outro devolve ErrUsoInvalido citando o comando entre aspas
//
// O comando vender precisa: converter a quantidade com strconv.Atoi e
// envolver ErrQuantidadeInvalida se não for número, buscar o produto no
// cardápio, dar baixa no estoque padrão, montar o Pedido carimbado com
// relogio.Agora(), acrescentar aos pedidos já gravados, salvar tudo de volta
// no arquivo, e escrever na saída exatamente:
//
//	vendido: 3 de "Pão de queijo" por R$ 13,50
func Executar(args []string, saida io.Writer, erros io.Writer, relogio Relogio) error {
	// TODO: implemente esta função.
	return nil
}
