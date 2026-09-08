# Módulo 11 — Banco de dados

A padaria ganha um banco de verdade. E, pela primeira vez no curso, uma
dependência externa.

```bash
go test ./modulos/11-banco/...
```

## A primeira dependência

Dez módulos com a biblioteca padrão e nada mais. Isso não é purismo: é que a
biblioteca padrão de Go cobre muito mais do que a de outras linguagens, e a
cultura da comunidade é acrescentar dependência com desconfiança.

Agora ela é inevitável, porque `database/sql` **não fala com banco nenhum**.
Ele é uma camada genérica; quem sabe conversar com um banco específico é um
driver, e driver não vem na biblioteca padrão.

O escolhido aqui é `modernc.org/sqlite`, um SQLite escrito em Go puro. A
escolha tem um motivo prático: você não precisa instalar servidor de banco
nenhum, nem Docker, para fazer este módulo, e o CI também não. Conceitualmente
tudo aqui vale igual para Postgres, e a seção final diz o que muda.

Repare no `go.mod` da raiz. Ele tem uma dependência direta e uma lista de
indiretas, que são as dependências das dependências. O `go.sum` guarda a
impressão digital criptográfica de cada uma: se alguém trocar o conteúdo de
uma versão já publicada, o build quebra em vez de rodar código diferente.

## Por que Go faz assim

**Importação em branco.**

```go
import _ "modernc.org/sqlite"
```

O underline importa o pacote **só pelo efeito colateral**. O driver, ao ser
carregado, se registra no `database/sql` com o nome `sqlite`. Nada dele é
usado por nome, e sem o underline o compilador recusaria o import não usado,
que é a regra do módulo 00 voltando.

É um padrão que divide opiniões, e é bom saber que ele existe: importar um
pacote pode executar código.

**`sql.Open` não abre nada.**
Ela valida os argumentos e monta o pool. A primeira conexão de verdade só
acontece no primeiro uso. Por isso o auxiliar de teste chama `PingContext`:
para descobrir agora que o caminho está errado, e não três funções depois.

**`*sql.DB` é um pool, não uma conexão.**
Ele é seguro para uso por várias goroutines ao mesmo tempo e deve viver
enquanto o programa viver. Abrir um por requisição HTTP é o erro clássico com
este pacote, e produz um serviço que esgota conexões sob carga. Guarde um só,
e ajuste com `SetMaxOpenConns`, `SetMaxIdleConns` e `SetConnMaxLifetime`.

**Contexto em toda chamada.**
`QueryContext`, `ExecContext`, `QueryRowContext`, `BeginTx`. Existem versões
sem contexto, e elas são resquício histórico: não use. Neste módulo o contexto
é só um valor que você recebe e repassa. O que ele faz de verdade, cancelar
trabalho quando o cliente desiste e impor prazo, é o módulo 13.

Duas regras que valem desde já: contexto é sempre o **primeiro** parâmetro, e
nunca vai guardado dentro de uma struct.

**Parâmetro, nunca concatenação.**
Os pontos de interrogação na consulta não são estilo. O banco recebe o
comando e os dados por caminhos separados, então um valor jamais vira comando.
`TestDemonstracaoInjecaoDeSQLNaoFunciona` manda um nome de produto contendo
`'; DROP TABLE produtos; --` e mostra a tabela intacta depois.

A regra é absoluta: se um valor veio de fora, ele entra como parâmetro. Não
existe caso em que concatenar é aceitável porque "esse campo é só um número".

**`sql.ErrNoRows` para de existir na fronteira do repositório.**
Para o banco, não achar linha é um erro. Para a padaria, é
`ErrProdutoNaoEncontrado`, que o resto do programa já sabe tratar desde o
módulo 03. Traduzir isso é trabalho do repositório, e é o que impede o
`net/http`, a linha de comando e os testes de precisarem saber o que é SQL.

Um dos testes de contrato verifica exatamente isso: ele falha se
`sql.ErrNoRows` vazar para quem chamou.

**A mesma bateria de testes, duas implementações.**
`TestContratoDoRepositorio` roda os mesmos seis casos contra o repositório em
memória e contra o SQL. Por dentro não têm nada em comum, um é um `map` e o
outro é um banco, e passam nos mesmos testes porque satisfazem a mesma
interface.

Isso se chama teste de contrato, e é a colheita do investimento feito no
módulo 05. Vale sempre que existir mais de uma implementação de uma interface:
você escreve a bateria uma vez e ganha a garantia de que as duas se comportam
igual.

## Transação: o assunto do módulo

`RegistrarVenda` faz duas escritas: baixa o estoque e grava o pedido. Se a
primeira acontecer e a segunda não, o banco passa a mentir, e ninguém percebe.

```go
tx, err := r.db.BeginTx(ctx, nil)
if err != nil { ... }
defer func() { _ = tx.Rollback() }()
// ... tudo com tx, nunca com r.db ...
return tx.Commit()
```

Três coisas para guardar.

O `defer` do rollback vem **antes** de qualquer trabalho, e por isso cobre
todo caminho de saída, inclusive um pânico. Depois de um `Commit` bem-sucedido
ele vira operação vazia e devolve `sql.ErrTxDone`, que é justamente o erro que
se descarta ali.

Dentro da transação, tudo passa por `tx`. Usar `r.db` no meio pega **outra
conexão do pool**, que está fora da transação, e o resultado é uma escrita que
não é desfeita no rollback. É um bug difícil de enxergar lendo o código.

E a transação existe para o caminho de erro, não para o de sucesso.
`TestRegistrarVendaRecusaEDesfaz` roda quatro formas de recusar e confere, em
cada uma, que o estoque não mudou e que nenhum pedido foi gravado.

## Pegadinhas

**`:memory:` com pool dá um banco por conexão.**
Esta custa muitas horas para quem não sabe. Um banco SQLite em memória
pertence à conexão que o criou. Como `*sql.DB` é um pool, a segunda consulta
pode pegar outra conexão, e ela enxerga um banco vazio: a tabela que você
acabou de criar simplesmente não existe. Os sintomas são intermitentes, o que
piora tudo.

As saídas são usar um arquivo, que é o que o teste deste módulo faz com
`t.TempDir()`, ou limitar o pool a uma conexão com `db.SetMaxOpenConns(1)`,
ou usar cache compartilhado no DSN. Arquivo temporário é o mais simples e o
mais parecido com produção.

**Esquecer `rows.Err()`.**
A verificação mais esquecida do `database/sql`. O `rows.Next()` devolve
`false` tanto no fim normal quanto quando a conexão morre no meio da leitura.
Sem o `rows.Err()` depois do laço, o segundo caso vira uma lista curta,
silenciosa, sem erro nenhum.

**Esquecer `rows.Close()`.**
Sem ele a conexão só volta para o pool quando o coletor de lixo resolver. Num
serviço movimentado, isso esgota o pool e trava tudo. `defer rows.Close()` na
linha seguinte ao `Query`, sempre.

**Marcador de parâmetro muda de banco para banco.**
SQLite e MySQL usam `?`. Postgres usa `$1`, `$2`. Oracle usa `:nome`. É o
principal motivo de uma consulta escrita para um banco não rodar noutro sem
edição, e uma das razões de existirem construtores de consulta.

**`sql.Open` não valida a conexão, e `defer db.Close()` no `main` é raro.**
O pool deve viver o tempo do programa. Fechar faz sentido em teste, com
`t.Cleanup`, e num desligamento ordenado, que é o módulo 13.

**Boolean em SQLite é inteiro.**
Não existe tipo booleano; o `database/sql` converte `bool` para 0 e 1 na ida e
de volta na volta. Funciona, e é bom saber que é conversão e não tipo nativo,
porque uma consulta escrita à mão no terminal vai mostrar `0` e `1`.

## O exercício

Seis implementações em `padaria/padaria.go`:

1. `CriarEsquema`, três linhas.
2. `NovoRepositorioSQL`, que recebe o banco de fora em vez de abri-lo.
3. `Salvar`, um upsert espelhado no `Repor` que já vem pronto.
4. `Buscar`, com a tradução do `sql.ErrNoRows`.
5. `Todos`, com o laço de leitura completo, incluindo o `rows.Err()`.
6. `RegistrarVenda`, a transação.

`Repor` e `QuantidadeEm` já estão implementados no arquivo, e servem de
modelo: um mostra o upsert com parâmetros, o outro mostra a tradução de
`sql.ErrNoRows` para o vocabulário da padaria.

Comece por `CriarEsquema` e `NovoRepositorioSQL`, porque sem elas o auxiliar
de teste nem monta o cenário e todos os testes falham pelo mesmo motivo.

## O que muda para Postgres

Menos do que parece, e o suficiente para incomodar:

| | SQLite | Postgres |
|---|---|---|
| driver | `modernc.org/sqlite` | `github.com/jackc/pgx/v5` |
| nome no `sql.Open` | `sqlite` | `pgx` |
| parâmetro | `?` | `$1`, `$2` |
| chave crescente | `INTEGER PRIMARY KEY AUTOINCREMENT` | `BIGSERIAL` ou `IDENTITY` |
| data e hora | texto, você formata | `TIMESTAMPTZ` nativo |
| concorrência | um escritor por vez | vários, com bloqueio por linha |

O que **não** muda é tudo que este módulo ensina: pool, contexto, parâmetro,
tradução de erro, laço de leitura e transação.

## Desafio, sem teste pronto

**Aponte o servidor do módulo 10 para o banco.** O `*Servidor` recebe um
`RepositorioDeProdutos`, e agora existem duas implementações. Troque a
implementação e repare quantas linhas dos handlers precisam mudar.

**Escreva uma migração de verdade.** Uma tabela `migracoes` guardando quais
arquivos já rodaram, e uma função que aplica só os que faltam, cada um dentro
da sua própria transação.

**Meça a diferença.** Insira mil produtos em um laço com mil `ExecContext`, e
depois os mesmos mil dentro de uma transação só. Cronometre os dois. A
diferença costuma ser de mais de uma ordem de grandeza, e entender por quê
explica boa parte do que um banco faz por baixo.

## Se travar

A solução está em `solucoes/11-banco/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
