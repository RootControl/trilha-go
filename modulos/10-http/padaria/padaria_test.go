package padaria

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var momentoDeTeste = time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)

// servidorDeTeste monta um servidor novo, com estoque e relógio próprios.
//
// Nenhum teste deste arquivo abre uma porta de rede. httptest.NewRequest monta
// a requisição e httptest.NewRecorder recebe a resposta, tudo em memória, o
// que é rápido e não conflita quando os testes rodam ao mesmo tempo.
func servidorDeTeste(t *testing.T) *Servidor {
	t.Helper()

	servidor := NovoServidor(CardapioPadrao(), EstoquePadrao(), RelogioFixo{Momento: momentoDeTeste})
	if servidor == nil {
		t.Fatal("NovoServidor devolveu nil")
	}

	return servidor
}

func chamar(t *testing.T, servidor http.Handler, metodo, alvo string, corpo string) *httptest.ResponseRecorder {
	t.Helper()

	var leitor *strings.Reader
	if corpo != "" {
		leitor = strings.NewReader(corpo)
	} else {
		leitor = strings.NewReader("")
	}

	requisicao := httptest.NewRequest(metodo, alvo, leitor)
	gravador := httptest.NewRecorder()
	servidor.ServeHTTP(gravador, requisicao)

	return gravador
}

func decodificar[T any](t *testing.T, gravador *httptest.ResponseRecorder) T {
	t.Helper()

	var valor T
	if err := json.NewDecoder(gravador.Body).Decode(&valor); err != nil {
		t.Fatalf("resposta não é JSON válido: %v\ncorpo: %s", err, gravador.Body.String())
	}

	return valor
}

func TestSaude(t *testing.T) {
	gravador := chamar(t, servidorDeTeste(t), http.MethodGet, "/saude", "")

	if gravador.Code != http.StatusOK {
		t.Errorf("status = %d, esperado 200", gravador.Code)
	}

	corpo := decodificar[map[string]string](t, gravador)
	if corpo["status"] != "ok" {
		t.Errorf(`status = %q, esperado "ok"`, corpo["status"])
	}
}

func TestListarProdutos(t *testing.T) {
	gravador := chamar(t, servidorDeTeste(t), http.MethodGet, "/produtos", "")

	if gravador.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200", gravador.Code)
	}

	// Result() devolve o que REALMENTE teria ido para o cliente, com os
	// cabeçalhos congelados no instante do WriteHeader. Header(), sem o
	// Result, devolve o mapa vivo, que continua aceitando escrita depois e
	// esconderia a ordem invertida.
	if tipo := gravador.Result().Header.Get("Content-Type"); !strings.HasPrefix(tipo, "application/json") {
		t.Errorf("Content-Type = %q, esperado application/json; confira a ordem em ResponderJSON", tipo)
	}

	produtos := decodificar[[]Produto](t, gravador)
	if len(produtos) != 4 {
		t.Fatalf("vieram %d produtos, esperado 4", len(produtos))
	}
	if produtos[0].Nome != "Broa" {
		t.Errorf("primeiro produto = %q, esperado Broa", produtos[0].Nome)
	}
}

func TestBuscarProduto(t *testing.T) {
	// O nome tem espaço e acento, então precisa ser escapado na URL. O
	// roteador desescapa antes de entregar em PathValue.
	alvo := "/produtos/" + url.PathEscape("Pão de queijo")

	gravador := chamar(t, servidorDeTeste(t), http.MethodGet, alvo, "")

	if gravador.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200; corpo: %s", gravador.Code, gravador.Body.String())
	}

	produto := decodificar[Produto](t, gravador)
	if produto.Nome != "Pão de queijo" {
		t.Errorf("Nome = %q, esperado Pão de queijo", produto.Nome)
	}
	if produto.PrecoEmCentavos != 450 {
		t.Errorf("PrecoEmCentavos = %d, esperado 450", produto.PrecoEmCentavos)
	}
}

func TestBuscarProdutoInexistente(t *testing.T) {
	gravador := chamar(t, servidorDeTeste(t), http.MethodGet, "/produtos/Croissant", "")

	if gravador.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado 404", gravador.Code)
	}

	// Erro também responde em JSON. Uma API que às vezes responde HTML é um
	// pesadelo para quem consome.
	erro := decodificar[RespostaDeErro](t, gravador)
	if !strings.Contains(erro.Erro, "Croissant") {
		t.Errorf("a mensagem deveria citar o produto, veio %q", erro.Erro)
	}
}

func TestCriarVenda(t *testing.T) {
	servidor := servidorDeTeste(t)

	corpo := `{"cliente":"Maria","produto":"Pão de queijo","quantidade":3}`
	gravador := chamar(t, servidor, http.MethodPost, "/vendas", corpo)

	if gravador.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado 201; corpo: %s", gravador.Code, gravador.Body.String())
	}

	venda := decodificar[RespostaDeVenda](t, gravador)
	if venda.Cliente != "Maria" {
		t.Errorf("Cliente = %q, esperado Maria", venda.Cliente)
	}
	if venda.TotalEmCentavos != 1350 {
		t.Errorf("TotalEmCentavos = %d, esperado 1350", venda.TotalEmCentavos)
	}
	if !venda.Em.Equal(momentoDeTeste) {
		t.Errorf("Em = %v, esperado %v: o relógio injetado não foi usado", venda.Em, momentoDeTeste)
	}

	if restante := servidor.estoque.Quantidade("Pão de queijo"); restante != 7 {
		t.Errorf("sobrou %d no estoque, esperado 7", restante)
	}
}

func TestCriarVendaSemClienteUsaOBalcao(t *testing.T) {
	corpo := `{"produto":"Broa","quantidade":1}`
	gravador := chamar(t, servidorDeTeste(t), http.MethodPost, "/vendas", corpo)

	if gravador.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado 201; corpo: %s", gravador.Code, gravador.Body.String())
	}

	venda := decodificar[RespostaDeVenda](t, gravador)
	if venda.Cliente != "balcão" {
		t.Errorf("Cliente = %q, esperado balcão", venda.Cliente)
	}
}

func TestCriarVendaRecusa(t *testing.T) {
	casos := []struct {
		nome           string
		corpo          string
		statusEsperado int
	}{
		{
			nome:           "corpo que não é JSON",
			corpo:          `isto não é json`,
			statusEsperado: http.StatusBadRequest,
		},
		{
			nome:           "produto fora do cardápio",
			corpo:          `{"produto":"Croissant","quantidade":1}`,
			statusEsperado: http.StatusNotFound,
		},
		{
			nome:           "mais do que tem em estoque",
			corpo:          `{"produto":"Broa","quantidade":99}`,
			statusEsperado: http.StatusConflict,
		},
		{
			nome:           "quantidade zero",
			corpo:          `{"produto":"Broa","quantidade":0}`,
			statusEsperado: http.StatusBadRequest,
		},
		{
			nome:           "quantidade negativa",
			corpo:          `{"produto":"Broa","quantidade":-5}`,
			statusEsperado: http.StatusBadRequest,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			servidor := servidorDeTeste(t)
			antes := servidor.estoque.Quantidade("Broa")

			gravador := chamar(t, servidor, http.MethodPost, "/vendas", caso.corpo)

			if gravador.Code != caso.statusEsperado {
				t.Errorf("status = %d, esperado %d; corpo: %s",
					gravador.Code, caso.statusEsperado, gravador.Body.String())
			}
			if depois := servidor.estoque.Quantidade("Broa"); depois != antes {
				t.Errorf("o estoque foi de %d para %d numa venda recusada", antes, depois)
			}
		})
	}
}

func TestStatusParaErro(t *testing.T) {
	casos := []struct {
		nome           string
		err            error
		statusEsperado int
	}{
		{"produto não encontrado", ErrProdutoNaoEncontrado, http.StatusNotFound},
		{"estoque insuficiente", ErrEstoqueInsuficiente, http.StatusConflict},
		{"quantidade inválida", ErrQuantidadeInvalida, http.StatusBadRequest},
		{"pedido inválido", ErrPedidoInvalido, http.StatusBadRequest},
		{"erro desconhecido", errors.New("o forno pegou fogo"), http.StatusInternalServerError},
		{
			// O erro chega embrulhado em camadas, como você viu no módulo 03.
			// É por isso que a tradução usa errors.Is e não comparação direta.
			"sentinela embrulhado duas vezes",
			fmt.Errorf("vender: %w", fmt.Errorf("baixar: %w", ErrEstoqueInsuficiente)),
			http.StatusConflict,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if recebido := StatusParaErro(caso.err); recebido != caso.statusEsperado {
				t.Errorf("StatusParaErro(%v) = %d, esperado %d", caso.err, recebido, caso.statusEsperado)
			}
		})
	}
}

func TestMetodoErradoDevolve405(t *testing.T) {
	// O mux do Go 1.22 responde 405 sozinho quando o caminho casa com algum
	// padrão mas o método não. Antes disso, isso era trabalho seu.
	gravador := chamar(t, servidorDeTeste(t), http.MethodPost, "/produtos", "")

	if gravador.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, esperado 405", gravador.Code)
	}
}

func TestRotaInexistenteDevolve404(t *testing.T) {
	gravador := chamar(t, servidorDeTeste(t), http.MethodGet, "/bolos", "")

	if gravador.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperado 404", gravador.Code)
	}
}

func TestComRegistro(t *testing.T) {
	var registrador RegistradorEmMemoria
	comRegistro := ComRegistro(servidorDeTeste(t), &registrador)

	chamar(t, comRegistro, http.MethodGet, "/produtos", "")
	chamar(t, comRegistro, http.MethodGet, "/produtos/Croissant", "")

	linhas := registrador.Linhas()
	if len(linhas) != 2 {
		t.Fatalf("registrou %d linhas, esperado 2", len(linhas))
	}
	if linhas[0] != "GET /produtos -> 200" {
		t.Errorf("linha 1 = %q, esperado %q", linhas[0], "GET /produtos -> 200")
	}
	if linhas[1] != "GET /produtos/Croissant -> 404" {
		t.Errorf("linha 2 = %q, esperado %q", linhas[1], "GET /produtos/Croissant -> 404")
	}
}

func TestComRegistroNaoAtrapalhaAResposta(t *testing.T) {
	var registrador RegistradorEmMemoria
	comRegistro := ComRegistro(servidorDeTeste(t), &registrador)

	gravador := chamar(t, comRegistro, http.MethodGet, "/produtos", "")

	if gravador.Code != http.StatusOK {
		t.Errorf("status = %d, esperado 200", gravador.Code)
	}
	if produtos := decodificar[[]Produto](t, gravador); len(produtos) != 4 {
		t.Errorf("vieram %d produtos, esperado 4", len(produtos))
	}
}

func TestDemonstracaoCabecalhoDepoisDoWriteHeaderNaoFazNada(t *testing.T) {
	// Este teste já passa. Ele reproduz, isolado, a pegadinha número um do
	// net/http.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("X-Tarde-Demais", "sim") // sem erro, sem aviso, sem efeito
		fmt.Fprint(w, "corpo")
	})

	gravador := httptest.NewRecorder()
	handler.ServeHTTP(gravador, httptest.NewRequest(http.MethodGet, "/", nil))

	// Repare na diferença entre os dois. O mapa vivo aceitou a escrita; o
	// que teria ido para o cliente, não.
	if gravador.Header().Get("X-Tarde-Demais") != "sim" {
		t.Fatal("o mapa vivo deveria ter aceitado a escrita")
	}
	if gravador.Result().Header.Get("X-Tarde-Demais") != "" {
		t.Fatal("era para o cabeçalho não ter saído na resposta")
	}

	t.Log("o cabeçalho entrou no mapa vivo e mesmo assim não foi enviado: WriteHeader já tinha congelado a resposta")
}

func TestDemonstracaoServidorEhHandler(t *testing.T) {
	// Este teste já passa depois que NovoServidor existir. Ele mostra que o
	// servidor inteiro é apenas um http.Handler, e por isso cabe dentro de um
	// middleware sem nenhuma adaptação.
	var handler http.Handler = servidorDeTeste(t)

	var registrador RegistradorEmMemoria
	handler = ComRegistro(handler, &registrador)
	handler = ComRegistro(handler, &registrador)

	gravador := httptest.NewRecorder()
	handler.ServeHTTP(gravador, httptest.NewRequest(http.MethodGet, "/saude", nil))

	if gravador.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200", gravador.Code)
	}
	if len(registrador.Linhas()) != 2 {
		t.Fatalf("duas camadas de middleware deveriam registrar 2 linhas, vieram %d",
			len(registrador.Linhas()))
	}

	t.Log("middleware empilha porque as duas pontas são a mesma interface de um método")
}

func TestDemonstracaoBytesBufferComoCorpo(t *testing.T) {
	// Este teste já passa. Qualquer io.Reader serve de corpo de requisição,
	// exatamente como no módulo 07.
	corpo := RespostaDeVenda{Cliente: "Maria", Quantidade: 1}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(corpo); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	requisicao := httptest.NewRequest(http.MethodPost, "/vendas", &buffer)
	if requisicao.Body == nil {
		t.Fatal("esperava um corpo")
	}

	t.Log("bytes.Buffer serve de corpo de requisição sem nenhuma conversão")
}
