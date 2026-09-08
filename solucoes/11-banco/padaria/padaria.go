// Package padaria guarda o código da Padaria do Seu Zé.
//
// Esta é a solução do módulo 11. Olhe depois de tentar, não antes.
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
// Módulo 11.
// ----------------------------------------------------------------------

// CriarEsquema aplica o Esquema no banco.
func CriarEsquema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, Esquema); err != nil {
		return fmt.Errorf("criar esquema: %w", err)
	}

	return nil
}

// NovoRepositorioSQL devolve um repositório apoiado no banco recebido.
func NovoRepositorioSQL(db *sql.DB) *RepositorioSQL {
	return &RepositorioSQL{db: db}
}

// Salvar guarda o produto, sobrescrevendo se já existir um com o mesmo nome.
func (r *RepositorioSQL) Salvar(ctx context.Context, produto Produto) error {
	if produto.Nome == "" {
		// Validação antes de falar com o banco. Ida ao banco é cara, e o erro
		// que o banco daria seria menos específico que este.
		return fmt.Errorf("salvar produto sem nome: %w", ErrProdutoInvalido)
	}

	const consulta = `
		INSERT INTO produtos (nome, preco_em_centavos, disponivel) VALUES (?, ?, ?)
		ON CONFLICT (nome) DO UPDATE SET
			preco_em_centavos = excluded.preco_em_centavos,
			disponivel        = excluded.disponivel`

	_, err := r.db.ExecContext(ctx, consulta, produto.Nome, produto.PrecoEmCentavos, produto.Disponivel)
	if err != nil {
		return fmt.Errorf("salvar %q: %w", produto.Nome, err)
	}

	return nil
}

// Buscar procura um produto pelo nome exato.
func (r *RepositorioSQL) Buscar(ctx context.Context, nome string) (Produto, error) {
	const consulta = `
		SELECT nome, preco_em_centavos, disponivel
		FROM produtos
		WHERE nome = ?`

	var produto Produto
	err := r.db.QueryRowContext(ctx, consulta, nome).
		Scan(&produto.Nome, &produto.PrecoEmCentavos, &produto.Disponivel)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// A tradução que impede o resto do programa de saber o que é SQL.
		// Quem chama já sabe lidar com ErrProdutoNaoEncontrado desde o
		// módulo 03, e não precisa aprender um erro novo por causa do banco.
		return Produto{}, fmt.Errorf("buscar %q: %w", nome, ErrProdutoNaoEncontrado)

	case err != nil:
		return Produto{}, fmt.Errorf("buscar %q: %w", nome, err)
	}

	return produto, nil
}

// Todos devolve os produtos guardados, em ordem alfabética de nome.
func (r *RepositorioSQL) Todos(ctx context.Context) ([]Produto, error) {
	const consulta = `
		SELECT nome, preco_em_centavos, disponivel
		FROM produtos
		ORDER BY nome`

	linhas, err := r.db.QueryContext(ctx, consulta)
	if err != nil {
		return nil, fmt.Errorf("listar produtos: %w", err)
	}
	// Sem este Close, a conexão volta para o pool só quando o coletor de lixo
	// resolver, e um serviço movimentado esgota o pool e trava.
	defer linhas.Close()

	todos := []Produto{}
	for linhas.Next() {
		var produto Produto
		if err := linhas.Scan(&produto.Nome, &produto.PrecoEmCentavos, &produto.Disponivel); err != nil {
			return nil, fmt.Errorf("ler produto: %w", err)
		}
		todos = append(todos, produto)
	}

	// A verificação mais esquecida do database/sql. Next devolve false tanto
	// no fim normal quanto quando a conexão morre no meio, e sem este Err o
	// segundo caso vira uma lista curta silenciosa.
	if err := linhas.Err(); err != nil {
		return nil, fmt.Errorf("listar produtos: %w", err)
	}

	return todos, nil
}

// RegistrarVenda dá baixa no estoque e grava o pedido, tudo ou nada.
func (r *RepositorioSQL) RegistrarVenda(ctx context.Context, relogio Relogio, cliente, nome string, quantidade int) (int, error) {
	if quantidade <= 0 {
		return 0, fmt.Errorf("vender %d de %q: %w", quantidade, nome, ErrQuantidadeInvalida)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("abrir transação: %w", err)
	}
	// Depois de um Commit bem-sucedido, este Rollback vira operação vazia e
	// devolve sql.ErrTxDone, que é justamente o que se descarta aqui. Em
	// qualquer outro caminho, inclusive um pânico, ele desfaz tudo.
	defer func() { _ = tx.Rollback() }()

	var precoEmCentavos int
	err = tx.QueryRowContext(ctx, `SELECT preco_em_centavos FROM produtos WHERE nome = ?`, nome).
		Scan(&precoEmCentavos)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("vender %q: %w", nome, ErrProdutoNaoEncontrado)
	case err != nil:
		return 0, fmt.Errorf("consultar produto %q: %w", nome, err)
	}

	var disponivel int
	err = tx.QueryRowContext(ctx, `SELECT quantidade FROM estoque WHERE produto = ?`, nome).
		Scan(&disponivel)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		disponivel = 0
	case err != nil:
		return 0, fmt.Errorf("consultar estoque de %q: %w", nome, err)
	}

	if disponivel < quantidade {
		return 0, fmt.Errorf("vender %d de %q, só tem %d: %w",
			quantidade, nome, disponivel, ErrEstoqueInsuficiente)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE estoque SET quantidade = quantidade - ? WHERE produto = ?`,
		quantidade, nome); err != nil {
		return 0, fmt.Errorf("baixar estoque de %q: %w", nome, err)
	}

	total := precoEmCentavos * quantidade

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO pedidos (cliente, produto, quantidade, total_em_centavos, em)
		 VALUES (?, ?, ?, ?, ?)`,
		cliente, nome, quantidade, total, relogio.Agora().Format(time.RFC3339)); err != nil {
		return 0, fmt.Errorf("gravar pedido: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("confirmar venda: %w", err)
	}

	return total, nil
}
