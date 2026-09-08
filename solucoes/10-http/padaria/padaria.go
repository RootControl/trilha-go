// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 10. Olhe depois de tentar, não antes.
package padaria

import (
	"encoding/json"
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
// Módulo 10.
// ----------------------------------------------------------------------

// Servidor atende a padaria pela rede.
type Servidor struct {
	cardapio []Produto
	estoque  Estoque
	relogio  Relogio
	mux      *http.ServeMux
}

// Esta linha garante, em tempo de compilação, que *Servidor é um http.Handler.
var _ http.Handler = (*Servidor)(nil)

// NovoServidor monta o servidor e registra as rotas.
func NovoServidor(cardapio []Produto, estoque Estoque, relogio Relogio) *Servidor {
	servidor := &Servidor{
		cardapio: cardapio,
		estoque:  estoque,
		relogio:  relogio,
		mux:      http.NewServeMux(),
	}

	// Desde o Go 1.22 o método e o curinga fazem parte do próprio padrão.
	// Antes disso era preciso um switch em r.Method dentro do handler, ou
	// uma biblioteca de roteamento. O mux também devolve 405 sozinho quando
	// o caminho casa mas o método não.
	servidor.mux.HandleFunc("GET /saude", func(w http.ResponseWriter, r *http.Request) {
		ResponderJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	servidor.mux.HandleFunc("GET /produtos", servidor.listarProdutos)
	servidor.mux.HandleFunc("GET /produtos/{nome}", servidor.buscarProduto)
	servidor.mux.HandleFunc("POST /vendas", servidor.criarVenda)

	return servidor
}

// ServeHTTP faz do *Servidor um http.Handler.
func (s *Servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ResponderJSON escreve o valor como JSON, com o status pedido.
func ResponderJSON(w http.ResponseWriter, status int, valor any) {
	// Cabeçalho primeiro. Depois do WriteHeader os cabeçalhos já foram para o
	// cliente, e mexer neles vira operação silenciosa e inútil.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// Se a codificação falhar aqui, não há o que fazer: o status já foi
	// enviado e não dá para voltar atrás. Por isso valores que podem falhar
	// ao codificar devem ser montados ANTES de começar a responder.
	_ = json.NewEncoder(w).Encode(valor)
}

// ResponderErro escreve uma RespostaDeErro com o status vindo de StatusParaErro.
func ResponderErro(w http.ResponseWriter, err error) {
	ResponderJSON(w, StatusParaErro(err), RespostaDeErro{Erro: err.Error()})
}

// StatusParaErro traduz um erro da padaria em código HTTP.
//
// Repare que o pacote da padaria não sabe o que é HTTP em lugar nenhum, e não
// deve saber. A tradução mora aqui, na borda, e usa errors.Is justamente
// porque os erros chegam embrulhados em camadas de contexto.
func StatusParaErro(err error) int {
	switch {
	case errors.Is(err, ErrProdutoNaoEncontrado):
		return http.StatusNotFound

	case errors.Is(err, ErrEstoqueInsuficiente):
		// 409 e não 400: o pedido está bem formado, o que conflita é o estado
		// atual do estoque. Depois de uma reposição, o mesmo pedido funciona.
		return http.StatusConflict

	case errors.Is(err, ErrQuantidadeInvalida), errors.Is(err, ErrPedidoInvalido):
		return http.StatusBadRequest

	default:
		// Erro que a padaria não previu é problema do servidor, não do
		// cliente. E, em produção, a mensagem crua não deveria vazar aqui.
		return http.StatusInternalServerError
	}
}

// listarProdutos responde GET /produtos com o cardápio inteiro em JSON.
func (s *Servidor) listarProdutos(w http.ResponseWriter, r *http.Request) {
	ResponderJSON(w, http.StatusOK, s.cardapio)
}

// buscarProduto responde GET /produtos/{nome} com um produto, ou 404.
func (s *Servidor) buscarProduto(w http.ResponseWriter, r *http.Request) {
	// PathValue já devolve o segmento decodificado: %C3%A3 chega como ã.
	produto, err := BuscarNoCardapio(s.cardapio, r.PathValue("nome"))
	if err != nil {
		ResponderErro(w, err)
		return
	}

	ResponderJSON(w, http.StatusOK, produto)
}

// criarVenda responde POST /vendas.
func (s *Servidor) criarVenda(w http.ResponseWriter, r *http.Request) {
	var pedido PedidoDeVenda

	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		// O erro do decodificador não vai para o cliente: ele expõe detalhe
		// interno e não ajuda quem chamou. Vira ErrPedidoInvalido.
		ResponderErro(w, fmt.Errorf("corpo ilegível: %w", ErrPedidoInvalido))
		return
	}

	produto, err := BuscarNoCardapio(s.cardapio, pedido.Produto)
	if err != nil {
		ResponderErro(w, err)
		return
	}

	if err := s.estoque.Baixar(pedido.Produto, pedido.Quantidade); err != nil {
		ResponderErro(w, err)
		return
	}

	cliente := pedido.Cliente
	if cliente == "" {
		cliente = "balcão"
	}

	ResponderJSON(w, http.StatusCreated, RespostaDeVenda{
		Cliente:         cliente,
		Produto:         produto,
		Quantidade:      pedido.Quantidade,
		TotalEmCentavos: produto.PrecoEmCentavos * pedido.Quantidade,
		Em:              s.relogio.Agora(),
	})
}

// ComRegistro embrulha um handler e anota uma linha por requisição.
//
// http.HandlerFunc é um tipo função que tem um método ServeHTTP. Converter
// uma função para esse tipo é tudo que falta para ela virar um http.Handler.
// É o mesmo truque que faz o Registrador do módulo 06 funcionar: interface de
// um método é fácil de satisfazer.
func ComRegistro(proximo http.Handler, registrador Registrador) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resposta := &respostaComStatus{ResponseWriter: w}

		proximo.ServeHTTP(resposta, r)

		registrador.Registrar(fmt.Sprintf("%s %s -> %d", r.Method, r.URL.Path, resposta.Status()))
	})
}
