package padaria

import (
	"slices"
	"testing"
	"time"
)

func cardapioDeTeste() []Produto {
	return []Produto{
		{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
		{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
		{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
		{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
	}
}

func pedidosDeTeste() []Pedido {
	momento := time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)
	return []Pedido{
		{Cliente: "Maria", Produto: Produto{Nome: "Broa", PrecoEmCentavos: 300}, Quantidade: 2, Em: momento},
		{Cliente: "João", Produto: Produto{Nome: "Pão francês", PrecoEmCentavos: 100}, Quantidade: 10, Em: momento},
		{Cliente: "Maria", Produto: Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450}, Quantidade: 3, Em: momento},
	}
}

func TestFiltrarSubstituiAsFuncoesAntigas(t *testing.T) {
	// Este teste é o argumento do módulo. A mesma função genérica produz o
	// resultado das duas funções escritas à mão, que trabalhavam em tipos
	// diferentes e tinham corpo idêntico.
	cardapio := cardapioDeTeste()
	pedidos := pedidosDeTeste()

	disponiveis := Filtrar(cardapio, func(p Produto) bool { return p.Disponivel })
	if !slices.Equal(disponiveis, ProdutosDisponiveis(cardapio)) {
		t.Errorf("Filtrar não reproduziu ProdutosDisponiveis:\n%+v\n%+v",
			disponiveis, ProdutosDisponiveis(cardapio))
	}

	daMaria := Filtrar(pedidos, func(p Pedido) bool { return p.Cliente == "Maria" })
	if !slices.Equal(daMaria, PedidosDoCliente(pedidos, "Maria")) {
		t.Errorf("Filtrar não reproduziu PedidosDoCliente:\n%+v\n%+v",
			daMaria, PedidosDoCliente(pedidos, "Maria"))
	}
}

func TestFiltrar(t *testing.T) {
	numeros := []int{1, 2, 3, 4, 5, 6}

	pares := Filtrar(numeros, func(n int) bool { return n%2 == 0 })
	if esperado := []int{2, 4, 6}; !slices.Equal(pares, esperado) {
		t.Errorf("Filtrar() = %v, esperado %v", pares, esperado)
	}

	nenhum := Filtrar(numeros, func(int) bool { return false })
	if len(nenhum) != 0 {
		t.Errorf("Filtrar() = %v, esperado vazio", nenhum)
	}

	var vazio []int
	if resultado := Filtrar(vazio, func(int) bool { return true }); len(resultado) != 0 {
		t.Errorf("Filtrar(nil) = %v, esperado vazio", resultado)
	}
}

func TestMapearSubstituiNomesDeProdutos(t *testing.T) {
	cardapio := cardapioDeTeste()

	nomes := Mapear(cardapio, func(p Produto) string { return p.Nome })

	if !slices.Equal(nomes, NomesDeProdutos(cardapio)) {
		t.Errorf("Mapear não reproduziu NomesDeProdutos:\n%v\n%v", nomes, NomesDeProdutos(cardapio))
	}
}

func TestMapearTrocaDeTipo(t *testing.T) {
	// T entra como Produto e R sai como int. Esta troca de tipo é o motivo de
	// Mapear precisar de dois parâmetros de tipo.
	precos := Mapear(cardapioDeTeste(), func(p Produto) int { return p.PrecoEmCentavos })

	if esperado := []int{300, 450, 100, 700}; !slices.Equal(precos, esperado) {
		t.Errorf("Mapear() = %v, esperado %v", precos, esperado)
	}
}

func TestSomar(t *testing.T) {
	if recebido := Somar([]int{1, 2, 3}); recebido != 6 {
		t.Errorf("Somar([]int) = %d, esperado 6", recebido)
	}
	if recebido := Somar([]float64{1.5, 2.5}); recebido != 4.0 {
		t.Errorf("Somar([]float64) = %v, esperado 4", recebido)
	}
	if recebido := Somar([]int{}); recebido != 0 {
		t.Errorf("Somar(vazio) = %d, esperado 0", recebido)
	}
}

func TestSomarAceitaTipoComNomeProprio(t *testing.T) {
	// Este é o teste do til. Centavos tem int como tipo subjacente, mas não é
	// int. Se a restrição Numero disser apenas `int`, sem o til, esta linha
	// nem compila.
	valores := []Centavos{300, 450, 100}

	if recebido := Somar(valores); recebido != Centavos(850) {
		t.Errorf("Somar([]Centavos) = %d, esperado 850", recebido)
	}
}

func TestSomarOTotalDeUmaLista(t *testing.T) {
	// Composição: Mapear produz os totais, Somar junta. Duas funções
	// genéricas resolvendo um problema que a padaria tem de verdade.
	totais := Mapear(pedidosDeTeste(), TotalEmCentavos)

	if recebido := Somar(totais); recebido != 2950 {
		t.Errorf("faturamento = %d, esperado 2950", recebido)
	}
}

func TestMaiorPor(t *testing.T) {
	cardapio := cardapioDeTeste()

	// Chave numérica.
	maisCaro, encontrou := MaiorPor(cardapio, func(p Produto) int { return p.PrecoEmCentavos })
	if !encontrou {
		t.Fatal("esperava encontrar algum produto")
	}
	if maisCaro.Nome != "Sonho" {
		t.Errorf("mais caro = %q, esperado Sonho", maisCaro.Nome)
	}

	// Chave de texto, no mesmo código genérico. cmp.Ordered cobre string.
	ultimoNaOrdem, _ := MaiorPor(cardapio, func(p Produto) string { return p.Nome })
	if ultimoNaOrdem.Nome != "Sonho" {
		t.Errorf("último em ordem alfabética = %q, esperado Sonho", ultimoNaOrdem.Nome)
	}
}

func TestMaiorPorEmpateFicaComOPrimeiro(t *testing.T) {
	produtos := []Produto{
		{Nome: "Broa", PrecoEmCentavos: 300},
		{Nome: "Bolo", PrecoEmCentavos: 300},
	}

	maior, _ := MaiorPor(produtos, func(p Produto) int { return p.PrecoEmCentavos })

	if maior.Nome != "Broa" {
		t.Errorf("empate = %q, esperado Broa", maior.Nome)
	}
}

func TestMaiorPorEmFatiaVazia(t *testing.T) {
	var vazio []Produto

	maior, encontrou := MaiorPor(vazio, func(p Produto) int { return p.PrecoEmCentavos })

	if encontrou {
		t.Error("não deveria ter encontrado nada")
	}
	if maior != (Produto{}) {
		t.Errorf("esperava o valor zero de Produto, veio %+v", maior)
	}
}

func TestAgruparPor(t *testing.T) {
	grupos := AgruparPor(pedidosDeTeste(), func(p Pedido) string { return p.Cliente })

	if len(grupos) != 2 {
		t.Fatalf("vieram %d grupos, esperado 2", len(grupos))
	}
	if len(grupos["Maria"]) != 2 {
		t.Errorf("Maria tem %d pedidos, esperado 2", len(grupos["Maria"]))
	}
	if len(grupos["João"]) != 1 {
		t.Errorf("João tem %d pedidos, esperado 1", len(grupos["João"]))
	}

	// A ordem dentro de cada grupo é a ordem original.
	if grupos["Maria"][0].Produto.Nome != "Broa" {
		t.Errorf("primeiro pedido da Maria = %q, esperado Broa", grupos["Maria"][0].Produto.Nome)
	}
}

func TestAgruparPorChaveNaoTextual(t *testing.T) {
	// A restrição é comparable, então bool serve de chave tão bem quanto
	// string. O que não serve é fatia, que não é comparável.
	grupos := AgruparPor(cardapioDeTeste(), func(p Produto) bool { return p.Disponivel })

	if len(grupos[true]) != 3 {
		t.Errorf("disponíveis = %d, esperado 3", len(grupos[true]))
	}
	if len(grupos[false]) != 1 {
		t.Errorf("esgotados = %d, esperado 1", len(grupos[false]))
	}
}

func TestAgruparPorVazioDevolveMapUsavel(t *testing.T) {
	var vazio []Pedido

	grupos := AgruparPor(vazio, func(p Pedido) string { return p.Cliente })

	if grupos == nil {
		t.Fatal("AgruparPor devolveu map nil; escrever nele derrubaria o programa")
	}
	if len(grupos) != 0 {
		t.Errorf("vieram %d grupos, esperado nenhum", len(grupos))
	}
}

func TestPilhaDeInteiros(t *testing.T) {
	// O valor zero já funciona: nenhuma NovaPilha foi chamada.
	var pilha Pilha[int]

	if pilha.Tamanho() != 0 {
		t.Errorf("pilha nova tem %d itens, esperado 0", pilha.Tamanho())
	}

	if _, tinha := pilha.Desempilhar(); tinha {
		t.Error("pilha vazia não deveria devolver item")
	}

	pilha.Empilhar(1)
	pilha.Empilhar(2)
	pilha.Empilhar(3)

	if pilha.Tamanho() != 3 {
		t.Errorf("Tamanho() = %d, esperado 3", pilha.Tamanho())
	}

	for _, esperado := range []int{3, 2, 1} {
		recebido, tinha := pilha.Desempilhar()
		if !tinha {
			t.Fatalf("esperava desempilhar %d, mas a pilha disse estar vazia", esperado)
		}
		if recebido != esperado {
			t.Errorf("Desempilhar() = %d, esperado %d", recebido, esperado)
		}
	}

	if pilha.Tamanho() != 0 {
		t.Errorf("depois de esvaziar, Tamanho() = %d, esperado 0", pilha.Tamanho())
	}
}

func TestPilhaDeProdutos(t *testing.T) {
	// O mesmo tipo, guardando struct. Nenhuma conversão, nenhum any, e o
	// compilador recusaria empilhar um int aqui.
	var pilha Pilha[Produto]

	pilha.Empilhar(Produto{Nome: "Broa", PrecoEmCentavos: 300})
	pilha.Empilhar(Produto{Nome: "Sonho", PrecoEmCentavos: 700})

	topo, tinha := pilha.Desempilhar()
	if !tinha {
		t.Fatal("esperava um produto")
	}
	if topo.Nome != "Sonho" {
		t.Errorf("topo = %q, esperado Sonho", topo.Nome)
	}
}

func TestPilhaVaziaDevolveOValorZero(t *testing.T) {
	var pilha Pilha[Produto]

	produto, tinha := pilha.Desempilhar()

	if tinha {
		t.Error("pilha vazia não deveria devolver item")
	}
	if produto != (Produto{}) {
		t.Errorf("esperava o valor zero de Produto, veio %+v", produto)
	}
}

func TestDemonstracaoInferenciaDeTipo(t *testing.T) {
	// Este teste já passa depois que Filtrar existir. Ele mostra que as duas
	// formas são a mesma coisa: você quase nunca escreve os tipos entre
	// colchetes, porque o compilador os deduz dos argumentos.
	numeros := []int{1, 2, 3, 4}
	ehPar := func(n int) bool { return n%2 == 0 }

	inferido := Filtrar(numeros, ehPar)
	explicito := Filtrar[int](numeros, ehPar)

	if !slices.Equal(inferido, explicito) {
		t.Fatalf("as duas formas deveriam dar no mesmo: %v e %v", inferido, explicito)
	}

	t.Log("Filtrar(numeros, ehPar) e Filtrar[int](numeros, ehPar) são a mesma chamada")
}
