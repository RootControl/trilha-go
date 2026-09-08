// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 12. Olhe depois de tentar, não antes.
package padaria

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
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
// Módulo 12.
// ----------------------------------------------------------------------

// EstoqueSeguro é o Estoque protegido para uso concorrente.
type EstoqueSeguro struct {
	mu          sync.RWMutex
	quantidades map[string]int
}

// NovoEstoqueSeguro devolve um estoque pronto para uso.
func NovoEstoqueSeguro() *EstoqueSeguro {
	return &EstoqueSeguro{quantidades: make(map[string]int)}
}

// Repor soma a quantidade ao que já existe no estoque daquele produto.
func (e *EstoqueSeguro) Repor(nome string, quantidade int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.quantidades[nome] += quantidade
}

// Quantidade devolve quantas unidades existem do produto.
func (e *EstoqueSeguro) Quantidade(nome string) int {
	// RLock deixa vários leitores entrarem ao mesmo tempo. Só um escritor
	// bloqueia todo mundo.
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.quantidades[nome]
}

// Baixar tira do estoque a quantidade vendida, se houver.
func (e *EstoqueSeguro) Baixar(nome string, quantidade int) error {
	// A validação que não toca no estado fica FORA da trava. Segurar um mutex
	// mais tempo do que o necessário é como se transforma um programa
	// concorrente num programa sequencial lento.
	if quantidade <= 0 {
		return fmt.Errorf("baixar %d de %q: %w", quantidade, nome, ErrQuantidadeInvalida)
	}

	// Uma trava só, cobrindo a verificação E a subtração. Este é o ponto do
	// método: com duas travas separadas, duas goroutines passam pela
	// verificação antes de qualquer uma subtrair, e o estoque fica negativo.
	e.mu.Lock()
	defer e.mu.Unlock()

	atual, existe := e.quantidades[nome]
	if !existe {
		return fmt.Errorf("baixar %q: %w", nome, ErrProdutoNaoEncontrado)
	}
	if atual < quantidade {
		return fmt.Errorf("baixar %d de %q, só tem %d: %w",
			quantidade, nome, atual, ErrEstoqueInsuficiente)
	}

	e.quantidades[nome] = atual - quantidade

	return nil
}

// AssarTudo assa todos os pedidos ao mesmo tempo, preservando a ordem.
func AssarTudo(forno Forno, pedidos []PedidoDeFornada) []Fornada {
	// A fatia é dimensionada ANTES do laço, e cada goroutine escreve num
	// índice próprio. Índices diferentes são endereços diferentes, então não
	// há corrida e não é preciso mutex. Um append concorrente aqui, sim,
	// seria corrida.
	fornadas := make([]Fornada, len(pedidos))

	var grupo sync.WaitGroup

	for i, pedido := range pedidos {
		// Add vem antes do go, sempre. Chamar Add lá dentro é uma corrida
		// com o Wait, e o Wait pode voltar antes de a goroutine existir.
		grupo.Add(1)

		go func() {
			// Done adiado: roda mesmo se o corpo entrar em pânico.
			defer grupo.Done()

			// i e pedido são novos a cada volta desde o Go 1.22. Em versões
			// anteriores, todas as goroutines veriam o último valor, e isso
			// era a pegadinha mais famosa da linguagem.
			fornadas[i] = forno.Assar(pedido)
		}()
	}

	grupo.Wait()

	return fornadas
}

// AssarEmCanal assa todos os pedidos e entrega cada fornada num canal.
func AssarEmCanal(forno Forno, pedidos []PedidoDeFornada) <-chan Fornada {
	canal := make(chan Fornada)

	var grupo sync.WaitGroup

	for _, pedido := range pedidos {
		grupo.Add(1)

		go func() {
			defer grupo.Done()

			// Canal sem buffer: este envio fica parado até alguém receber.
			// É por isso que quem chama PRECISA ler o canal, senão estas
			// goroutines vazam.
			canal <- forno.Assar(pedido)
		}()
	}

	// Uma goroutine só para esperar e fechar. Sem ela, ou o Wait travaria a
	// função antes de devolver o canal, ou o canal nunca fecharia e o for
	// range de quem chamou esperaria para sempre.
	go func() {
		grupo.Wait()
		close(canal)
	}()

	return canal
}

// Coletar lê tudo que sair do canal até ele fechar.
func Coletar(canal <-chan Fornada) []Fornada {
	var fornadas []Fornada

	// O for range sobre canal termina sozinho quando o canal fecha.
	for fornada := range canal {
		fornadas = append(fornadas, fornada)
	}

	// A ordem de chegada depende do escalonador e muda a cada execução.
	// Ordenar é o que torna o resultado testável.
	slices.SortFunc(fornadas, func(a, b Fornada) int {
		if c := cmp.Compare(a.Produto, b.Produto); c != 0 {
			return c
		}
		return cmp.Compare(a.Quantidade, b.Quantidade)
	})

	return fornadas
}

// EsperarComPrazo espera uma fornada, desistindo depois do prazo.
func EsperarComPrazo(canal <-chan Fornada, prazo time.Duration) (Fornada, bool) {
	select {
	case fornada, aberto := <-canal:
		// Canal fechado devolve na hora, com o valor zero e aberto igual a
		// false. Sem esta verificação, um canal fechado pareceria uma fornada
		// vazia que chegou a tempo.
		if !aberto {
			return Fornada{}, false
		}
		return fornada, true

	case <-time.After(prazo):
		return Fornada{}, false
	}
}

// ContarPorProduto soma quantas unidades foram assadas de cada produto.
func ContarPorProduto(fornadas []Fornada) map[string]int {
	total := make(map[string]int)

	for _, fornada := range fornadas {
		total[fornada.Produto] += fornada.Quantidade
	}

	return total
}
