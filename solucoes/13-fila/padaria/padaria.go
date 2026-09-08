// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 13. Olhe depois de tentar, não antes.
package padaria

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ----------------------------------------------------------------------
// Vindo do módulo 12.
// ----------------------------------------------------------------------

// PedidoDeFornada é uma quantidade de um produto para assar.
type PedidoDeFornada struct {
	Produto    string
	Quantidade int
}

// Fornada é o resultado de assar.
type Fornada struct {
	Produto    string
	Quantidade int
}

// ErrTrabalhadoresInvalido é devolvido quando a fila é chamada sem gente para
// trabalhar.
var ErrTrabalhadoresInvalido = errors.New("número de trabalhadores inválido")

// Forno assa uma fornada.
//
// Compare com a versão do módulo 12: o método ganhou um context.Context na
// frente e um error no fim. Essa é a assinatura de qualquer coisa que possa
// demorar, e ela avisa duas coisas a quem lê: isto pode ser cancelado, e isto
// pode falhar.
type Forno interface {
	Assar(ctx context.Context, pedido PedidoDeFornada) (Fornada, error)
}

// FornoLento demora um pouco para assar cada fornada.
type FornoLento struct {
	Demora time.Duration
}

// ----------------------------------------------------------------------
// Módulo 13.
// ----------------------------------------------------------------------

// Assar espera a demora configurada, mas desiste se o contexto for cancelado.
func (f FornoLento) Assar(ctx context.Context, pedido PedidoDeFornada) (Fornada, error) {
	// time.NewTimer com Stop adiado, e não time.After. O canal devolvido por
	// time.After não é coletado enquanto o prazo não vencer, então uma função
	// que desiste rápido e é chamada muito deixa um rastro de temporizadores
	// vivos. Com NewTimer você tem o Stop.
	temporizador := time.NewTimer(f.Demora)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		return Fornada{Produto: pedido.Produto, Quantidade: pedido.Quantidade}, nil

	case <-ctx.Done():
		// ctx.Err() já distingue context.Canceled de context.DeadlineExceeded.
		// Devolver ele, e não um erro seu, deixa quem chamou usar errors.Is
		// contra os sentinelas da biblioteca padrão.
		return Fornada{}, ctx.Err()
	}
}

// AssarComPrazo assa dando no máximo o prazo indicado.
func AssarComPrazo(ctx context.Context, forno Forno, pedido PedidoDeFornada, prazo time.Duration) (Fornada, error) {
	// O filho herda o cancelamento do pai E ganha um prazo próprio. Quem
	// vencer primeiro cancela o filho; o pai nunca é afetado.
	ctxComPrazo, cancelar := context.WithTimeout(ctx, prazo)
	defer cancelar()

	return forno.Assar(ctxComPrazo, pedido)
}

// ProcessarFila assa todos os pedidos com um número fixo de trabalhadores.
func ProcessarFila(ctx context.Context, forno Forno, trabalhadores int, pedidos []PedidoDeFornada) ([]Fornada, error) {
	if trabalhadores <= 0 {
		return nil, ErrTrabalhadoresInvalido
	}
	if len(pedidos) == 0 {
		return []Fornada{}, nil
	}

	// tarefa carrega o índice junto com o pedido. É o que permite a cada
	// trabalhador escrever no lugar certo da fatia de saída sem coordenação.
	type tarefa struct {
		indice int
		pedido PedidoDeFornada
	}

	// Contexto cancelável PRÓPRIO, derivado do recebido. Assim esta função
	// consegue mandar os seus trabalhadores pararem sem tocar no contexto de
	// quem chamou. O defer garante que, mesmo saindo pelo caminho feliz,
	// ninguém fica pendurado.
	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	fila := make(chan tarefa)
	fornadas := make([]Fornada, len(pedidos))

	// Buffer do tamanho do time: nenhum trabalhador fica preso tentando
	// reportar um erro que ninguém está lendo ainda.
	erros := make(chan error, trabalhadores)

	var grupo sync.WaitGroup

	for range trabalhadores {
		grupo.Add(1)

		go func() {
			defer grupo.Done()

			// for range sobre canal: cada trabalhador pega a próxima tarefa
			// livre e para sozinho quando o canal fecha. Não é preciso
			// dividir os pedidos entre eles: quem termina antes pega mais.
			for t := range fila {
				fornada, err := forno.Assar(ctx, t.pedido)
				if err != nil {
					erros <- err
					cancelar() // avisa os colegas e o alimentador
					return
				}

				fornadas[t.indice] = fornada
			}
		}()
	}

	// Alimentador. Ele fecha a fila ao terminar, e é isso que faz os
	// trabalhadores saírem do for range.
	go func() {
		defer close(fila)

		for i, pedido := range pedidos {
			select {
			case fila <- tarefa{indice: i, pedido: pedido}:
			case <-ctx.Done():
				// Sem este braço, o alimentador ficaria preso tentando
				// entregar uma tarefa a trabalhadores que já desistiram. É
				// assim que se vaza goroutine.
				return
			}
		}
	}()

	grupo.Wait()

	// Erro de trabalhador tem prioridade: ele diz o que aconteceu de verdade,
	// enquanto o contexto só diria que foi cancelado.
	select {
	case err := <-erros:
		return nil, err
	default:
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return fornadas, nil
}

// chaveDeContexto é o tipo das chaves que esta padaria guarda em contexto.
type chaveDeContexto string

const chaveCliente chaveDeContexto = "cliente"

// ContextoComCliente devolve um contexto filho carregando o nome do cliente.
func ContextoComCliente(ctx context.Context, cliente string) context.Context {
	return context.WithValue(ctx, chaveCliente, cliente)
}

// ClienteDoContexto lê o nome do cliente guardado no contexto.
func ClienteDoContexto(ctx context.Context) (string, bool) {
	cliente, ok := ctx.Value(chaveCliente).(string)
	return cliente, ok
}
