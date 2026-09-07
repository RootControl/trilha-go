package padaria

import (
	"slices"
	"testing"
)

func TestDisponiveis(t *testing.T) {
	casos := []struct {
		nome     string
		cardapio []Produto
		esperado []Produto
	}{
		{
			nome: "filtra o que está esgotado",
			cardapio: []Produto{
				{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
				{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
				{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
			},
			esperado: []Produto{
				{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
				{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
			},
		},
		{
			nome: "tudo esgotado devolve vazio",
			cardapio: []Produto{
				{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
			},
			esperado: nil,
		},
		{
			nome:     "cardápio vazio",
			cardapio: []Produto{},
			esperado: nil,
		},
		{
			// Uma fatia nil não é um caso especial em Go. Ela tem len 0, e o
			// laço range simplesmente não roda nenhuma vez.
			nome:     "cardápio nil não explode",
			cardapio: nil,
			esperado: nil,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := Disponiveis(caso.cardapio)

			// slices.Equal considera fatia nil e fatia vazia como iguais,
			// porque as duas têm len 0. É o jeito certo de comparar aqui.
			if !slices.Equal(recebido, caso.esperado) {
				t.Errorf("Disponiveis() = %+v, esperado %+v", recebido, caso.esperado)
			}
		})
	}
}

func TestDisponiveisNaoMexeNoCardapioOriginal(t *testing.T) {
	cardapio := []Produto{
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
	}

	Disponiveis(cardapio)

	if len(cardapio) != 2 {
		t.Fatalf("o cardápio original ficou com %d itens, deveria continuar com 2", len(cardapio))
	}
	if cardapio[1].Nome != "Sonho" {
		t.Errorf("o cardápio original foi alterado: %+v", cardapio)
	}
}

func TestMaisBarato(t *testing.T) {
	casos := []struct {
		nome            string
		cardapio        []Produto
		esperado        Produto
		esperaEncontrar bool
	}{
		{
			nome: "acha o mais barato",
			cardapio: []Produto{
				{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: true},
				{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
				{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
			},
			esperado:        Produto{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
			esperaEncontrar: true,
		},
		{
			nome: "empate fica com o primeiro",
			cardapio: []Produto{
				{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
				{Nome: "Broa", PrecoEmCentavos: 100, Disponivel: true},
			},
			esperado:        Produto{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
			esperaEncontrar: true,
		},
		{
			nome:            "cardápio vazio devolve o valor zero e false",
			cardapio:        nil,
			esperado:        Produto{},
			esperaEncontrar: false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido, encontrou := MaisBarato(caso.cardapio)

			if encontrou != caso.esperaEncontrar {
				t.Fatalf("encontrou = %v, esperado %v", encontrou, caso.esperaEncontrar)
			}
			if recebido != caso.esperado {
				t.Errorf("MaisBarato() = %+v, esperado %+v", recebido, caso.esperado)
			}
		})
	}
}

func TestReajustar(t *testing.T) {
	cardapio := []Produto{
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
	}

	Reajustar(cardapio, 10)

	if cardapio[0].PrecoEmCentavos != 110 {
		t.Errorf("preço do pão francês = %d, esperado 110", cardapio[0].PrecoEmCentavos)
	}
	if cardapio[1].PrecoEmCentavos != 495 {
		t.Errorf("preço do pão de queijo = %d, esperado 495", cardapio[1].PrecoEmCentavos)
	}
}

func TestReajustarAlteraOCardapioDeQuemChamou(t *testing.T) {
	// Este é o teste que faz par com TestAplicarDescontoNaoMexeNoOriginal, do
	// módulo 01. Lá, passar uma struct protegeu o original. Aqui, passar uma
	// fatia não protege nada: o cabeçalho da fatia é copiado, mas o array por
	// baixo dele é o mesmo. Quem recebe uma fatia consegue mexer no seu dado.
	cardapio := []Produto{
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
	}

	Reajustar(cardapio, 100)

	if cardapio[0].PrecoEmCentavos == 100 {
		t.Error("o preço não mudou: Reajustar precisa alterar o cardápio no lugar")
	}
	if cardapio[0].PrecoEmCentavos != 200 {
		t.Errorf("preço = %d, esperado 200", cardapio[0].PrecoEmCentavos)
	}
}

func TestNovoEstoqueAceitaEscrita(t *testing.T) {
	estoque := NovoEstoque()

	if estoque == nil {
		t.Fatal("NovoEstoque devolveu nil, e escrever num map nil derruba o programa")
	}

	// Se NovoEstoque tivesse devolvido nil, a linha abaixo entraria em pânico
	// em vez de falhar bonito.
	estoque["Pão francês"] = 10

	if estoque["Pão francês"] != 10 {
		t.Errorf("quantidade = %d, esperado 10", estoque["Pão francês"])
	}
}

func TestRepor(t *testing.T) {
	estoque := NovoEstoque()

	Repor(estoque, "Pão francês", 10)
	if QuantidadeEm(estoque, "Pão francês") != 10 {
		t.Errorf("depois da primeira reposição = %d, esperado 10", QuantidadeEm(estoque, "Pão francês"))
	}

	Repor(estoque, "Pão francês", 5)
	if QuantidadeEm(estoque, "Pão francês") != 15 {
		t.Errorf("depois da segunda reposição = %d, esperado 15", QuantidadeEm(estoque, "Pão francês"))
	}

	Repor(estoque, "Sonho", 3)
	if QuantidadeEm(estoque, "Sonho") != 3 {
		t.Errorf("produto novo = %d, esperado 3", QuantidadeEm(estoque, "Sonho"))
	}
}

func TestReporAlteraOEstoqueDeQuemChamou(t *testing.T) {
	// Map, assim como fatia, não protege quem chamou. Repor não devolve nada
	// e mesmo assim o estoque daqui muda.
	estoque := NovoEstoque()

	Repor(estoque, "Pão francês", 10)

	if len(estoque) != 1 {
		t.Fatalf("o estoque ficou com %d chaves, esperado 1", len(estoque))
	}
}

func TestQuantidadeEm(t *testing.T) {
	estoque := Estoque{"Pão francês": 10}

	casos := []struct {
		nome     string
		estoque  Estoque
		produto  string
		esperado int
	}{
		{nome: "produto que existe", estoque: estoque, produto: "Pão francês", esperado: 10},
		{nome: "produto que não existe devolve zero", estoque: estoque, produto: "Sonho", esperado: 0},
		{nome: "estoque nil devolve zero sem explodir", estoque: nil, produto: "Sonho", esperado: 0},
		{nome: "estoque vazio devolve zero", estoque: NovoEstoque(), produto: "Sonho", esperado: 0},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := QuantidadeEm(caso.estoque, caso.produto)
			if recebido != caso.esperado {
				t.Errorf("QuantidadeEm(%v, %q) = %d, esperado %d", caso.estoque, caso.produto, recebido, caso.esperado)
			}
		})
	}
}

func TestNomesEmOrdem(t *testing.T) {
	casos := []struct {
		nome     string
		estoque  Estoque
		esperado []string
	}{
		{
			nome:     "ordem alfabética",
			estoque:  Estoque{"Sonho": 3, "Broa": 1, "Pão francês": 10},
			esperado: []string{"Broa", "Pão francês", "Sonho"},
		},
		{
			nome:     "estoque vazio",
			estoque:  NovoEstoque(),
			esperado: nil,
		},
		{
			nome:     "estoque nil",
			estoque:  nil,
			esperado: nil,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := NomesEmOrdem(caso.estoque)
			if !slices.Equal(recebido, caso.esperado) {
				t.Errorf("NomesEmOrdem() = %v, esperado %v", recebido, caso.esperado)
			}
		})
	}
}

func TestNomesEmOrdemEhEstavel(t *testing.T) {
	// A ordem de um range sobre map é aleatória de propósito e muda a cada
	// execução. Rodar cem vezes e exigir o mesmo resultado prova que você
	// ordenou de verdade, em vez de ter tido sorte uma vez.
	estoque := Estoque{"Sonho": 3, "Broa": 1, "Pão francês": 10, "Bolo": 2, "Café": 5}
	esperado := []string{"Bolo", "Broa", "Café", "Pão francês", "Sonho"}

	for range 100 {
		if recebido := NomesEmOrdem(estoque); !slices.Equal(recebido, esperado) {
			t.Fatalf("NomesEmOrdem() = %v, esperado %v", recebido, esperado)
		}
	}
}

func TestDemonstracaoFatiaCompartilhaMemoria(t *testing.T) {
	// Este teste já passa. Ele não exercita o seu código: está aqui para você
	// ver acontecer a pegadinha mais famosa de Go, com os próprios olhos.
	cardapio := []Produto{
		{Nome: "Pão francês"},
		{Nome: "Pão de queijo"},
		{Nome: "Sonho"},
	}

	// promocao enxerga os dois primeiros, mas continua carregando o array
	// inteiro por baixo: len 2, cap 3.
	promocao := cardapio[:2]
	if len(promocao) != 2 || cap(promocao) != 3 {
		t.Fatalf("len=%d cap=%d, esperado len=2 cap=3", len(promocao), cap(promocao))
	}

	// Como ainda cabe dentro da capacidade, o append NÃO cria um array novo.
	// Ele escreve na posição 2 do array que o cardápio também usa.
	promocao = append(promocao, Produto{Nome: "Broa"})

	if cardapio[2].Nome != "Broa" {
		t.Fatalf("o Sonho deveria ter sido sobrescrito, veio %q", cardapio[2].Nome)
	}
	t.Logf("o append em promocao apagou o Sonho do cardápio: %q", cardapio[2].Nome)
}
