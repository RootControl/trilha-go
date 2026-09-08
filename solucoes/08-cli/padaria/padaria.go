// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 08. Olhe depois de tentar, não antes.
package padaria

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
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
// Módulo 08.
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
// A função recebe io.Writer, e não escreve com fmt.Println. fmt.Println vai
// sempre para o os.Stdout do processo, o que torna a função impossível de
// testar sem truque. Fprintln escreve onde mandarem.
func EscreverCardapio(saida io.Writer, cardapio []Produto, comoJSON bool) error {
	if comoJSON {
		return escreverJSON(saida, cardapio)
	}

	for _, produto := range cardapio {
		if _, err := fmt.Fprintln(saida, Descrever(produto)); err != nil {
			return fmt.Errorf("escrever cardápio: %w", err)
		}
	}

	return nil
}

// EscreverPedidos escreve os pedidos na saída.
func EscreverPedidos(saida io.Writer, pedidos []Pedido, comoJSON bool) error {
	if comoJSON {
		return escreverJSON(saida, pedidos)
	}

	if len(pedidos) == 0 {
		if _, err := fmt.Fprintln(saida, "nenhum pedido registrado"); err != nil {
			return fmt.Errorf("escrever pedidos: %w", err)
		}
		return nil
	}

	for _, pedido := range pedidos {
		total := pedido.Produto.PrecoEmCentavos * pedido.Quantidade

		// O layout "2006-01-02 15:04" não é um formato inventado: é a data de
		// referência do Go. O README explica.
		_, err := fmt.Fprintf(saida, "%s | %s | %d x %s | %s\n",
			pedido.Em.Format("2006-01-02 15:04"),
			pedido.Cliente,
			pedido.Quantidade,
			pedido.Produto.Nome,
			FormatarPreco(total))
		if err != nil {
			return fmt.Errorf("escrever pedidos: %w", err)
		}
	}

	return nil
}

// Executar é o programa inteiro, sem depender de nada global.
func Executar(args []string, saida io.Writer, erros io.Writer, relogio Relogio) error {
	// flag.Parse, o do pacote, lê o os.Args global e escreve no os.Stderr
	// global. Um FlagSet próprio recebe os argumentos que você der e escreve
	// onde você mandar, e é isso que torna esta função testável.
	conjunto := flag.NewFlagSet("padaria", flag.ContinueOnError)
	conjunto.SetOutput(erros)

	caminho := conjunto.String("arquivo", "pedidos.json", "onde os pedidos são gravados")
	cliente := conjunto.String("cliente", "balcão", "nome de quem está comprando")
	comoJSON := conjunto.Bool("json", false, "escreve a saída em JSON")

	conjunto.Usage = func() {
		fmt.Fprint(conjunto.Output(), UsoDoPrograma)
		conjunto.PrintDefaults()
	}

	if err := conjunto.Parse(args); err != nil {
		// Pedir ajuda com -h não é falhar. O FlagSet já imprimiu o uso.
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("%w: %v", ErrUsoInvalido, err)
	}

	restantes := conjunto.Args()
	if len(restantes) == 0 {
		conjunto.Usage()
		return fmt.Errorf("nenhum comando informado: %w", ErrUsoInvalido)
	}

	switch comando := restantes[0]; comando {
	case "cardapio":
		return EscreverCardapio(saida, CardapioPadrao(), *comoJSON)

	case "pedidos":
		pedidos, err := CarregarPedidosDeArquivo(*caminho)
		if err != nil {
			return err
		}
		return EscreverPedidos(saida, pedidos, *comoJSON)

	case "vender":
		if len(restantes) != 3 {
			conjunto.Usage()
			return fmt.Errorf("vender precisa de produto e quantidade: %w", ErrUsoInvalido)
		}
		return venderPeloTerminal(saida, *caminho, *cliente, restantes[1], restantes[2], relogio)

	default:
		conjunto.Usage()
		return fmt.Errorf("comando desconhecido %q: %w", comando, ErrUsoInvalido)
	}
}

// venderPeloTerminal é o comando vender, separado só para Executar continuar
// legível. Um switch de despacho fica melhor quando cada braço tem uma linha.
func venderPeloTerminal(saida io.Writer, caminho, cliente, nomeDoProduto, quantidadeTexto string, relogio Relogio) error {
	quantidade, err := strconv.Atoi(quantidadeTexto)
	if err != nil {
		return fmt.Errorf("quantidade %q não é um número: %w", quantidadeTexto, ErrQuantidadeInvalida)
	}

	produto, err := BuscarNoCardapio(CardapioPadrao(), nomeDoProduto)
	if err != nil {
		return err
	}

	if err := EstoquePadrao().Baixar(nomeDoProduto, quantidade); err != nil {
		return err
	}

	pedidos, err := CarregarPedidosDeArquivo(caminho)
	if err != nil {
		return err
	}

	pedidos = append(pedidos, Pedido{
		Cliente:    cliente,
		Produto:    produto,
		Quantidade: quantidade,
		Em:         relogio.Agora(),
	})

	if err := SalvarPedidosEmArquivo(caminho, pedidos); err != nil {
		return err
	}

	total := produto.PrecoEmCentavos * quantidade
	if _, err := fmt.Fprintf(saida, "vendido: %d de %q por %s\n",
		quantidade, nomeDoProduto, FormatarPreco(total)); err != nil {
		return fmt.Errorf("escrever confirmação: %w", err)
	}

	return nil
}
