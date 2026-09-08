// Package padaria guarda o código da Padaria do Seu Zé.
//
// Neste módulo a padaria ganha um banco de dados de verdade, e a mesma
// interface do módulo 05 passa a ter duas implementações: uma em memória e
// uma em SQL.
package padaria

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	// Importação em branco: o pacote é carregado só pelo efeito colateral de
	// registrar o driver "sqlite" no database/sql. Nada dele é usado por
	// nome, e por isso o underline. Sem esta linha, sql.Open("sqlite", ...)
	// falharia dizendo que o driver é desconhecido.
	_ "modernc.org/sqlite"
)

// ----------------------------------------------------------------------
// Vindo dos módulos anteriores.
// ----------------------------------------------------------------------

// Produto é uma coisa que a padaria vende.
type Produto struct {
	Nome            string `json:"nome"`
	PrecoEmCentavos int    `json:"preco_em_centavos"`
	Disponivel      bool   `json:"disponivel"`
}

// Relogio tira time.Now de dentro da regra de negócio.
type Relogio interface{ Agora() time.Time }

// RelogioDoSistema é a implementação que vai para produção.
type RelogioDoSistema struct{}

// Agora devolve a hora de verdade.
func (RelogioDoSistema) Agora() time.Time { return time.Now() }

// RelogioFixo devolve sempre o mesmo instante, escolhido pelo teste.
type RelogioFixo struct{ Momento time.Time }

// Agora devolve sempre o mesmo instante.
func (r RelogioFixo) Agora() time.Time { return r.Momento }

// Erros sentinela da padaria.
var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	ErrEstoqueInsuficiente  = errors.New("estoque insuficiente")
	ErrQuantidadeInvalida   = errors.New("quantidade inválida")
	ErrProdutoInvalido      = errors.New("produto inválido")
)

// ----------------------------------------------------------------------
// A interface, agora com contexto.
// ----------------------------------------------------------------------

// RepositorioDeProdutos é onde os produtos da padaria moram.
//
// Comparada com a do módulo 05, ela ganhou um context.Context em cada método.
// Por enquanto trate o contexto como um valor que você recebe e repassa
// adiante: toda chamada ao banco leva um. O que ele faz de verdade, cancelar
// trabalho e impor prazo, é o módulo 13.
//
// A regra: contexto é sempre o primeiro parâmetro, e nunca vai guardado
// dentro de uma struct.
type RepositorioDeProdutos interface {
	Salvar(ctx context.Context, produto Produto) error
	Buscar(ctx context.Context, nome string) (Produto, error)
	Todos(ctx context.Context) ([]Produto, error)
}

var (
	_ RepositorioDeProdutos = (*RepositorioEmMemoria)(nil)
	_ RepositorioDeProdutos = (*RepositorioSQL)(nil)
)

// ----------------------------------------------------------------------
// A implementação em memória, pronta, do módulo 05.
// ----------------------------------------------------------------------

// RepositorioEmMemoria guarda produtos num map.
type RepositorioEmMemoria struct {
	produtos map[string]Produto
}

// NovoRepositorioEmMemoria devolve um repositório pronto para uso.
func NovoRepositorioEmMemoria() *RepositorioEmMemoria {
	return &RepositorioEmMemoria{produtos: make(map[string]Produto)}
}

// Salvar guarda o produto, sobrescrevendo se já existir.
func (r *RepositorioEmMemoria) Salvar(_ context.Context, produto Produto) error {
	if produto.Nome == "" {
		return fmt.Errorf("salvar produto sem nome: %w", ErrProdutoInvalido)
	}
	r.produtos[produto.Nome] = produto
	return nil
}

// Buscar procura um produto pelo nome exato.
func (r *RepositorioEmMemoria) Buscar(_ context.Context, nome string) (Produto, error) {
	produto, existe := r.produtos[nome]
	if !existe {
		return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)
	}
	return produto, nil
}

// Todos devolve os produtos guardados, em ordem alfabética de nome.
func (r *RepositorioEmMemoria) Todos(_ context.Context) ([]Produto, error) {
	nomes := make([]string, 0, len(r.produtos))
	for nome := range r.produtos {
		nomes = append(nomes, nome)
	}
	slices.Sort(nomes)

	todos := make([]Produto, 0, len(nomes))
	for _, nome := range nomes {
		todos = append(todos, r.produtos[nome])
	}
	return todos, nil
}

// ----------------------------------------------------------------------
// O esquema do banco.
// ----------------------------------------------------------------------

// Esquema cria as três tabelas da padaria.
//
// Está num const em vez de num arquivo separado só para o módulo caber num
// arquivo. Em projeto de verdade o esquema mora em arquivos .sql numerados,
// embutidos no binário com //go:embed, e aplicados por uma ferramenta de
// migração que sabe quais já rodaram.
//
// IF NOT EXISTS deixa CriarEsquema ser chamada mais de uma vez sem estragar
// nada, o que é o mínimo que se espera de uma migração.
const Esquema = `
CREATE TABLE IF NOT EXISTS produtos (
	nome              TEXT    PRIMARY KEY,
	preco_em_centavos INTEGER NOT NULL,
	disponivel        INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS estoque (
	produto    TEXT    PRIMARY KEY,
	quantidade INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS pedidos (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	cliente           TEXT    NOT NULL,
	produto           TEXT    NOT NULL,
	quantidade        INTEGER NOT NULL,
	total_em_centavos INTEGER NOT NULL,
	em                TEXT    NOT NULL
);
`

// ----------------------------------------------------------------------
// Exemplo resolvido: um upsert com parâmetro.
// ----------------------------------------------------------------------

// RepositorioSQL guarda os produtos num banco relacional.
//
// O campo é um *sql.DB, e isso não é uma conexão: é um POOL de conexões, que
// pode ser usado por várias goroutines ao mesmo tempo e deve viver enquanto o
// programa viver. Abrir um por requisição é o erro clássico com este pacote.
type RepositorioSQL struct {
	db *sql.DB
}

// Repor soma quantidade ao estoque do produto, criando a linha se preciso.
//
// Já está pronto, e é o modelo dos métodos que você vai escrever.
//
// Repare em duas coisas.
//
// Os valores entram como PARÂMETROS, os pontos de interrogação, e nunca
// grudados no texto da consulta. Isso não é estilo: montar SQL com
// concatenação é como se abre uma injeção de SQL. O banco recebe a consulta e
// os dados por caminhos separados, então um valor nunca vira comando.
//
// ON CONFLICT ... DO UPDATE é o upsert do SQLite e do Postgres. Ele evita a
// corrida entre "verificar se existe" e "inserir", que é um bug real quando
// duas requisições chegam juntas.
func (r *RepositorioSQL) Repor(ctx context.Context, nome string, quantidade int) error {
	const consulta = `
		INSERT INTO estoque (produto, quantidade) VALUES (?, ?)
		ON CONFLICT (produto) DO UPDATE SET quantidade = quantidade + excluded.quantidade`

	if _, err := r.db.ExecContext(ctx, consulta, nome, quantidade); err != nil {
		return fmt.Errorf("repor %q: %w", nome, err)
	}

	return nil
}

// QuantidadeEm devolve quantas unidades existem do produto. Já está pronto.
//
// Repare no tratamento de sql.ErrNoRows: para o banco, "nenhuma linha" é um
// erro; para a padaria, é quantidade zero. Traduzir isso é trabalho do
// repositório, e é o que impede o resto do programa de saber o que é SQL.
func (r *RepositorioSQL) QuantidadeEm(ctx context.Context, nome string) (int, error) {
	const consulta = `SELECT quantidade FROM estoque WHERE produto = ?`

	var quantidade int
	err := r.db.QueryRowContext(ctx, consulta, nome).Scan(&quantidade)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, nil
	case err != nil:
		return 0, fmt.Errorf("consultar estoque de %q: %w", nome, err)
	}

	return quantidade, nil
}

// ----------------------------------------------------------------------
// Módulo 11: é aqui que você trabalha.
// ----------------------------------------------------------------------

// CriarEsquema aplica o Esquema no banco.
//
// Uma chamada a ExecContext com a constante Esquema, e o erro envolvido com
// contexto. Três linhas.
func CriarEsquema(ctx context.Context, db *sql.DB) error {
	// TODO: implemente esta função.
	return nil
}

// NovoRepositorioSQL devolve um repositório apoiado no banco recebido.
//
// Ele NÃO abre o banco e NÃO cria o esquema. Receber o *sql.DB de fora é o que
// permite ao teste entregar um banco descartável e ao programa entregar o de
// produção, exatamente como o relógio do módulo 06.
func NovoRepositorioSQL(db *sql.DB) *RepositorioSQL {
	// TODO: implemente esta função.
	return nil
}

// Salvar guarda o produto, sobrescrevendo se já existir um com o mesmo nome.
//
// Produto sem nome é recusado com ErrProdutoInvalido, antes de tocar no banco.
// Use o mesmo upsert do Repor, mas substituindo os valores em vez de somar.
func (r *RepositorioSQL) Salvar(ctx context.Context, produto Produto) error {
	// TODO: implemente este método.
	return nil
}

// Buscar procura um produto pelo nome exato.
//
// QueryRowContext seguido de Scan. Quando não há linha, o Scan devolve
// sql.ErrNoRows, e o seu trabalho é traduzir isso para
// ErrProdutoNaoEncontrado, envolvido com o nome procurado.
func (r *RepositorioSQL) Buscar(ctx context.Context, nome string) (Produto, error) {
	// TODO: implemente este método.
	return Produto{}, nil
}

// Todos devolve os produtos guardados, em ordem alfabética de nome.
//
// O laço de leitura tem quatro obrigações, e a terceira é a mais esquecida:
//
//  1. defer rows.Close()
//  2. for rows.Next() com rows.Scan dentro
//  3. rows.Err() DEPOIS do laço, porque o Next devolve false tanto no fim
//     normal quanto quando a conexão cai no meio da leitura
//  4. envolver os erros com contexto
//
// A ordenação vem do banco, com ORDER BY, e não de um slices.Sort depois.
func (r *RepositorioSQL) Todos(ctx context.Context) ([]Produto, error) {
	// TODO: implemente este método.
	return nil, nil
}

// RegistrarVenda dá baixa no estoque e grava o pedido, tudo ou nada.
//
// Este é o método que justifica o módulo. Ele precisa de uma TRANSAÇÃO,
// porque baixar o estoque sem gravar o pedido, ou gravar o pedido sem baixar o
// estoque, deixaria o banco mentindo.
//
// A forma:
//
//	tx, err := r.db.BeginTx(ctx, nil)
//	if err != nil { ... }
//	defer tx.Rollback() // vira operação vazia depois de um Commit bem-sucedido
//	... use tx.QueryRowContext e tx.ExecContext, nunca r.db ...
//	return tx.Commit()
//
// O que precisa acontecer dentro dela:
//
//  1. quantidade menor ou igual a zero devolve ErrQuantidadeInvalida
//  2. buscar preço do produto; sem linha, ErrProdutoNaoEncontrado
//  3. buscar quantidade em estoque; sem linha ou menor que a pedida,
//     ErrEstoqueInsuficiente
//  4. UPDATE no estoque
//  5. INSERT no pedido, com a hora vinda de relogio.Agora() formatada em
//     RFC3339
//
// Devolve o total em centavos.
func (r *RepositorioSQL) RegistrarVenda(ctx context.Context, relogio Relogio, cliente, nome string, quantidade int) (int, error) {
	// TODO: implemente este método.
	return 0, nil
}
