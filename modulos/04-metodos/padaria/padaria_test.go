package padaria

import (
	"errors"
	"fmt"
	"testing"
)

func TestDescrever(t *testing.T) {
	casos := []struct {
		nome     string
		produto  Produto
		esperado string
	}{
		{
			nome:     "produto à venda",
			produto:  Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
			esperado: "Pão de queijo: R$ 4,50",
		},
		{
			nome:     "produto esgotado",
			produto:  Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: false},
			esperado: "Pão de queijo: esgotado",
		},
		{
			nome:     "produto sem nome",
			produto:  Produto{Nome: "", PrecoEmCentavos: 300, Disponivel: true},
			esperado: "produto sem nome: R$ 3,00",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			// Agora a chamada é produto.Descrever(), não Descrever(produto).
			if recebido := caso.produto.Descrever(); recebido != caso.esperado {
				t.Errorf("Descrever() = %q, esperado %q", recebido, caso.esperado)
			}
		})
	}
}

func TestReajustarAlteraOProduto(t *testing.T) {
	// Se você deixou o receptor como Produto em vez de *Produto, este teste
	// compila, roda sem erro nenhum e falha aqui. É o ponto do módulo.
	produto := Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}

	produto.Reajustar(10)

	if produto.PrecoEmCentavos != 495 {
		t.Errorf("preço = %d, esperado 495; o receptor precisa ser *Produto",
			produto.PrecoEmCentavos)
	}
}

func TestEsgotarAlteraOProduto(t *testing.T) {
	produto := Produto{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: true}

	produto.Esgotar()

	if produto.Disponivel {
		t.Error("o produto deveria estar indisponível; o receptor precisa ser *Produto")
	}
	if recebido := produto.Descrever(); recebido != "Sonho: esgotado" {
		t.Errorf("Descrever() = %q, esperado %q", recebido, "Sonho: esgotado")
	}
}

func TestMetodoDePonteiroFuncionaEmElementoDeFatia(t *testing.T) {
	// Elemento de fatia é endereçável, então o Go monta o &cardapio[0]
	// sozinho e o método de ponteiro funciona direto. Isso NÃO vale para
	// elemento de map: leia o README.
	cardapio := []Produto{
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
	}

	for i := range cardapio {
		cardapio[i].Reajustar(100)
	}

	if cardapio[0].PrecoEmCentavos != 200 {
		t.Errorf("preço do primeiro = %d, esperado 200", cardapio[0].PrecoEmCentavos)
	}
	if cardapio[1].PrecoEmCentavos != 900 {
		t.Errorf("preço do segundo = %d, esperado 900", cardapio[1].PrecoEmCentavos)
	}
}

func TestEstoqueQuantidade(t *testing.T) {
	estoque := Estoque{"Pão francês": 10}

	casos := []struct {
		nome     string
		estoque  Estoque
		produto  string
		esperado int
	}{
		{nome: "produto que existe", estoque: estoque, produto: "Pão francês", esperado: 10},
		{nome: "produto ausente", estoque: estoque, produto: "Sonho", esperado: 0},
		{nome: "estoque nil não explode", estoque: nil, produto: "Sonho", esperado: 0},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if recebido := caso.estoque.Quantidade(caso.produto); recebido != caso.esperado {
				t.Errorf("Quantidade(%q) = %d, esperado %d", caso.produto, recebido, caso.esperado)
			}
		})
	}
}

func TestEstoqueReporComReceptorDeValorAindaAlteraOMap(t *testing.T) {
	// Este é o contraponto de TestReajustarAlteraOProduto. O receptor de
	// Repor é Estoque, não *Estoque, e mesmo assim o estoque de quem chamou
	// muda. O README explica por que a regra é diferente aqui.
	estoque := NovoEstoque()

	estoque.Repor("Pão francês", 10)
	estoque.Repor("Pão francês", 5)

	if recebido := estoque.Quantidade("Pão francês"); recebido != 15 {
		t.Errorf("quantidade = %d, esperado 15", recebido)
	}
}

func TestEstoqueBaixarContinuaFuncionandoComoMetodo(t *testing.T) {
	estoque := NovoEstoque()
	estoque.Repor("Pão de queijo", 10)

	if err := estoque.Baixar("Pão de queijo", 4); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if recebido := estoque.Quantidade("Pão de queijo"); recebido != 6 {
		t.Errorf("sobrou %d, esperado 6", recebido)
	}

	err := estoque.Baixar("Pão de queijo", 99)
	if !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Errorf("errors.Is não achou ErrEstoqueInsuficiente em %v", err)
	}
}

func TestCaixaRegistrar(t *testing.T) {
	// Mesma armadilha do Reajustar: receptor de valor faz este teste falhar.
	var caixa Caixa

	caixa.Registrar(450)
	caixa.Registrar(900)

	if caixa.Vendas != 2 {
		t.Errorf("vendas = %d, esperado 2; o receptor precisa ser *Caixa", caixa.Vendas)
	}
	if caixa.TotalEmCentavos != 1350 {
		t.Errorf("total = %d, esperado 1350; o receptor precisa ser *Caixa", caixa.TotalEmCentavos)
	}
}

func TestCaixaTicketMedio(t *testing.T) {
	casos := []struct {
		nome     string
		caixa    Caixa
		esperado int
	}{
		{
			nome:     "média exata",
			caixa:    Caixa{TotalEmCentavos: 2700, Vendas: 3},
			esperado: 900,
		},
		{
			nome:     "divisão inteira corta a parte quebrada",
			caixa:    Caixa{TotalEmCentavos: 1000, Vendas: 3},
			esperado: 333,
		},
		{
			// O valor zero de Caixa precisa funcionar, e dividir por zero
			// derrubaria o programa.
			nome:     "caixa sem venda nenhuma",
			caixa:    Caixa{},
			esperado: 0,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if recebido := caso.caixa.TicketMedio(); recebido != caso.esperado {
				t.Errorf("TicketMedio() = %d, esperado %d", recebido, caso.esperado)
			}
		})
	}
}

func TestCaixaString(t *testing.T) {
	caixa := Caixa{TotalEmCentavos: 2700, Vendas: 3}
	esperado := "caixa: 3 vendas, R$ 27,00"

	if recebido := caixa.String(); recebido != esperado {
		t.Errorf("String() = %q, esperado %q", recebido, esperado)
	}

	// O pacote fmt procura um método String sozinho. Você não registrou
	// nada em lugar nenhum, não implementou interface nenhuma de forma
	// explícita, e mesmo assim %v passa a usar o seu texto.
	if recebido := fmt.Sprintf("%v", caixa); recebido != esperado {
		t.Errorf("fmt.Sprintf(%%v) = %q, esperado %q", recebido, esperado)
	}
	if recebido := fmt.Sprintf("%s", caixa); recebido != esperado {
		t.Errorf("fmt.Sprintf(%%s) = %q, esperado %q", recebido, esperado)
	}
}

func TestCaixaVazioString(t *testing.T) {
	var caixa Caixa
	esperado := "caixa: 0 vendas, R$ 0,00"

	if recebido := caixa.String(); recebido != esperado {
		t.Errorf("String() = %q, esperado %q", recebido, esperado)
	}
}

// reajustarComReceptorDeValor é um método deliberadamente errado, definido
// aqui no teste, para você ver a armadilha rodando.
func (p Produto) reajustarComReceptorDeValor(percentual int) {
	p.PrecoEmCentavos += p.PrecoEmCentavos * percentual / 100
}

func TestDemonstracaoReceptorDeValorNaoAlteraNada(t *testing.T) {
	// Este teste já passa. Ele existe para você ver o bug mais silencioso
	// deste módulo com os próprios olhos.
	produto := Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}

	produto.reajustarComReceptorDeValor(100)

	if produto.PrecoEmCentavos != 450 {
		t.Fatalf("preço = %d; era para o método de receptor de valor não ter mudado nada",
			produto.PrecoEmCentavos)
	}

	t.Logf("o método rodou, dobrou o preço da própria cópia, não deu erro nenhum, e o preço aqui continua %d", produto.PrecoEmCentavos)
}
