// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 07. Olhe depois de tentar, não antes.
package padaria

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// Erros sentinela da padaria.
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
func NovoEstoque() Estoque { return make(Estoque) }

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func (e Estoque) Repor(nome string, quantidade int) { e[nome] += quantidade }

// Quantidade devolve quantas unidades existem do produto no estoque.
func (e Estoque) Quantidade(nome string) int { return e[nome] }

// ----------------------------------------------------------------------
// Módulo 07.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string `json:"nome"`
	PrecoEmCentavos int    `json:"preco_em_centavos"`
	Disponivel      bool   `json:"disponivel"`
}

// Pedido é um cliente levando uma quantidade de um produto, com a hora.
//
// A etiqueta omitempty faz o campo sumir do arquivo quando está no valor
// zero. A etiqueta "-" faz o campo nunca aparecer, nem na escrita nem na
// leitura: Cancelado é estado de execução, não coisa que se grava.
type Pedido struct {
	Cliente    string    `json:"cliente"`
	Produto    Produto   `json:"produto"`
	Quantidade int       `json:"quantidade"`
	Em         time.Time `json:"em"`
	Observacao string    `json:"observacao,omitempty"`
	Cancelado  bool      `json:"-"`
}

// SalvarPedidos escreve os pedidos em JSON num io.Writer.
//
// Encoder escreve direto no destino, sem montar o documento inteiro na
// memória antes. Para uma padaria tanto faz; para um arquivo de um giga, é a
// diferença entre funcionar e não funcionar.
func SalvarPedidos(w io.Writer, pedidos []Pedido) error {
	codificador := json.NewEncoder(w)
	codificador.SetIndent("", "  ")

	if err := codificador.Encode(pedidos); err != nil {
		return fmt.Errorf("codificar pedidos: %w", err)
	}

	return nil
}

// CarregarPedidos lê pedidos em JSON de um io.Reader.
func CarregarPedidos(r io.Reader) ([]Pedido, error) {
	var pedidos []Pedido

	if err := json.NewDecoder(r).Decode(&pedidos); err != nil {
		return nil, fmt.Errorf("decodificar pedidos: %w", err)
	}

	return pedidos, nil
}

// SalvarPedidosEmArquivo grava os pedidos no caminho indicado.
//
// O defer aqui não é o `defer arquivo.Close()` de uma linha que aparece em
// todo tutorial. Numa escrita, Close pode devolver erro, e esse erro quer
// dizer que os dados podem não ter chegado ao disco. Descartar é perder a
// única chance de saber.
//
// O retorno nomeado é o que permite ao defer corrigir o erro depois do
// return. A condição `err == nil` garante que um erro de escrita, que é mais
// informativo, não seja sobrescrito por um erro de fechamento.
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

	if err := SalvarPedidos(arquivo, pedidos); err != nil {
		return fmt.Errorf("salvar em %q: %w", caminho, err)
	}

	return nil
}

// CarregarPedidosDeArquivo lê os pedidos gravados no caminho indicado.
func CarregarPedidosDeArquivo(caminho string) ([]Pedido, error) {
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		// Arquivo que ainda não existe é padaria que ainda não vendeu.
		// errors.Is atravessa o *PathError que o os devolve e chega no
		// sentinela da biblioteca padrão.
		if errors.Is(err, fs.ErrNotExist) {
			return []Pedido{}, nil
		}

		return nil, fmt.Errorf("ler %q: %w", caminho, err)
	}

	pedidos, err := CarregarPedidos(bytes.NewReader(conteudo))
	if err != nil {
		return nil, fmt.Errorf("ler %q: %w", caminho, err)
	}

	return pedidos, nil
}
