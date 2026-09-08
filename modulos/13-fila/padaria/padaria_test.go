package padaria

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func pedidosDeTeste(quantos int) []PedidoDeFornada {
	pedidos := make([]PedidoDeFornada, quantos)
	for i := range pedidos {
		pedidos[i] = PedidoDeFornada{Produto: "Broa", Quantidade: i + 1}
	}
	return pedidos
}

// fornoQueConta registra quantas fornadas estiveram no forno ao mesmo tempo.
type fornoQueConta struct {
	demora      time.Duration
	simultaneas atomic.Int64
	maximo      atomic.Int64
	assadas     atomic.Int64
}

func (f *fornoQueConta) Assar(ctx context.Context, pedido PedidoDeFornada) (Fornada, error) {
	agora := f.simultaneas.Add(1)
	for {
		maior := f.maximo.Load()
		if agora <= maior || f.maximo.CompareAndSwap(maior, agora) {
			break
		}
	}
	defer f.simultaneas.Add(-1)

	temporizador := time.NewTimer(f.demora)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		f.assadas.Add(1)
		return Fornada{Produto: pedido.Produto, Quantidade: pedido.Quantidade}, nil
	case <-ctx.Done():
		return Fornada{}, ctx.Err()
	}
}

// fornoQueQuebra falha na enésima fornada.
type fornoQueQuebra struct {
	falharNa int64
	assadas  atomic.Int64
	err      error
}

func (f *fornoQueQuebra) Assar(ctx context.Context, pedido PedidoDeFornada) (Fornada, error) {
	if f.assadas.Add(1) == f.falharNa {
		return Fornada{}, f.err
	}
	return Fornada{Produto: pedido.Produto, Quantidade: pedido.Quantidade}, nil
}

// ----------------------------------------------------------------------
// Espera cancelável.
// ----------------------------------------------------------------------

func TestAssarNoCaminhoFeliz(t *testing.T) {
	forno := FornoLento{Demora: 5 * time.Millisecond}

	fornada, err := forno.Assar(context.Background(), PedidoDeFornada{Produto: "Broa", Quantidade: 3})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fornada.Produto != "Broa" || fornada.Quantidade != 3 {
		t.Errorf("fornada = %+v, esperado {Broa 3}", fornada)
	}
}

func TestAssarDesisteQuandoOContextoEhCancelado(t *testing.T) {
	// Um time.Sleep comum não teria como sair daqui: a espera de dez segundos
	// aconteceria inteira. O teste mede que a desistência é rápida.
	forno := FornoLento{Demora: 10 * time.Second}

	ctx, cancelar := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancelar()
	}()
	defer cancelar()

	inicio := time.Now()
	_, err := forno.Assar(ctx, PedidoDeFornada{Produto: "Broa", Quantidade: 1})
	decorrido := time.Since(inicio)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is não achou context.Canceled em %v", err)
	}
	if decorrido > time.Second {
		t.Errorf("demorou %v para desistir: a espera não é cancelável", decorrido)
	}
}

func TestAssarComContextoJaCancelado(t *testing.T) {
	forno := FornoLento{Demora: 10 * time.Second}

	ctx, cancelar := context.WithCancel(context.Background())
	cancelar() // já nasce cancelado

	inicio := time.Now()
	_, err := forno.Assar(ctx, PedidoDeFornada{Produto: "Broa", Quantidade: 1})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is não achou context.Canceled em %v", err)
	}
	if decorrido := time.Since(inicio); decorrido > 100*time.Millisecond {
		t.Errorf("demorou %v com contexto já cancelado", decorrido)
	}
}

func TestAssarComPrazo(t *testing.T) {
	t.Run("dá tempo", func(t *testing.T) {
		forno := FornoLento{Demora: 5 * time.Millisecond}

		fornada, err := AssarComPrazo(context.Background(), forno,
			PedidoDeFornada{Produto: "Broa", Quantidade: 2}, time.Second)

		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if fornada.Quantidade != 2 {
			t.Errorf("Quantidade = %d, esperado 2", fornada.Quantidade)
		}
	})

	t.Run("estoura o prazo", func(t *testing.T) {
		forno := FornoLento{Demora: 10 * time.Second}

		inicio := time.Now()
		_, err := AssarComPrazo(context.Background(), forno,
			PedidoDeFornada{Produto: "Broa", Quantidade: 1}, 30*time.Millisecond)

		// Prazo estourado e cancelamento são erros DIFERENTES, e quem chamou
		// costuma querer reagir de formas diferentes a cada um.
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("errors.Is não achou context.DeadlineExceeded em %v", err)
		}
		if errors.Is(err, context.Canceled) {
			t.Error("prazo estourado não é cancelamento")
		}
		if decorrido := time.Since(inicio); decorrido > time.Second {
			t.Errorf("demorou %v, esperado desistir em 30ms", decorrido)
		}
	})

	t.Run("o pai continua válido depois", func(t *testing.T) {
		// O prazo vale só para o filho. Cancelar o filho não pode estragar o
		// contexto de quem chamou.
		ctxPai := context.Background()
		forno := FornoLento{Demora: 10 * time.Second}

		_, _ = AssarComPrazo(ctxPai, forno, PedidoDeFornada{Produto: "Broa"}, 10*time.Millisecond)

		if err := ctxPai.Err(); err != nil {
			t.Errorf("o contexto do pai foi afetado: %v", err)
		}
	})
}

// ----------------------------------------------------------------------
// A fila.
// ----------------------------------------------------------------------

func TestProcessarFilaPreservaAOrdem(t *testing.T) {
	pedidos := pedidosDeTeste(10)

	fornadas, err := ProcessarFila(context.Background(), FornoLento{Demora: time.Millisecond}, 3, pedidos)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(fornadas) != len(pedidos) {
		t.Fatalf("vieram %d fornadas, esperado %d", len(fornadas), len(pedidos))
	}
	for i, pedido := range pedidos {
		if fornadas[i].Quantidade != pedido.Quantidade {
			t.Errorf("posição %d: quantidade = %d, esperado %d",
				i, fornadas[i].Quantidade, pedido.Quantidade)
		}
	}
}

func TestProcessarFilaRespeitaOLimiteDeTrabalhadores(t *testing.T) {
	// Esta é a diferença para o AssarTudo do módulo 12, que disparava uma
	// goroutine por pedido. Aqui, vinte pedidos e três trabalhadores nunca
	// põem quatro fornadas no forno ao mesmo tempo.
	const trabalhadores = 3

	forno := &fornoQueConta{demora: 10 * time.Millisecond}

	if _, err := ProcessarFila(context.Background(), forno, trabalhadores, pedidosDeTeste(20)); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if maximo := forno.maximo.Load(); maximo > trabalhadores {
		t.Errorf("%d fornadas ao mesmo tempo, o limite era %d", maximo, trabalhadores)
	}
	if maximo := forno.maximo.Load(); maximo < 2 {
		t.Errorf("no máximo %d por vez: os trabalhadores estão em fila", maximo)
	}
	if assadas := forno.assadas.Load(); assadas != 20 {
		t.Errorf("assou %d fornadas, esperado 20", assadas)
	}
}

func TestProcessarFilaSemTrabalhadores(t *testing.T) {
	_, err := ProcessarFila(context.Background(), FornoLento{}, 0, pedidosDeTeste(3))

	if !errors.Is(err, ErrTrabalhadoresInvalido) {
		t.Errorf("errors.Is não achou ErrTrabalhadoresInvalido em %v", err)
	}
}

func TestProcessarFilaSemPedidos(t *testing.T) {
	fornadas, err := ProcessarFila(context.Background(), FornoLento{}, 3, nil)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(fornadas) != 0 {
		t.Errorf("vieram %d fornadas, esperado nenhuma", len(fornadas))
	}
}

func TestProcessarFilaComContextoJaCancelado(t *testing.T) {
	ctx, cancelar := context.WithCancel(context.Background())
	cancelar()

	inicio := time.Now()
	_, err := ProcessarFila(ctx, FornoLento{Demora: time.Second}, 3, pedidosDeTeste(100))

	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is não achou context.Canceled em %v", err)
	}
	// Cem pedidos de um segundo cada. Se a fila ignorasse o contexto, isto
	// levaria minutos.
	if decorrido := time.Since(inicio); decorrido > time.Second {
		t.Errorf("demorou %v com contexto já cancelado", decorrido)
	}
}

func TestProcessarFilaCancelaNoMeio(t *testing.T) {
	forno := &fornoQueConta{demora: 20 * time.Millisecond}

	ctx, cancelar := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancelar()
	}()
	defer cancelar()

	inicio := time.Now()
	_, err := ProcessarFila(ctx, forno, 2, pedidosDeTeste(200))
	decorrido := time.Since(inicio)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is não achou context.Canceled em %v", err)
	}
	// Duzentos pedidos, dois trabalhadores, vinte milissegundos cada: dois
	// segundos se fosse até o fim. O cancelamento tem que cortar isso.
	if decorrido > time.Second {
		t.Errorf("demorou %v depois do cancelamento", decorrido)
	}
	if assadas := forno.assadas.Load(); assadas >= 200 {
		t.Errorf("assou %d fornadas: o cancelamento não interrompeu nada", assadas)
	}
}

func TestProcessarFilaComPrazo(t *testing.T) {
	ctx, cancelar := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancelar()

	_, err := ProcessarFila(ctx, FornoLento{Demora: 20 * time.Millisecond}, 2, pedidosDeTeste(100))

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("errors.Is não achou context.DeadlineExceeded em %v", err)
	}
}

func TestProcessarFilaParaNoPrimeiroErro(t *testing.T) {
	erroDoForno := errors.New("o forno apagou")
	forno := &fornoQueQuebra{falharNa: 5, err: erroDoForno}

	_, err := ProcessarFila(context.Background(), forno, 2, pedidosDeTeste(500))

	if !errors.Is(err, erroDoForno) {
		t.Fatalf("errors.Is não achou o erro do forno em %v", err)
	}
	// O erro do trabalhador precisa ganhar do erro de contexto: dizer "o
	// forno apagou" ajuda muito mais que dizer "cancelado".
	if errors.Is(err, context.Canceled) {
		t.Error("o erro devolvido deveria ser o do forno, não o do cancelamento")
	}
	if assadas := forno.assadas.Load(); assadas >= 500 {
		t.Errorf("tentou assar %d: a fila não parou no primeiro erro", assadas)
	}
}

// ----------------------------------------------------------------------
// Valores no contexto.
// ----------------------------------------------------------------------

func TestClienteNoContexto(t *testing.T) {
	ctx := ContextoComCliente(context.Background(), "Maria")

	cliente, tem := ClienteDoContexto(ctx)
	if !tem {
		t.Fatal("esperava achar o cliente")
	}
	if cliente != "Maria" {
		t.Errorf("cliente = %q, esperado Maria", cliente)
	}
}

func TestClienteAusenteNoContexto(t *testing.T) {
	if _, tem := ClienteDoContexto(context.Background()); tem {
		t.Error("contexto vazio não deveria ter cliente")
	}
}

func TestClienteAtravessaContextosFilhos(t *testing.T) {
	// Valor guardado no pai continua visível nos filhos, inclusive nos que
	// foram criados por outro motivo, como um prazo.
	ctx := ContextoComCliente(context.Background(), "Maria")

	ctxFilho, cancelar := context.WithTimeout(ctx, time.Minute)
	defer cancelar()

	cliente, tem := ClienteDoContexto(ctxFilho)
	if !tem || cliente != "Maria" {
		t.Errorf("o filho perdeu o valor do pai: %q, %v", cliente, tem)
	}
}

// ----------------------------------------------------------------------
// Demonstrações.
// ----------------------------------------------------------------------

func TestDemonstracaoOsDoisErrosDeContextoSaoDiferentes(t *testing.T) {
	// Este teste já passa. Cancelamento e prazo estourado são sentinelas
	// distintos, e confundir os dois é comum. Um significa "alguém desistiu",
	// o outro significa "demorou demais", e um serviço reage diferente a cada.
	cancelado, cancelar := context.WithCancel(context.Background())
	cancelar()

	vencido, cancelar2 := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancelar2()
	<-vencido.Done()

	if !errors.Is(cancelado.Err(), context.Canceled) {
		t.Fatal("esperava context.Canceled")
	}
	if errors.Is(cancelado.Err(), context.DeadlineExceeded) {
		t.Fatal("cancelamento não é prazo estourado")
	}
	if !errors.Is(vencido.Err(), context.DeadlineExceeded) {
		t.Fatal("esperava context.DeadlineExceeded")
	}

	t.Logf("cancelado: %v | vencido: %v", cancelado.Err(), vencido.Err())
}

func TestDemonstracaoCancelarFilhoNaoAfetaOPai(t *testing.T) {
	// Este teste já passa. O cancelamento desce na árvore, nunca sobe.
	pai, cancelarPai := context.WithCancel(context.Background())
	defer cancelarPai()

	filho, cancelarFilho := context.WithCancel(pai)
	neto, cancelarNeto := context.WithCancel(filho)
	defer cancelarNeto()

	cancelarFilho()

	if filho.Err() == nil {
		t.Fatal("o filho deveria estar cancelado")
	}
	if neto.Err() == nil {
		t.Fatal("o neto deveria ter sido cancelado junto com o pai dele")
	}
	if pai.Err() != nil {
		t.Fatal("o pai não pode ser afetado pelo filho")
	}

	t.Log("cancelamento desce a árvore inteira e nunca sobe")
}
