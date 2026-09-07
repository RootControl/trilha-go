package padaria

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func repositorioDeTeste(t *testing.T) *RepositorioEmMemoria {
	t.Helper()

	repositorio := NovoRepositorioEmMemoria()
	if repositorio == nil {
		t.Fatal("NovoRepositorioEmMemoria devolveu nil")
	}

	produtos := []Produto{
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
	}
	for _, produto := range produtos {
		if err := repositorio.Salvar(produto); err != nil {
			t.Fatalf("erro inesperado ao salvar %q: %v", produto.Nome, err)
		}
	}

	return repositorio
}

func estoqueDeTeste() Estoque {
	estoque := NovoEstoque()
	estoque.Repor("Pão de queijo", 10)
	estoque.Repor("Pão francês", 50)
	return estoque
}

func TestRepositorioSalvarEBuscar(t *testing.T) {
	repositorio := repositorioDeTeste(t)

	produto, err := repositorio.Buscar("Pão de queijo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if produto.PrecoEmCentavos != 450 {
		t.Errorf("preço = %d, esperado 450", produto.PrecoEmCentavos)
	}

	// Salvar de novo com o mesmo nome sobrescreve.
	if err := repositorio.Salvar(Produto{Nome: "Pão de queijo", PrecoEmCentavos: 500, Disponivel: true}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	produto, _ = repositorio.Buscar("Pão de queijo")
	if produto.PrecoEmCentavos != 500 {
		t.Errorf("depois de sobrescrever, preço = %d, esperado 500", produto.PrecoEmCentavos)
	}
}

func TestRepositorioBuscarNaoEncontra(t *testing.T) {
	repositorio := repositorioDeTeste(t)

	_, err := repositorio.Buscar("Croissant")
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, ErrProdutoNaoEncontrado) {
		t.Errorf("errors.Is não achou ErrProdutoNaoEncontrado em %v", err)
	}
	if !strings.Contains(err.Error(), "Croissant") {
		t.Errorf("a mensagem deveria citar o produto procurado, veio %q", err.Error())
	}
}

func TestRepositorioRecusaProdutoSemNome(t *testing.T) {
	repositorio := repositorioDeTeste(t)

	err := repositorio.Salvar(Produto{PrecoEmCentavos: 100, Disponivel: true})
	if !errors.Is(err, ErrProdutoInvalido) {
		t.Errorf("errors.Is não achou ErrProdutoInvalido em %v", err)
	}
}

func TestRepositorioTodosEmOrdem(t *testing.T) {
	repositorio := repositorioDeTeste(t)

	esperado := []Produto{
		{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
	}

	// Roda várias vezes porque a ordem de um range sobre map muda a cada
	// execução, como você viu no módulo 02.
	for range 20 {
		if recebido := repositorio.Todos(); !slices.Equal(recebido, esperado) {
			t.Fatalf("Todos() = %+v, esperado %+v", recebido, esperado)
		}
	}
}

func TestRepositorioVazio(t *testing.T) {
	repositorio := NovoRepositorioEmMemoria()
	if repositorio == nil {
		t.Fatal("NovoRepositorioEmMemoria devolveu nil")
	}

	if todos := repositorio.Todos(); len(todos) != 0 {
		t.Errorf("repositório novo tem %d produtos, esperado nenhum", len(todos))
	}

	// Se o map interno não tiver sido inicializado, esta linha entra em
	// pânico em vez de falhar bonito.
	if err := repositorio.Salvar(Produto{Nome: "Broa", PrecoEmCentavos: 300}); err != nil {
		t.Fatalf("erro inesperado ao salvar no repositório novo: %v", err)
	}
}

func TestErroDeEstoqueEhUmError(t *testing.T) {
	// Escrever Error() string foi tudo que este tipo precisou para virar um
	// error. Não existe "implements" em lugar nenhum do código.
	var err error = &ErroDeEstoque{Produto: "Pão de queijo", Pedido: 99, Disponivel: 10}

	mensagem := err.Error()
	if mensagem == "" {
		t.Fatal("Error() devolveu string vazia")
	}

	// A mensagem é conferida por pedaços, e não por igualdade exata, pela
	// mesma razão do módulo 03: texto de erro não é contrato.
	for _, trecho := range []string{"Pão de queijo", strconv.Itoa(99), strconv.Itoa(10)} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("a mensagem deveria citar %q, veio %q", trecho, mensagem)
		}
	}
}

func TestErroDeEstoqueContinuaSendoAchadoPorErrorsIs(t *testing.T) {
	// Este é o pagamento da dívida do módulo 03. Quem só quer saber se faltou
	// estoque continua escrevendo o mesmo errors.Is de antes, sem saber que
	// agora existe um tipo por trás. Quem quiser os números usa errors.As.
	estoque := estoqueDeTeste()

	err := estoque.Baixar("Pão de queijo", 99)
	if err == nil {
		t.Fatal("esperava erro")
	}

	if !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Error("errors.Is não achou ErrEstoqueInsuficiente: falta o método Unwrap")
	}

	var erroDeEstoque *ErroDeEstoque
	if !errors.As(err, &erroDeEstoque) {
		t.Fatal("errors.As não conseguiu extrair *ErroDeEstoque")
	}
	if erroDeEstoque.Pedido != 99 {
		t.Errorf("Pedido = %d, esperado 99", erroDeEstoque.Pedido)
	}
	if erroDeEstoque.Disponivel != 10 {
		t.Errorf("Disponivel = %d, esperado 10", erroDeEstoque.Disponivel)
	}
}

func TestVenderComORepositorioReal(t *testing.T) {
	repositorio := repositorioDeTeste(t)
	estoque := estoqueDeTeste()

	total, err := Vender(repositorio, estoque, "Pão de queijo", 3)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if total != 1350 {
		t.Errorf("total = %d, esperado 1350", total)
	}
	if restante := estoque.Quantidade("Pão de queijo"); restante != 7 {
		t.Errorf("sobrou %d, esperado 7", restante)
	}
}

func TestVenderPropagaOsErros(t *testing.T) {
	casos := []struct {
		nome              string
		produto           string
		quantidade        int
		sentinelaEsperada error
	}{
		{"produto fora do repositório", "Croissant", 1, ErrProdutoNaoEncontrado},
		{"estoque insuficiente", "Pão de queijo", 99, ErrEstoqueInsuficiente},
		{"quantidade inválida", "Pão de queijo", 0, ErrQuantidadeInvalida},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			total, err := Vender(repositorioDeTeste(t), estoqueDeTeste(), caso.produto, caso.quantidade)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}
			if total != 0 {
				t.Errorf("com erro, o total deveria ser 0, veio %d", total)
			}
		})
	}
}

func TestVenderDeixaOsNumerosDoErroPassarem(t *testing.T) {
	// O erro nasce dentro de Baixar como *ErroDeEstoque, é envolvido por
	// Vender com fmt.Errorf, e mesmo assim errors.As alcança o tipo lá dentro.
	// É isso que permite a padaria responder "não tenho 99, tenho 10".
	_, err := Vender(repositorioDeTeste(t), estoqueDeTeste(), "Pão de queijo", 99)

	var erroDeEstoque *ErroDeEstoque
	if !errors.As(err, &erroDeEstoque) {
		t.Fatalf("errors.As não achou *ErroDeEstoque em %v; confira se Vender usou %%w", err)
	}
	if erroDeEstoque.Disponivel != 10 {
		t.Errorf("Disponivel = %d, esperado 10", erroDeEstoque.Disponivel)
	}
}

// ----------------------------------------------------------------------
// Dublês de teste: tipos que satisfazem BuscadorDeProdutos sem serem
// repositório nenhum.
// ----------------------------------------------------------------------

// buscadorFalso guarda o que devolver e conta quantas vezes foi chamado.
// Ele não sabe que BuscadorDeProdutos existe, e mesmo assim serve.
type buscadorFalso struct {
	produtos map[string]Produto
	chamadas int
}

func (b *buscadorFalso) Buscar(nome string) (Produto, error) {
	b.chamadas++

	produto, existe := b.produtos[nome]
	if !existe {
		return Produto{}, ErrProdutoNaoEncontrado
	}

	return produto, nil
}

// buscadorQuebrado devolve sempre o mesmo erro. Escrever um repositório que
// falha de propósito seria trabalhoso; escrever isto custa quatro linhas.
type buscadorQuebrado struct {
	err error
}

func (b buscadorQuebrado) Buscar(string) (Produto, error) {
	return Produto{}, b.err
}

func TestVenderFuncionaComUmDubleDeTeste(t *testing.T) {
	falso := &buscadorFalso{
		produtos: map[string]Produto{
			"Pão de queijo": {Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
		},
	}
	estoque := estoqueDeTeste()

	total, err := Vender(falso, estoque, "Pão de queijo", 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if total != 900 {
		t.Errorf("total = %d, esperado 900", total)
	}

	// De brinde, o dublê consegue responder uma pergunta que o repositório
	// real não responderia: Vender chamou Buscar quantas vezes?
	if falso.chamadas != 1 {
		t.Errorf("Buscar foi chamado %d vezes, esperado 1", falso.chamadas)
	}
}

func TestVenderPropagaErroDoBuscador(t *testing.T) {
	erroDoBanco := errors.New("conexão recusada")

	_, err := Vender(buscadorQuebrado{err: erroDoBanco}, estoqueDeTeste(), "Pão de queijo", 1)

	if !errors.Is(err, erroDoBanco) {
		t.Errorf("errors.Is não achou o erro do buscador em %v", err)
	}
}

func TestDemonstracaoPonteiroNilDentroDeInterfaceNaoEhNil(t *testing.T) {
	// Este teste já passa. Ele reproduz a armadilha mais famosa de interface
	// em Go, e ela custa tempo de gente experiente.
	var erroConcreto *ErroDeEstoque // é nil

	if erroConcreto != nil {
		t.Fatal("o ponteiro deveria ser nil")
	}

	// Guardar esse ponteiro nil numa interface produz uma interface que NÃO é
	// nil, porque a interface carrega duas coisas: o tipo e o valor. O valor
	// é nil, mas o tipo é *ErroDeEstoque, e isso já basta para a comparação
	// com nil dar false.
	var err error = erroConcreto

	if err == nil {
		t.Fatal("era para a interface não ser nil, mesmo carregando um ponteiro nil")
	}

	t.Log("o ponteiro é nil, a interface que o carrega não é: por isso funções devolvem error, e nunca *MeuErro")
}
