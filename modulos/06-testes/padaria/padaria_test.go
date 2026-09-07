package padaria

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"
)

// Este arquivo é o material de referência do módulo. Além de checar o seu
// código, ele usa, comentada, cada técnica que o README explica. Leia de cima
// a baixo antes de implementar qualquer coisa.

// ----------------------------------------------------------------------
// Auxiliares. Repare no t.Helper() em cada um.
// ----------------------------------------------------------------------

// exigirSemErro interrompe o teste se houver erro.
//
// A chamada a t.Helper() faz o Go apontar a falha para a linha de QUEM chamou
// este auxiliar, e não para a linha de dentro dele. Sem isso, todas as falhas
// do arquivo apontariam para o mesmo lugar inútil.
func exigirSemErro(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}

// padariaDeTeste monta o cenário que quase todo teste daqui precisa.
//
// Um auxiliar que constrói o cenário costuma ser chamado de fixture. Ele
// devolve tudo pronto e falha rápido se o preparo der errado, porque um teste
// que quebra no preparo não está testando nada.
func padariaDeTeste(t *testing.T) (*RepositorioEmMemoria, Estoque) {
	t.Helper()

	repositorio := NovoRepositorioEmMemoria()
	exigirSemErro(t, repositorio.Salvar(Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}))
	exigirSemErro(t, repositorio.Salvar(Produto{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true}))

	estoque := NovoEstoque()
	estoque.Repor("Pão de queijo", 10)
	estoque.Repor("Pão francês", 50)

	return repositorio, estoque
}

// ----------------------------------------------------------------------
// Relógio: tirando time.Now da regra de negócio.
// ----------------------------------------------------------------------

func TestRelogioFixoDevolveSempreOMesmoInstante(t *testing.T) {
	momento := time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)
	relogio := RelogioFixo{Momento: momento}

	primeira := relogio.Agora()
	segunda := relogio.Agora()

	if !primeira.Equal(momento) {
		t.Errorf("Agora() = %v, esperado %v", primeira, momento)
	}
	if !primeira.Equal(segunda) {
		t.Error("duas chamadas devolveram instantes diferentes: o relógio precisa ser fixo")
	}
}

func TestRelogioDoSistemaDevolveAHoraDeVerdade(t *testing.T) {
	antes := time.Now()
	agora := RelogioDoSistema{}.Agora()
	depois := time.Now()

	if agora.Before(antes) || agora.After(depois) {
		t.Errorf("Agora() = %v, esperado algo entre %v e %v", agora, antes, depois)
	}
}

func TestNovoPedidoCarimbaAHoraDoRelogio(t *testing.T) {
	// Este é o teste que só existe porque o relógio virou parâmetro. Sem a
	// interface, a única forma de checar o carimbo seria comparar com
	// time.Now() e torcer para a máquina não estar lenta naquele instante.
	momento := time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)
	produto := Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true}

	pedido := NovoPedido(RelogioFixo{Momento: momento}, "Maria", produto, 3)

	if !pedido.Em.Equal(momento) {
		t.Errorf("Em = %v, esperado %v", pedido.Em, momento)
	}
	if pedido.Cliente != "Maria" {
		t.Errorf("Cliente = %q, esperado %q", pedido.Cliente, "Maria")
	}
	if pedido.Quantidade != 3 {
		t.Errorf("Quantidade = %d, esperado 3", pedido.Quantidade)
	}
	if pedido.Produto != produto {
		t.Errorf("Produto = %+v, esperado %+v", pedido.Produto, produto)
	}
}

// ----------------------------------------------------------------------
// Espião: um dublê que guarda o que recebeu.
// ----------------------------------------------------------------------

func TestRegistradorEmMemoriaGuardaNaOrdem(t *testing.T) {
	var registrador RegistradorEmMemoria

	registrador.Registrar("primeira")
	registrador.Registrar("segunda")

	esperado := []string{"primeira", "segunda"}
	if recebido := registrador.Linhas(); !slices.Equal(recebido, esperado) {
		t.Errorf("Linhas() = %v, esperado %v", recebido, esperado)
	}
}

func TestLinhasDevolveCopia(t *testing.T) {
	// Se Linhas devolver a fatia interna, quem chamou consegue estragar o
	// espião sem querer. Módulo 02, agora aplicado a um detalhe de projeto.
	var registrador RegistradorEmMemoria
	registrador.Registrar("original")

	linhas := registrador.Linhas()
	if len(linhas) != 1 {
		t.Fatalf("Linhas() devolveu %d linhas, esperado 1", len(linhas))
	}

	linhas[0] = "adulterada"

	if depois := registrador.Linhas(); depois[0] != "original" {
		t.Errorf("o espião foi alterado de fora: %q; use slices.Clone", depois[0])
	}
}

func TestVenderComRegistroAnotaOSucesso(t *testing.T) {
	repositorio, estoque := padariaDeTeste(t)
	var registrador RegistradorEmMemoria

	total, err := VenderComRegistro(&registrador, repositorio, estoque, "Pão de queijo", 3)

	exigirSemErro(t, err)
	if total != 1350 {
		t.Errorf("total = %d, esperado 1350", total)
	}

	esperado := []string{`vendido: 3 de "Pão de queijo" por R$ 13,50`}
	if recebido := registrador.Linhas(); !slices.Equal(recebido, esperado) {
		t.Errorf("Linhas() = %q, esperado %q", recebido, esperado)
	}

	if restante := estoque.Quantidade("Pão de queijo"); restante != 7 {
		t.Errorf("sobrou %d, esperado 7", restante)
	}
}

func TestVenderComRegistroAnotaAFalha(t *testing.T) {
	repositorio, estoque := padariaDeTeste(t)
	var registrador RegistradorEmMemoria

	total, err := VenderComRegistro(&registrador, repositorio, estoque, "Pão de queijo", 99)

	// t.Error continua o teste, t.Fatal para na hora. Aqui é Fatal, porque
	// as verificações seguintes não fazem sentido se não houve erro nenhum.
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Errorf("errors.Is não achou ErrEstoqueInsuficiente em %v", err)
	}
	if total != 0 {
		t.Errorf("com erro, o total deveria ser 0, veio %d", total)
	}

	linhas := registrador.Linhas()
	if len(linhas) != 1 {
		t.Fatalf("registrou %d linhas, esperado 1", len(linhas))
	}
	esperado := "falhou: " + err.Error()
	if linhas[0] != esperado {
		t.Errorf("linha = %q, esperado %q", linhas[0], esperado)
	}
}

// ----------------------------------------------------------------------
// Diferença entre produtos: o que uma biblioteca de asserção faria.
// ----------------------------------------------------------------------

func TestDiferencaEntreProdutos(t *testing.T) {
	base := Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}

	casos := []struct {
		nome     string
		esperado Produto
		recebido Produto
		saida    string
	}{
		{
			nome:     "iguais não geram saída",
			esperado: base,
			recebido: base,
			saida:    "",
		},
		{
			nome:     "só o nome",
			esperado: base,
			recebido: Produto{Nome: "Sonho", PrecoEmCentavos: 300, Disponivel: true},
			saida:    `Nome: esperado "Broa", recebido "Sonho"`,
		},
		{
			nome:     "só o preço",
			esperado: base,
			recebido: Produto{Nome: "Broa", PrecoEmCentavos: 700, Disponivel: true},
			saida:    "PrecoEmCentavos: esperado 300, recebido 700",
		},
		{
			nome:     "só a disponibilidade",
			esperado: base,
			recebido: Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: false},
			saida:    "Disponivel: esperado true, recebido false",
		},
		{
			nome:     "tudo diferente, uma linha por campo",
			esperado: base,
			recebido: Produto{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
			saida: `Nome: esperado "Broa", recebido "Sonho"` + "\n" +
				"PrecoEmCentavos: esperado 300, recebido 700" + "\n" +
				"Disponivel: esperado true, recebido false",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			// Estes subtestes não compartilham nada, então podem rodar ao
			// mesmo tempo. Desde o Go 1.22 a variável do laço é nova a cada
			// volta, e este padrão parou de ser uma armadilha.
			t.Parallel()

			if recebido := DiferencaEntreProdutos(caso.esperado, caso.recebido); recebido != caso.saida {
				t.Errorf("DiferencaEntreProdutos() =\n%s\n\nesperado:\n%s", recebido, caso.saida)
			}
		})
	}
}

// ----------------------------------------------------------------------
// Exemplo executável: documentação que o go test roda.
// ----------------------------------------------------------------------

// ExampleDiferencaEntreProdutos aparece na documentação do pacote, e ao mesmo
// tempo é um teste: o go test compara a saída com o bloco Output abaixo. Se a
// função mudar de comportamento, o exemplo da documentação quebra o build em
// vez de ficar mentindo em silêncio.
func ExampleDiferencaEntreProdutos() {
	esperado := Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}
	recebido := Produto{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: true}

	fmt.Println(DiferencaEntreProdutos(esperado, recebido))

	// Output:
	// Nome: esperado "Broa", recebido "Sonho"
	// PrecoEmCentavos: esperado 300, recebido 700
}

// ----------------------------------------------------------------------
// Benchmark: não roda no go test comum.
// ----------------------------------------------------------------------

// BenchmarkFormatarPreco mede quanto custa formatar um preço.
//
// Benchmarks só rodam com -bench. Experimente:
//
//	go test -bench=. -benchmem ./modulos/06-testes/...
//
// b.Loop é a forma atual de escrever o laço, desde o Go 1.24. Ela impede que
// o compilador descarte a chamada por perceber que o resultado não é usado.
func BenchmarkFormatarPreco(b *testing.B) {
	for b.Loop() {
		FormatarPreco(123456)
	}
}
