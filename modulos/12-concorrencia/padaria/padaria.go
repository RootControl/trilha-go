// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria assa vários pães ao mesmo tempo, e o estoque
// aprende a sobreviver a isso.
package padaria

import (
	"errors"
	"sync"
	"time"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// Erros sentinela da padaria.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")
)

// ----------------------------------------------------------------------
// O forno.
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

// Forno assa uma fornada. Uma interface de um método, de novo.
type Forno interface {
	Assar(pedido PedidoDeFornada) Fornada
}

// FornoLento demora um pouco para assar cada fornada.
//
// Ele existe para o paralelismo ficar visível: com um forno instantâneo, não
// haveria diferença mensurável entre assar em sequência e assar junto.
type FornoLento struct {
	Demora time.Duration
}

// Assar espera a demora configurada e devolve a fornada.
func (f FornoLento) Assar(pedido PedidoDeFornada) Fornada {
	time.Sleep(f.Demora)
	return Fornada{Produto: pedido.Produto, Quantidade: pedido.Quantidade}
}

// ----------------------------------------------------------------------
// O estoque do módulo 02, sem proteção nenhuma.
//
// Ele continua aqui de propósito. Foi este mapa que o módulo 10 colocou atrás
// de um servidor HTTP, e é ele que o detector de corrida acusa quando duas
// requisições chegam ao mesmo tempo.
// ----------------------------------------------------------------------

// Estoque diz quantas unidades a padaria tem de cada produto.
//
// NÃO é seguro para uso concorrente. Um map em Go não é: duas goroutines
// escrevendo ao mesmo tempo podem corromper a tabela de hash, e o runtime
// derruba o programa quando percebe.
type Estoque map[string]int

// NovoEstoque devolve um Estoque pronto para receber escrita.
func NovoEstoque() Estoque { return make(Estoque) }

// Repor soma a quantidade ao que já existe.
func (e Estoque) Repor(nome string, quantidade int) { e[nome] += quantidade }

// Quantidade devolve quantas unidades existem do produto.
func (e Estoque) Quantidade(nome string) int { return e[nome] }

// ----------------------------------------------------------------------
// Módulo 12: é aqui que você trabalha.
// ----------------------------------------------------------------------

// EstoqueSeguro é o Estoque de novo, agora protegido para uso concorrente.
//
// O mutex vem PRIMEIRO na struct e o dado que ele protege vem logo abaixo.
// Não é regra da linguagem, é convenção de leitura: quem abrir o arquivo
// entende na hora o que está protegido pelo quê.
//
// Repare que o campo é um sync.RWMutex, e não um ponteiro. O valor zero de um
// mutex já é um mutex destravado e pronto, então EstoqueSeguro não precisaria
// de construtor se não fosse pelo map lá dentro, que precisa.
//
// E uma regra que não tem exceção: depois que um mutex é usado, a struct que
// o contém NÃO pode mais ser copiada. É por isso que todos os métodos abaixo
// têm receptor de ponteiro, inclusive os que só leem.
type EstoqueSeguro struct {
	mu          sync.RWMutex
	quantidades map[string]int
}

// NovoEstoqueSeguro devolve um estoque pronto para uso.
func NovoEstoqueSeguro() *EstoqueSeguro {
	// TODO: implemente esta função.
	return nil
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
//
// Escrita: trava exclusiva, com Lock e Unlock. O padrão é travar e adiar o
// destravamento na linha seguinte:
//
//	e.mu.Lock()
//	defer e.mu.Unlock()
func (e *EstoqueSeguro) Repor(nome string, quantidade int) {
	// TODO: implemente este método.
}

// Quantidade devolve quantas unidades existem do produto.
//
// Leitura: trava compartilhada, com RLock e RUnlock. Vários leitores entram
// ao mesmo tempo; um escritor espera todos saírem.
//
// Ler sem travar nada é erro, mesmo que a leitura pareça inofensiva. Uma
// leitura simultânea a uma escrita é uma corrida, e o resultado não é "um
// valor velho": é comportamento indefinido.
func (e *EstoqueSeguro) Quantidade(nome string) int {
	// TODO: implemente este método.
	return 0
}

// Baixar tira do estoque a quantidade vendida, se houver.
//
// A verificação e a subtração precisam acontecer dentro da MESMA trava. Se
// você ler com RLock, soltar, e depois escrever com Lock, duas goroutines
// podem passar pela verificação antes de qualquer uma subtrair, e o estoque
// fica negativo. Esse bug tem nome: verificar-e-agir.
//
// Erros: ErrQuantidadeInvalida, ErrProdutoNaoEncontrado, ErrEstoqueInsuficiente,
// como sempre.
func (e *EstoqueSeguro) Baixar(nome string, quantidade int) error {
	// TODO: implemente este método.
	return nil
}

// AssarTudo assa todos os pedidos ao mesmo tempo e devolve as fornadas NA
// MESMA ORDEM dos pedidos.
//
// Use uma goroutine por pedido e um sync.WaitGroup para esperar todas.
//
// A ordem é o detalhe interessante. A saída de cada goroutine vai para uma
// posição própria de uma fatia já dimensionada com make antes do laço:
//
//	fornadas := make([]Fornada, len(pedidos))
//	... dentro da goroutine i: fornadas[i] = ...
//
// Isso NÃO precisa de mutex, e não é sorte. Cada goroutine escreve num índice
// diferente, ou seja, em memória diferente, e o detector de corrida concorda.
// Um append concorrente na mesma fatia, esse sim, seria corrida.
func AssarTudo(forno Forno, pedidos []PedidoDeFornada) []Fornada {
	// TODO: implemente esta função.
	return nil
}

// AssarEmCanal assa todos os pedidos e entrega cada fornada num canal, à
// medida que ficam prontas.
//
// A função devolve o canal IMEDIATAMENTE, sem esperar nada, e o trabalho
// continua ao fundo. Quem chamou lê com `for fornada := range canal`.
//
// Duas responsabilidades suas:
//
//   - quem escreve no canal é quem o fecha, nunca quem lê;
//   - o canal precisa ser fechado quando a última fornada sair, senão o
//     `for range` de quem chamou espera para sempre.
//
// A forma usual é uma goroutine que espera o WaitGroup e então fecha:
//
//	go func() {
//		grupo.Wait()
//		close(canal)
//	}()
//
// O tipo de retorno é `<-chan Fornada`, um canal somente-leitura. Isso é
// documentação verificada pelo compilador: quem recebe não consegue escrever
// nem fechar por engano.
func AssarEmCanal(forno Forno, pedidos []PedidoDeFornada) <-chan Fornada {
	// TODO: implemente esta função.
	return nil
}

// Coletar lê tudo que sair do canal até ele fechar.
//
// Um `for range` sobre canal termina sozinho quando o canal é fechado. Como a
// ordem de chegada não é previsível, ordene o resultado por nome de produto
// antes de devolver, senão nem dá para testar.
func Coletar(canal <-chan Fornada) []Fornada {
	// TODO: implemente esta função.
	return nil
}

// EsperarComPrazo espera uma fornada sair do canal, desistindo depois do
// prazo. O bool diz se chegou a tempo.
//
// Este é o trabalho do select: esperar em mais de uma coisa ao mesmo tempo e
// seguir com a primeira que acontecer.
//
//	select {
//	case fornada := <-canal:
//		...
//	case <-time.After(prazo):
//		...
//	}
//
// Canal fechado conta como não ter chegado nada: uma recepção de canal
// fechado devolve na hora, com o valor zero.
func EsperarComPrazo(canal <-chan Fornada, prazo time.Duration) (Fornada, bool) {
	// TODO: implemente esta função.
	return Fornada{}, false
}

// ContarPorProduto soma quantas unidades foram assadas de cada produto.
//
// Só para você reparar numa coisa: esta função não tem concorrência nenhuma,
// e é a resposta certa. Nem todo problema perto de goroutines vira goroutine.
func ContarPorProduto(fornadas []Fornada) map[string]int {
	// TODO: implemente esta função.
	return nil
}
