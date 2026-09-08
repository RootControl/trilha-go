// Package padaria guarda o código da Padaria do Seu Zé.
//
// Até agora tudo que a padaria sabia morria quando o programa terminava.
// Neste módulo os pedidos passam a sobreviver ao desligamento.
package padaria

import (
	"errors"
	"fmt"
	"io"
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
// Módulo 07: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
//
// TODO: acrescente as etiquetas de JSON. Sem elas, o encoding/json usa o nome
// do campo em Go, e o arquivo sairia com "Nome" e "PrecoEmCentavos". A padaria
// combinou com o resto do mundo que o arquivo usa minúsculas com underline:
//
//	nome, preco_em_centavos, disponivel
type Produto struct {
	Nome            string
	PrecoEmCentavos int
	Disponivel      bool
}

// Pedido é um cliente levando uma quantidade de um produto, com a hora.
//
// TODO: acrescente as etiquetas de JSON, seguindo estas regras:
//
//	Cliente     ->  "cliente"
//	Produto     ->  "produto"
//	Quantidade  ->  "quantidade"
//	Em          ->  "em"
//	Observacao  ->  "observacao", e some do arquivo quando está vazia
//	Cancelado   ->  nunca aparece no arquivo
//
// As duas últimas usam vocabulário de etiqueta que o README explica.
type Pedido struct {
	Cliente    string
	Produto    Produto
	Quantidade int
	Em         time.Time
	Observacao string
	Cancelado  bool
}

// SalvarPedidos escreve os pedidos em JSON num io.Writer.
//
// Repare que o parâmetro é io.Writer, e não um nome de arquivo. Essa escolha
// é o assunto do módulo: a função passa a servir para arquivo, para resposta
// HTTP, para um buffer de memória dentro do teste, e para a saída do
// terminal, sem mudar uma linha.
//
// Use json.NewEncoder com SetIndent("", "  "), para o arquivo ficar legível
// por gente. O Encoder termina a saída com uma quebra de linha, e o arquivo
// em testdata/ conta com isso.
func SalvarPedidos(w io.Writer, pedidos []Pedido) error {
	// TODO: implemente esta função.
	return nil
}

// CarregarPedidos lê pedidos em JSON de um io.Reader.
//
// Envolva o erro do decodificador com contexto antes de devolver.
func CarregarPedidos(r io.Reader) ([]Pedido, error) {
	// TODO: implemente esta função.
	return nil, nil
}

// SalvarPedidosEmArquivo grava os pedidos no caminho indicado, criando ou
// sobrescrevendo o arquivo.
//
// O retorno tem nome, err, e isso é de propósito: só assim o defer consegue
// alterar o erro que a função vai devolver. Fechar um arquivo depois de
// escrever pode falhar, e essa falha significa que os dados podem não ter
// chegado ao disco. Um `defer arquivo.Close()` seco joga esse erro fora. O
// README mostra o padrão inteiro.
func SalvarPedidosEmArquivo(caminho string, pedidos []Pedido) (err error) {
	// TODO: implemente esta função.
	return nil
}

// CarregarPedidosDeArquivo lê os pedidos gravados no caminho indicado.
//
// Decisão de produto, e ela precisa estar escrita: arquivo que ainda não
// existe não é erro. É uma padaria que abriu hoje e ainda não vendeu nada.
// Nesse caso devolva uma fatia vazia e nil.
//
// Qualquer outro problema de leitura, incluindo JSON inválido, é erro de
// verdade e sobe envolvido com o caminho.
func CarregarPedidosDeArquivo(caminho string) ([]Pedido, error) {
	// TODO: implemente esta função.
	return nil, nil
}
