package padaria

import (
	"errors"
	"strings"
	"testing"
)

func cardapioDeTeste() []Produto {
	return []Produto{
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
	}
}

func estoqueDeTeste() Estoque {
	estoque := NovoEstoque()
	Repor(estoque, "Pão francês", 50)
	Repor(estoque, "Pão de queijo", 10)
	Repor(estoque, "Sonho", 0) // existe no estoque, mas zerado
	return estoque
}

func TestSentinelasExistemESaoDistintos(t *testing.T) {
	sentinelas := map[string]error{
		"ErrProdutoNaoEncontrado": ErrProdutoNaoEncontrado,
		"ErrEstoqueInsuficiente":  ErrEstoqueInsuficiente,
		"ErrQuantidadeInvalida":   ErrQuantidadeInvalida,
	}

	for nome, err := range sentinelas {
		if err == nil {
			t.Errorf("%s ainda é nil: crie o erro com errors.New", nome)
			continue
		}
		if err.Error() == "" {
			t.Errorf("%s não tem mensagem", nome)
		}
	}

	// Dois erros criados com errors.New nunca são iguais, mesmo que a
	// mensagem seja idêntica. Cada chamada devolve um valor novo.
	if errors.Is(ErrEstoqueInsuficiente, ErrProdutoNaoEncontrado) {
		t.Error("os sentinelas precisam ser valores distintos")
	}
}

func TestBuscarProdutoEncontra(t *testing.T) {
	produto, err := BuscarProduto(cardapioDeTeste(), "Pão de queijo")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if produto.PrecoEmCentavos != 450 {
		t.Errorf("preço = %d, esperado 450", produto.PrecoEmCentavos)
	}
}

func TestBuscarProdutoNaoEncontra(t *testing.T) {
	produto, err := BuscarProduto(cardapioDeTeste(), "Croissant")

	if err == nil {
		t.Fatal("esperava erro para produto que não está no cardápio")
	}
	if !errors.Is(err, ErrProdutoNaoEncontrado) {
		t.Errorf("errors.Is não achou ErrProdutoNaoEncontrado em %v", err)
	}

	// O erro precisa dizer QUAL produto faltou. É para isso que serve
	// envolver com fmt.Errorf em vez de devolver o sentinela pelado.
	if !strings.Contains(err.Error(), "Croissant") {
		t.Errorf("a mensagem deveria citar o produto procurado, veio %q", err.Error())
	}

	// Quando devolve erro, o outro retorno fica no valor zero.
	if produto != (Produto{}) {
		t.Errorf("com erro, o produto devolvido deveria ser o valor zero, veio %+v", produto)
	}
}

func TestBaixarComSucesso(t *testing.T) {
	estoque := estoqueDeTeste()

	if err := Baixar(estoque, "Pão de queijo", 4); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if restante := QuantidadeEm(estoque, "Pão de queijo"); restante != 6 {
		t.Errorf("sobrou %d no estoque, esperado 6", restante)
	}
}

func TestBaixarRecusa(t *testing.T) {
	casos := []struct {
		nome              string
		produto           string
		quantidade        int
		sentinelaEsperada error
	}{
		{
			nome:              "quantidade zero",
			produto:           "Pão de queijo",
			quantidade:        0,
			sentinelaEsperada: ErrQuantidadeInvalida,
		},
		{
			nome:              "quantidade negativa",
			produto:           "Pão de queijo",
			quantidade:        -3,
			sentinelaEsperada: ErrQuantidadeInvalida,
		},
		{
			nome:              "produto que nunca entrou no estoque",
			produto:           "Croissant",
			quantidade:        1,
			sentinelaEsperada: ErrProdutoNaoEncontrado,
		},
		{
			// Sonho está no estoque, zerado. Isso é diferente de não estar
			// no estoque, e o erro precisa ser diferente também.
			nome:              "produto no estoque, mas zerado",
			produto:           "Sonho",
			quantidade:        1,
			sentinelaEsperada: ErrEstoqueInsuficiente,
		},
		{
			nome:              "quer mais do que tem",
			produto:           "Pão de queijo",
			quantidade:        11,
			sentinelaEsperada: ErrEstoqueInsuficiente,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			estoque := estoqueDeTeste()
			antes := QuantidadeEm(estoque, caso.produto)

			err := Baixar(estoque, caso.produto, caso.quantidade)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}

			// Falhou, não mexeu. Uma operação que dá erro no meio não pode
			// deixar o estoque num estado esquisito.
			if depois := QuantidadeEm(estoque, caso.produto); depois != antes {
				t.Errorf("o estoque mudou de %d para %d mesmo com erro", antes, depois)
			}
		})
	}
}

func TestVenderComSucesso(t *testing.T) {
	estoque := estoqueDeTeste()

	total, err := Vender(cardapioDeTeste(), estoque, "Pão de queijo", 3)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if total != 1350 {
		t.Errorf("total = %d, esperado 1350", total)
	}
	if restante := QuantidadeEm(estoque, "Pão de queijo"); restante != 7 {
		t.Errorf("sobrou %d no estoque, esperado 7", restante)
	}
}

func TestVenderPropagaOSentinelaDeDuasCamadasAbaixo(t *testing.T) {
	// Este é o teste que justifica o módulo inteiro. O erro nasce dentro de
	// Baixar, é envolvido lá, é envolvido de novo dentro de Vender, e mesmo
	// assim quem chamou consegue perguntar "isso foi falta de estoque?" e
	// receber sim. O contexto se acumula sem esconder a causa.
	casos := []struct {
		nome              string
		produto           string
		quantidade        int
		sentinelaEsperada error
	}{
		{
			nome:              "produto fora do cardápio",
			produto:           "Croissant",
			quantidade:        1,
			sentinelaEsperada: ErrProdutoNaoEncontrado,
		},
		{
			nome:              "estoque insuficiente",
			produto:           "Pão de queijo",
			quantidade:        99,
			sentinelaEsperada: ErrEstoqueInsuficiente,
		},
		{
			nome:              "quantidade inválida",
			produto:           "Pão de queijo",
			quantidade:        0,
			sentinelaEsperada: ErrQuantidadeInvalida,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			total, err := Vender(cardapioDeTeste(), estoqueDeTeste(), caso.produto, caso.quantidade)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}
			if total != 0 {
				t.Errorf("com erro, o total deveria ser 0, veio %d", total)
			}
			if !strings.Contains(err.Error(), caso.produto) {
				t.Errorf("a mensagem deveria citar %q, veio %q", caso.produto, err.Error())
			}
		})
	}
}

func TestVenderNaoConsomeEstoqueQuandoFalha(t *testing.T) {
	estoque := estoqueDeTeste()

	if _, err := Vender(cardapioDeTeste(), estoque, "Pão de queijo", 99); err == nil {
		t.Fatal("esperava erro")
	}

	if restante := QuantidadeEm(estoque, "Pão de queijo"); restante != 10 {
		t.Errorf("o estoque foi consumido mesmo com a venda falhando: sobrou %d, esperado 10", restante)
	}
}

func TestDemonstracaoIgualdadeDiretaNaoEnxergaErroEmbrulhado(t *testing.T) {
	// Este teste já passa. Ele existe para você ver por que errors.Is
	// precisa existir, em vez de você comparar com ==.
	_, err := Vender(cardapioDeTeste(), estoqueDeTeste(), "Pão de queijo", 99)
	if err == nil {
		t.Skip("implemente Vender para ver esta demonstração")
	}

	if err == ErrEstoqueInsuficiente { //nolint:errorlint // é o ponto do teste
		t.Fatal("não era para a comparação direta funcionar aqui")
	}
	if !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Fatal("errors.Is deveria achar o sentinela")
	}

	t.Logf("o erro final é %q, e == falharia; errors.Is acha o sentinela lá no fundo", err)
}
