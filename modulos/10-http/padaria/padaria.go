// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria passa a atender pela rede. Sem framework: tudo que
// aparece aqui vem da biblioteca padrão.
package padaria

import (
	"errors"
	"fmt"
	"net/http"
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

// Estoque diz quantas unidades a padaria tem de cada produto.
type Estoque map[string]int

// Relogio tira time.Now de dentro da regra de negócio.
type Relogio interface{ Agora() time.Time }

// RelogioDoSistema é a implementação que vai para produção.
type RelogioDoSistema struct{}

// Agora devolve a hora de verdade.
func (RelogioDoSistema) Agora() time.Time { return time.Now() }

// RelogioFixo devolve sempre o mesmo instante, escolhido pelo teste.
type RelogioFixo struct{ Momento time.Time }

// Agora devolve sempre o mesmo instante.
func (r RelogioFixo) Agora() time.Time { return r.Momento }

// Registrador anota o que a padaria fez, como no módulo 06.
type Registrador interface{ Registrar(linha string) }

// RegistradorEmMemoria guarda as linhas numa fatia.
type RegistradorEmMemoria struct{ linhas []string }

// Registrar guarda mais uma linha.
func (r *RegistradorEmMemoria) Registrar(linha string) { r.linhas = append(r.linhas, linha) }

// Linhas devolve uma cópia de tudo que foi registrado.
func (r *RegistradorEmMemoria) Linhas() []string {
	copia := make([]string, len(r.linhas))
	copy(copia, r.linhas)
	return copia
}

// Erros sentinela da padaria.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")

	// ErrPedidoInvalido é novo: o corpo da requisição não deu para entender.
	ErrPedidoInvalido = errors.New("pedido inválido")
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

// BuscarNoCardapio procura um produto pelo nome exato.
func BuscarNoCardapio(cardapio []Produto, nome string) (Produto, error) {
	for _, produto := range cardapio {
		if produto.Nome == nome {
			return produto, nil
		}
	}
	return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
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

// ----------------------------------------------------------------------
// Formatos que entram e saem pela rede.
// ----------------------------------------------------------------------

// PedidoDeVenda é o corpo que o cliente manda em POST /vendas.
type PedidoDeVenda struct {
	Cliente    string `json:"cliente"`
	Produto    string `json:"produto"`
	Quantidade int    `json:"quantidade"`
}

// RespostaDeVenda é o corpo devolvido quando a venda dá certo.
type RespostaDeVenda struct {
	Cliente         string    `json:"cliente"`
	Produto         Produto   `json:"produto"`
	Quantidade      int       `json:"quantidade"`
	TotalEmCentavos int       `json:"total_em_centavos"`
	Em              time.Time `json:"em"`
}

// RespostaDeErro é o corpo devolvido em qualquer resposta de erro.
//
// Uma API que responde erro em HTML numa hora e em JSON noutra é um pesadelo
// para quem consome. Escolha um formato e use sempre.
type RespostaDeErro struct {
	Erro string `json:"erro"`
}

// ----------------------------------------------------------------------
// Exemplo resolvido: capturando o status para o middleware.
// ----------------------------------------------------------------------

// respostaComStatus embute um http.ResponseWriter e lembra o status enviado.
//
// Isto é composição por embutimento, do módulo 05: por ter um
// http.ResponseWriter embutido sem nome de campo, este tipo já ganha todos os
// métodos dele e já é um http.ResponseWriter. Só WriteHeader é reescrito.
//
// O padrão existe porque a interface http.ResponseWriter não tem como
// PERGUNTAR qual status foi enviado, só como enviar. Para registrar o status
// numa linha de log, é preciso interceptar.
type respostaComStatus struct {
	http.ResponseWriter
	status int
}

// WriteHeader guarda o status e repassa para o writer de verdade.
func (r *respostaComStatus) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Status devolve o status enviado, ou 200 quando o handler não chamou
// WriteHeader.
//
// Um handler que só escreve o corpo nunca chama WriteHeader, e o net/http
// manda 200 sozinho. Sem este padrão, o log registraria zero.
func (r *respostaComStatus) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

// ----------------------------------------------------------------------
// Módulo 10: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Servidor atende a padaria pela rede.
//
// Ele guarda tudo de que os handlers precisam. Repare que não há variável
// global em lugar nenhum: cada teste monta o seu próprio servidor, com o seu
// próprio estoque e o seu próprio relógio.
type Servidor struct {
	cardapio []Produto
	estoque  Estoque
	relogio  Relogio
	mux      *http.ServeMux
}

// NovoServidor monta o servidor e registra as rotas.
//
// Registre exatamente estas quatro, no padrão do Go 1.22, que aceita método e
// curinga no mesmo texto:
//
//	GET /saude            responde 200 com {"status":"ok"}
//	GET /produtos         responde 200 com o cardápio
//	GET /produtos/{nome}  responde 200 com um produto, ou 404
//	POST /vendas          responde 201 com a venda
//
// O curinga {nome} é lido depois com r.PathValue("nome").
func NovoServidor(cardapio []Produto, estoque Estoque, relogio Relogio) *Servidor {
	// TODO: implemente esta função.
	return nil
}

// ServeHTTP faz do *Servidor um http.Handler.
//
// http.Handler é uma interface de um método só, exatamente como as do módulo
// 05:
//
//	type Handler interface {
//		ServeHTTP(ResponseWriter, *Request)
//	}
//
// Implemente delegando para o mux interno. É isso que permite passar o
// servidor inteiro para um middleware, ou para o httptest, como se fosse um
// handler qualquer.
func (s *Servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implemente este método.
}

// ResponderJSON escreve o valor como JSON, com o status pedido.
//
// A ORDEM IMPORTA e é a pegadinha mais frequente do net/http:
//
//  1. w.Header().Set("Content-Type", "application/json; charset=utf-8")
//  2. w.WriteHeader(status)
//  3. json.NewEncoder(w).Encode(valor)
//
// Depois do WriteHeader os cabeçalhos já foram enviados, e mexer neles não
// faz nada, sem erro e sem aviso. O teste confere o Content-Type justamente
// para pegar essa inversão.
func ResponderJSON(w http.ResponseWriter, status int, valor any) {
	// TODO: implemente esta função.
}

// ResponderErro escreve uma RespostaDeErro com o status vindo de
// StatusParaErro, usando a mensagem do erro.
func ResponderErro(w http.ResponseWriter, err error) {
	// TODO: implemente esta função.
}

// StatusParaErro traduz um erro da padaria em código HTTP.
//
// Este é o ponto de encontro entre o módulo 03 e este. O domínio não sabe o
// que é HTTP, e não deve saber: quem traduz é a camada de rede, aqui, com
// errors.Is atravessando os embrulhos.
//
//	ErrProdutoNaoEncontrado  ->  404
//	ErrEstoqueInsuficiente   ->  409, porque o estado atual conflita com o
//	                             pedido, e o cliente pode tentar de novo
//	                             depois de uma reposição
//	ErrQuantidadeInvalida    ->  400
//	ErrPedidoInvalido        ->  400
//	qualquer outro           ->  500
func StatusParaErro(err error) int {
	// TODO: implemente esta função.
	return http.StatusInternalServerError
}

// listarProdutos responde GET /produtos com o cardápio inteiro em JSON.
func (s *Servidor) listarProdutos(w http.ResponseWriter, r *http.Request) {
	// TODO: implemente este método.
}

// buscarProduto responde GET /produtos/{nome} com um produto, ou 404.
func (s *Servidor) buscarProduto(w http.ResponseWriter, r *http.Request) {
	// TODO: implemente este método.
}

// criarVenda responde POST /vendas.
//
// O que precisa acontecer:
//
//  1. Decodificar o corpo em PedidoDeVenda. Corpo ilegível vira
//     ErrPedidoInvalido.
//  2. Buscar o produto no cardápio.
//  3. Dar baixa no estoque.
//  4. Responder 201 com a RespostaDeVenda preenchida, carimbada com
//     s.relogio.Agora().
//
// Qualquer erro no caminho vai para ResponderErro, que já escolhe o status.
// Cliente vazio vira "balcão", como no módulo 08.
func (s *Servidor) criarVenda(w http.ResponseWriter, r *http.Request) {
	// TODO: implemente este método.
}

// ComRegistro embrulha um handler e anota uma linha por requisição.
//
// A assinatura func(http.Handler) http.Handler é a forma canônica de
// middleware em Go. Ela não é uma regra da linguagem nem de um framework: é
// só o que sai naturalmente de uma interface de um método.
//
// Anote exatamente uma linha por requisição, neste formato:
//
//	GET /produtos -> 200
//
// Use respostaComStatus, já pronta acima, para descobrir o status. E use
// http.HandlerFunc para transformar a sua função num http.Handler: ela é um
// tipo função que tem um método ServeHTTP, e por isso satisfaz a interface.
func ComRegistro(proximo http.Handler, registrador Registrador) http.Handler {
	// TODO: implemente esta função.
	return proximo
}
