// Package padaria guarda o código da Padaria do Seu Zé.
//
// No módulo 12 a padaria aprendeu a assar em paralelo. Aqui ela aprende a
// PARAR: a desistir de um trabalho que demorou demais, e a fechar a loja sem
// jogar fora o pão que já está no forno.
package padaria

import (
	"context"
	"errors"
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
// Módulo 13: é aqui que você trabalha.
// ----------------------------------------------------------------------

// Assar espera a demora configurada, mas desiste se o contexto for cancelado.
//
// Este método é o coração do módulo. Um time.Sleep comum é surdo: uma vez
// começado, nada o interrompe, e quem chamou fica preso até o fim mesmo que já
// não queira mais o resultado. A forma de tornar uma espera cancelável é
// esperar em duas coisas ao mesmo tempo, com o select do módulo 12:
//
//	temporizador := time.NewTimer(f.Demora)
//	defer temporizador.Stop()
//
//	select {
//	case <-temporizador.C:
//		...assou...
//	case <-ctx.Done():
//		...desistiu...
//	}
//
// Cancelado, devolva a fornada zerada e ctx.Err(), que já diz se foi
// cancelamento ou prazo estourado.
//
// Use time.NewTimer com Stop adiado, e não time.After. O README explica.
func (f FornoLento) Assar(ctx context.Context, pedido PedidoDeFornada) (Fornada, error) {
	// TODO: implemente este método.
	return Fornada{}, nil
}

// AssarComPrazo assa dando no máximo o prazo indicado.
//
// Monte um contexto filho com context.WithTimeout a partir do que você
// recebeu, e passe o filho para o forno.
//
// A função WithTimeout devolve DOIS valores, e o segundo é uma função de
// cancelamento que você é OBRIGADO a chamar, sempre, mesmo no caminho de
// sucesso. Ela libera os recursos do temporizador interno. A forma é:
//
//	ctxComPrazo, cancelar := context.WithTimeout(ctx, prazo)
//	defer cancelar()
//
// Esquecer esse defer é vazamento de memória, e o go vet reclama.
func AssarComPrazo(ctx context.Context, forno Forno, pedido PedidoDeFornada, prazo time.Duration) (Fornada, error) {
	// TODO: implemente esta função.
	return Fornada{}, nil
}

// ProcessarFila assa todos os pedidos com um número fixo de trabalhadores.
//
// Diferente do AssarTudo do módulo 12, que disparava uma goroutine por
// pedido, aqui o paralelismo é LIMITADO: dez mil pedidos com três
// trabalhadores usam três goroutines, não dez mil. É assim que se protege um
// recurso escasso, seja um forno, uma conexão de banco ou uma API de
// terceiro com limite de chamadas.
//
// A forma clássica, e é ela que você vai escrever:
//
//  1. um canal de tarefas, onde cada tarefa carrega o pedido E o índice dele;
//  2. N goroutinas iguais lendo desse canal com for range;
//  3. uma goroutine alimentadora que enfileira tudo e fecha o canal no fim;
//  4. um WaitGroup esperando os trabalhadores;
//  5. as fornadas numa fatia dimensionada por índice, como no módulo 12.
//
// O contexto atravessa tudo. Se ele for cancelado, ou se algum forno falhar,
// a fila para de aceitar trabalho novo e a função devolve o erro. Para isso,
// derive um contexto cancelável seu:
//
//	ctx, cancelar := context.WithCancel(ctx)
//	defer cancelar()
//
// Assim você consegue mandar todo mundo parar sem mexer no contexto de quem
// chamou.
//
// Regras da saída: trabalhadores menor ou igual a zero devolve
// ErrTrabalhadoresInvalido. Lista vazia devolve fatia vazia e nil. Em caso de
// erro, devolva nil e o erro.
func ProcessarFila(ctx context.Context, forno Forno, trabalhadores int, pedidos []PedidoDeFornada) ([]Fornada, error) {
	// TODO: implemente esta função.
	return nil, nil
}

// chaveDeContexto é o tipo das chaves que esta padaria guarda em contexto.
//
// O tipo é PRIVADO de propósito, e isso não é preciosismo. Um contexto
// atravessa pacotes que não se conhecem, e se a chave fosse a string
// "cliente", qualquer outro pacote poderia usar a mesma string e sobrescrever
// o seu valor sem ninguém perceber. Com um tipo próprio e não exportado, a
// colisão é impossível: só este pacote consegue construir uma chave destas.
type chaveDeContexto string

const chaveCliente chaveDeContexto = "cliente"

// ContextoComCliente devolve um contexto filho carregando o nome do cliente.
//
// Use context.WithValue.
func ContextoComCliente(ctx context.Context, cliente string) context.Context {
	// TODO: implemente esta função.
	return ctx
}

// ClienteDoContexto lê o nome do cliente guardado no contexto.
//
// O bool diz se havia algum. Lembre que ctx.Value devolve um any, então
// precisa de asserção de tipo, e ela também tem a forma com dois retornos:
//
//	cliente, ok := ctx.Value(chaveCliente).(string)
func ClienteDoContexto(ctx context.Context) (string, bool) {
	// TODO: implemente esta função.
	return "", false
}
