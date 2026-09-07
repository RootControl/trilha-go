package padaria

import "testing"

func TestNovoProduto(t *testing.T) {
	produto := NovoProduto("Pão de queijo", 450)

	if produto.Nome != "Pão de queijo" {
		t.Errorf("Nome = %q, esperado %q", produto.Nome, "Pão de queijo")
	}
	if produto.PrecoEmCentavos != 450 {
		t.Errorf("PrecoEmCentavos = %d, esperado %d", produto.PrecoEmCentavos, 450)
	}
	if !produto.Disponivel {
		t.Error("produto recém-criado deveria estar disponível")
	}
}

func TestValorZeroDeProdutoEhUsavel(t *testing.T) {
	// Este teste existe para provar um ponto: em Go você pode declarar uma
	// struct sem inicializar nada e ela já está pronta para uso. Não existe
	// null, não existe undefined, não existe None. Cada campo recebe o valor
	// zero do seu tipo.
	var produto Produto

	if produto.Nome != "" {
		t.Errorf("o valor zero de string deveria ser vazio, veio %q", produto.Nome)
	}
	if produto.PrecoEmCentavos != 0 {
		t.Errorf("o valor zero de int deveria ser 0, veio %d", produto.PrecoEmCentavos)
	}
	if produto.Disponivel {
		t.Error("o valor zero de bool deveria ser false")
	}

	// E, mais importante: dá para passar esse produto adiante sem estourar.
	if recebido := Descrever(produto); recebido != "produto sem nome: esgotado" {
		t.Errorf("Descrever(Produto{}) = %q, esperado %q", recebido, "produto sem nome: esgotado")
	}
}

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
			nome:     "produto esgotado não mostra preço",
			produto:  Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: false},
			esperado: "Pão de queijo: esgotado",
		},
		{
			nome:     "produto de graça ainda mostra preço",
			produto:  Produto{Nome: "Copo d'água", PrecoEmCentavos: 0, Disponivel: true},
			esperado: "Copo d'água: R$ 0,00",
		},
		{
			nome:     "produto sem nome",
			produto:  Produto{Nome: "", PrecoEmCentavos: 300, Disponivel: true},
			esperado: "produto sem nome: R$ 3,00",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := Descrever(caso.produto)
			if recebido != caso.esperado {
				t.Errorf("Descrever(%+v) = %q, esperado %q", caso.produto, recebido, caso.esperado)
			}
		})
	}
}

func TestAplicarDesconto(t *testing.T) {
	casos := []struct {
		nome          string
		preco         int
		percentual    int
		precoEsperado int
	}{
		{nome: "dez por cento", preco: 450, percentual: 10, precoEsperado: 405},
		{nome: "metade do preço", preco: 1000, percentual: 50, precoEsperado: 500},
		{nome: "sem desconto", preco: 450, percentual: 0, precoEsperado: 450},
		{nome: "de graça", preco: 450, percentual: 100, precoEsperado: 0},
		{
			// 333 * 10 / 100 dá 33,3, e a divisão de inteiro corta o 0,3.
			// O desconto vira 33 centavos, não 34. Leia o README.
			nome:          "divisão de inteiro corta a parte quebrada",
			preco:         333,
			percentual:    10,
			precoEsperado: 300,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			original := Produto{Nome: "Pão de queijo", PrecoEmCentavos: caso.preco, Disponivel: true}

			comDesconto := AplicarDesconto(original, caso.percentual)

			if comDesconto.PrecoEmCentavos != caso.precoEsperado {
				t.Errorf("preço com desconto = %d, esperado %d", comDesconto.PrecoEmCentavos, caso.precoEsperado)
			}
			if comDesconto.Nome != original.Nome {
				t.Errorf("o desconto não deveria mexer no nome: %q", comDesconto.Nome)
			}
			if !comDesconto.Disponivel {
				t.Error("o desconto não deveria tirar o produto de venda")
			}
		})
	}
}

func TestAplicarDescontoNaoMexeNoOriginal(t *testing.T) {
	// Se você veio de JavaScript, Python ou PHP, este é o teste que mais
	// importa no módulo. Lá, passar um objeto para uma função entrega uma
	// referência, e quem recebeu consegue mexer no seu dado. Em Go, o
	// argumento é uma cópia.
	original := Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}

	AplicarDesconto(original, 50)

	if original.PrecoEmCentavos != 450 {
		t.Errorf("o produto original mudou para %d, deveria continuar 450", original.PrecoEmCentavos)
	}
}

func TestTotalEmCentavos(t *testing.T) {
	casos := []struct {
		nome     string
		pedido   Pedido
		esperado int
	}{
		{
			nome: "três pães de queijo",
			pedido: Pedido{
				Cliente:    "Maria",
				Produto:    Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
				Quantidade: 3,
			},
			esperado: 1350,
		},
		{
			nome: "pedido sem quantidade não custa nada",
			pedido: Pedido{
				Cliente:    "Maria",
				Produto:    Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
				Quantidade: 0,
			},
			esperado: 0,
		},
		{
			nome:     "o valor zero de Pedido também precisa funcionar",
			pedido:   Pedido{},
			esperado: 0,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := TotalEmCentavos(caso.pedido)
			if recebido != caso.esperado {
				t.Errorf("TotalEmCentavos(%+v) = %d, esperado %d", caso.pedido, recebido, caso.esperado)
			}
		})
	}
}

func TestProdutosSaoComparaveis(t *testing.T) {
	// Bônus de graça: struct em Go compara com == quando todos os campos são
	// comparáveis. Você não precisa escrever equals, nem __eq__, nem
	// deepEqual para um caso simples como este.
	a := NovoProduto("Pão de queijo", 450)
	b := NovoProduto("Pão de queijo", 450)

	if a != b {
		t.Errorf("dois produtos iguais deveriam ser ==, veio %+v e %+v", a, b)
	}
}
