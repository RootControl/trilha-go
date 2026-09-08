package padaria

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ----------------------------------------------------------------------
// Estoque seguro.
//
// Rode este arquivo com -race. Sem ele, uma corrida pode passar despercebida
// mil vezes e falhar na produção:
//
//	go test -race ./modulos/12-concorrencia/...
// ----------------------------------------------------------------------

func TestEstoqueSeguroSobEscritaConcorrente(t *testing.T) {
	estoque := NovoEstoqueSeguro()
	if estoque == nil {
		t.Fatal("NovoEstoqueSeguro devolveu nil")
	}

	const goroutines = 50
	const reposicoes = 100

	var grupo sync.WaitGroup
	for range goroutines {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			for range reposicoes {
				estoque.Repor("Broa", 1)
			}
		}()
	}
	grupo.Wait()

	esperado := goroutines * reposicoes
	if recebido := estoque.Quantidade("Broa"); recebido != esperado {
		t.Errorf("quantidade = %d, esperado %d: reposições se perderam", recebido, esperado)
	}
}

func TestEstoqueSeguroNaoDeixaOEstoqueNegativo(t *testing.T) {
	// Este é o teste que pega o bug de verificar-e-agir. Se a verificação e a
	// subtração acontecerem em travas separadas, mais goroutines passam pela
	// verificação do que há estoque, e o total vendido passa de 40.
	const disponivel = 40
	const tentativas = 200

	estoque := NovoEstoqueSeguro()
	estoque.Repor("Broa", disponivel)

	var vendidas atomic.Int64

	var grupo sync.WaitGroup
	for range tentativas {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			if err := estoque.Baixar("Broa", 1); err == nil {
				vendidas.Add(1)
			}
		}()
	}
	grupo.Wait()

	if vendidas.Load() != disponivel {
		t.Errorf("vendeu %d unidades, esperado exatamente %d", vendidas.Load(), disponivel)
	}
	if restante := estoque.Quantidade("Broa"); restante != 0 {
		t.Errorf("sobrou %d, esperado 0", restante)
	}
}

func TestEstoqueSeguroLendoEEscrevendoAoMesmoTempo(t *testing.T) {
	// Este teste quase não afirma nada: ele existe para o -race olhar. Uma
	// leitura simultânea a uma escrita já é corrida, mesmo que o valor lido
	// pareça razoável.
	estoque := NovoEstoqueSeguro()
	estoque.Repor("Broa", 1000)

	var grupo sync.WaitGroup
	for range 20 {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			for range 100 {
				estoque.Repor("Broa", 1)
			}
		}()

		grupo.Add(1)
		go func() {
			defer grupo.Done()
			for range 100 {
				_ = estoque.Quantidade("Broa")
			}
		}()
	}
	grupo.Wait()
}

func TestEstoqueSeguroRecusa(t *testing.T) {
	casos := []struct {
		nome              string
		produto           string
		quantidade        int
		sentinelaEsperada error
	}{
		{"quantidade zero", "Broa", 0, ErrQuantidadeInvalida},
		{"quantidade negativa", "Broa", -1, ErrQuantidadeInvalida},
		{"produto que não existe", "Croissant", 1, ErrProdutoNaoEncontrado},
		{"mais do que tem", "Broa", 99, ErrEstoqueInsuficiente},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			estoque := NovoEstoqueSeguro()
			estoque.Repor("Broa", 5)

			err := estoque.Baixar(caso.produto, caso.quantidade)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}
			if restante := estoque.Quantidade("Broa"); restante != 5 {
				t.Errorf("o estoque mudou para %d numa baixa recusada", restante)
			}
		})
	}
}

// ----------------------------------------------------------------------
// Assar em paralelo.
// ----------------------------------------------------------------------

// fornoQueConta registra quantas fornadas estiveram no forno ao mesmo tempo.
//
// Os contadores são atomic.Int64 porque este próprio dublê é usado por várias
// goroutines. Para um contador simples, atômico é mais leve que mutex.
type fornoQueConta struct {
	demora      time.Duration
	simultaneas atomic.Int64
	maximo      atomic.Int64
}

func (f *fornoQueConta) Assar(pedido PedidoDeFornada) Fornada {
	agora := f.simultaneas.Add(1)

	for {
		maiorAteAgora := f.maximo.Load()
		if agora <= maiorAteAgora || f.maximo.CompareAndSwap(maiorAteAgora, agora) {
			break
		}
	}

	time.Sleep(f.demora)
	f.simultaneas.Add(-1)

	return Fornada{Produto: pedido.Produto, Quantidade: pedido.Quantidade}
}

func pedidosDeTeste() []PedidoDeFornada {
	return []PedidoDeFornada{
		{Produto: "Broa", Quantidade: 5},
		{Produto: "Pão de queijo", Quantidade: 20},
		{Produto: "Pão francês", Quantidade: 100},
		{Produto: "Sonho", Quantidade: 3},
	}
}

func TestAssarTudoPreservaAOrdem(t *testing.T) {
	// As fornadas ficam prontas em ordem imprevisível, e mesmo assim a saída
	// precisa acompanhar a entrada. É o que a fatia dimensionada por índice
	// resolve.
	pedidos := pedidosDeTeste()

	fornadas := AssarTudo(FornoLento{Demora: 2 * time.Millisecond}, pedidos)

	if len(fornadas) != len(pedidos) {
		t.Fatalf("vieram %d fornadas, esperado %d", len(fornadas), len(pedidos))
	}
	for i, pedido := range pedidos {
		if fornadas[i].Produto != pedido.Produto {
			t.Errorf("posição %d = %q, esperado %q", i, fornadas[i].Produto, pedido.Produto)
		}
		if fornadas[i].Quantidade != pedido.Quantidade {
			t.Errorf("posição %d: quantidade = %d, esperado %d",
				i, fornadas[i].Quantidade, pedido.Quantidade)
		}
	}
}

func TestAssarTudoAssaDeVerdadeAoMesmoTempo(t *testing.T) {
	// Em vez de cronometrar, que dá teste instável, o forno conta quantas
	// fornadas estiveram dentro dele simultaneamente.
	forno := &fornoQueConta{demora: 20 * time.Millisecond}

	AssarTudo(forno, pedidosDeTeste())

	if maximo := forno.maximo.Load(); maximo < 2 {
		t.Errorf("no máximo %d fornada esteve no forno por vez: as goroutines estão em fila", maximo)
	}
}

func TestAssarTudoComListaVazia(t *testing.T) {
	fornadas := AssarTudo(FornoLento{}, nil)

	if len(fornadas) != 0 {
		t.Errorf("vieram %d fornadas, esperado nenhuma", len(fornadas))
	}
}

// ----------------------------------------------------------------------
// Canais.
// ----------------------------------------------------------------------

func TestAssarEmCanalEntregaTudoEFecha(t *testing.T) {
	pedidos := pedidosDeTeste()

	canal := AssarEmCanal(FornoLento{Demora: 2 * time.Millisecond}, pedidos)

	// Se o canal não for fechado, este for range espera para sempre e o teste
	// morre por prazo esgotado em vez de falhar bonito.
	fornadas := Coletar(canal)

	if len(fornadas) != len(pedidos) {
		t.Fatalf("vieram %d fornadas, esperado %d", len(fornadas), len(pedidos))
	}

	// Coletar ordena por nome, então a comparação é previsível.
	esperado := []string{"Broa", "Pão de queijo", "Pão francês", "Sonho"}
	for i, nome := range esperado {
		if fornadas[i].Produto != nome {
			t.Errorf("posição %d = %q, esperado %q", i, fornadas[i].Produto, nome)
		}
	}
}

func TestAssarEmCanalDevolveNaHora(t *testing.T) {
	// A função devolve o canal imediatamente e o trabalho continua ao fundo.
	// Se ela esperasse tudo assar antes de devolver, esta medição estouraria.
	inicio := time.Now()

	canal := AssarEmCanal(FornoLento{Demora: 200 * time.Millisecond}, pedidosDeTeste())

	if decorrido := time.Since(inicio); decorrido > 100*time.Millisecond {
		t.Errorf("AssarEmCanal demorou %v para devolver: ela não pode esperar o forno", decorrido)
	}

	Coletar(canal) // não deixa goroutine pendurada
}

func TestColetarComCanalJaFechado(t *testing.T) {
	canal := make(chan Fornada)
	close(canal)

	if fornadas := Coletar(canal); len(fornadas) != 0 {
		t.Errorf("vieram %d fornadas, esperado nenhuma", len(fornadas))
	}
}

func TestEsperarComPrazo(t *testing.T) {
	t.Run("chega a tempo", func(t *testing.T) {
		canal := AssarEmCanal(FornoLento{Demora: 2 * time.Millisecond}, pedidosDeTeste()[:1])

		fornada, chegou := EsperarComPrazo(canal, time.Second)

		if !chegou {
			t.Fatal("esperava a fornada dentro do prazo")
		}
		if fornada.Produto != "Broa" {
			t.Errorf("Produto = %q, esperado Broa", fornada.Produto)
		}
	})

	t.Run("estoura o prazo", func(t *testing.T) {
		canal := make(chan Fornada) // ninguém vai escrever

		inicio := time.Now()
		_, chegou := EsperarComPrazo(canal, 30*time.Millisecond)

		if chegou {
			t.Error("nada foi enviado, não era para chegar nada")
		}
		if decorrido := time.Since(inicio); decorrido < 25*time.Millisecond {
			t.Errorf("desistiu em %v, antes do prazo de 30ms", decorrido)
		}
	})

	t.Run("canal fechado não é fornada", func(t *testing.T) {
		canal := make(chan Fornada)
		close(canal)

		if _, chegou := EsperarComPrazo(canal, time.Second); chegou {
			t.Error("canal fechado devolve o valor zero, e isso não é uma fornada")
		}
	})
}

func TestContarPorProduto(t *testing.T) {
	fornadas := []Fornada{
		{Produto: "Broa", Quantidade: 5},
		{Produto: "Broa", Quantidade: 3},
		{Produto: "Sonho", Quantidade: 2},
	}

	total := ContarPorProduto(fornadas)

	if total["Broa"] != 8 {
		t.Errorf("Broa = %d, esperado 8", total["Broa"])
	}
	if total["Sonho"] != 2 {
		t.Errorf("Sonho = %d, esperado 2", total["Sonho"])
	}
	if len(total) != 2 {
		t.Errorf("vieram %d produtos, esperado 2", len(total))
	}
}

// ----------------------------------------------------------------------
// Demonstrações.
// ----------------------------------------------------------------------

func TestDemonstracaoCanalFechadoDevolveOValorZero(t *testing.T) {
	// Este teste já passa. Receber de canal fechado NÃO bloqueia e NÃO entra
	// em pânico: devolve na hora o valor zero, com o segundo retorno em false.
	// Confundir isso com "chegou um valor vazio" é bug comum.
	canal := make(chan Fornada, 1)
	canal <- Fornada{Produto: "Broa", Quantidade: 1}
	close(canal)

	primeira, aberto := <-canal
	if !aberto || primeira.Produto != "Broa" {
		t.Fatal("o valor enviado antes do close ainda deveria estar lá")
	}

	segunda, aindaAberto := <-canal
	if aindaAberto {
		t.Fatal("o canal está fechado e vazio")
	}
	if segunda != (Fornada{}) {
		t.Fatalf("esperava o valor zero, veio %+v", segunda)
	}

	t.Log("fechar não descarta o que já estava no canal; depois disso vem o valor zero com aberto=false")
}

func TestDemonstracaoCanalSemBufferSincroniza(t *testing.T) {
	// Este teste já passa. Num canal sem buffer, o envio só termina quando
	// alguém recebe. Os dois lados se encontram, e é isso que faz o canal ser
	// uma ferramenta de sincronização, e não só um cano de dados.
	canal := make(chan Fornada)
	ordem := make(chan string, 2)

	go func() {
		ordem <- "antes do envio"
		canal <- Fornada{Produto: "Broa"}
		ordem <- "depois do envio"
	}()

	time.Sleep(20 * time.Millisecond)
	if len(ordem) != 1 {
		t.Fatalf("o envio deveria estar parado esperando alguém receber, ordem tem %d", len(ordem))
	}

	<-canal

	if <-ordem != "antes do envio" {
		t.Fatal("ordem inesperada")
	}
	if <-ordem != "depois do envio" {
		t.Fatal("o envio só destrava depois da recepção")
	}

	t.Log("o envio ficou parado 20ms e só seguiu quando alguém recebeu")
}

func TestDemonstracaoCorridaNoMapComum(t *testing.T) {
	// Esta é a resposta à pergunta deixada no fim do módulo 10.
	//
	// O teste é pulado por padrão porque ele TEM uma corrida de propósito, e
	// faria o CI falhar. Rode assim, para ver o detector trabalhar:
	//
	//	MOSTRAR_CORRIDA=1 go test -race -run Corrida ./modulos/12-concorrencia/...
	if os.Getenv("MOSTRAR_CORRIDA") == "" {
		t.Skip("rode com MOSTRAR_CORRIDA=1 go test -race -run Corrida ./modulos/12-concorrencia/...")
	}

	estoque := NovoEstoque() // o map comum, sem proteção nenhuma

	var grupo sync.WaitGroup
	for range 100 {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			estoque.Repor("Broa", 1)
		}()
	}
	grupo.Wait()

	t.Logf("esperado 100, veio %d", estoque.Quantidade("Broa"))
}
