package padaria

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var momentoDeTeste = time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)

// bancoDeTeste abre um banco descartável, cria o esquema e fecha no fim.
//
// O caminho vem de t.TempDir, e não de ":memory:". O motivo está no README, e
// é uma das armadilhas mais caras de tempo com database/sql.
func bancoDeTeste(t *testing.T) *sql.DB {
	t.Helper()

	caminho := filepath.Join(t.TempDir(), "padaria.db")

	db, err := sql.Open("sqlite", caminho)
	if err != nil {
		t.Fatalf("abrir banco: %v", err)
	}
	// t.Cleanup roda no fim do teste mesmo se ele falhar, e é o lugar certo
	// para desfazer o que o auxiliar montou.
	t.Cleanup(func() { db.Close() })

	// sql.Open NÃO conecta: ele só valida os argumentos e prepara o pool. Se
	// o caminho estiver errado, você só descobre no primeiro uso. Ping é como
	// se descobre cedo.
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("conectar no banco: %v", err)
	}

	if err := CriarEsquema(context.Background(), db); err != nil {
		t.Fatalf("criar esquema: %v", err)
	}

	return db
}

func repositorioSQLDeTeste(t *testing.T) *RepositorioSQL {
	t.Helper()

	repositorio := NovoRepositorioSQL(bancoDeTeste(t))
	if repositorio == nil {
		t.Fatal("NovoRepositorioSQL devolveu nil")
	}

	return repositorio
}

// ----------------------------------------------------------------------
// Teste de contrato: a MESMA bateria contra as duas implementações.
// ----------------------------------------------------------------------

func TestContratoDoRepositorio(t *testing.T) {
	// Esta é a colheita do módulo 05. As duas implementações não têm nada em
	// comum por dentro, uma é um map e a outra é SQL, e passam pelos mesmos
	// testes porque satisfazem a mesma interface.
	implementacoes := map[string]func(t *testing.T) RepositorioDeProdutos{
		"em memória": func(t *testing.T) RepositorioDeProdutos {
			return NovoRepositorioEmMemoria()
		},
		"SQL": func(t *testing.T) RepositorioDeProdutos {
			return repositorioSQLDeTeste(t)
		},
	}

	for nome, montar := range implementacoes {
		t.Run(nome, func(t *testing.T) {
			ctx := context.Background()

			t.Run("salvar e buscar", func(t *testing.T) {
				repositorio := montar(t)

				broa := Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}
				if err := repositorio.Salvar(ctx, broa); err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}

				recebido, err := repositorio.Buscar(ctx, "Broa")
				if err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}
				if recebido != broa {
					t.Errorf("Buscar() = %+v, esperado %+v", recebido, broa)
				}
			})

			t.Run("salvar duas vezes sobrescreve", func(t *testing.T) {
				repositorio := montar(t)

				if err := repositorio.Salvar(ctx, Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}); err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}
				if err := repositorio.Salvar(ctx, Produto{Nome: "Broa", PrecoEmCentavos: 350, Disponivel: false}); err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}

				recebido, err := repositorio.Buscar(ctx, "Broa")
				if err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}
				if recebido.PrecoEmCentavos != 350 {
					t.Errorf("preço = %d, esperado 350", recebido.PrecoEmCentavos)
				}
				// O booleano precisa ter sobrevivido à ida e à volta.
				if recebido.Disponivel {
					t.Error("Disponivel = true, esperado false")
				}

				todos, err := repositorio.Todos(ctx)
				if err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}
				if len(todos) != 1 {
					t.Errorf("vieram %d produtos, esperado 1: sobrescrever não pode duplicar", len(todos))
				}
			})

			t.Run("buscar o que não existe", func(t *testing.T) {
				repositorio := montar(t)

				_, err := repositorio.Buscar(ctx, "Croissant")

				if err == nil {
					t.Fatal("esperava erro")
				}
				if !errors.Is(err, ErrProdutoNaoEncontrado) {
					t.Errorf("errors.Is não achou ErrProdutoNaoEncontrado em %v", err)
				}
				if !strings.Contains(err.Error(), "Croissant") {
					t.Errorf("a mensagem deveria citar o produto, veio %q", err.Error())
				}
				// O erro do banco não pode vazar para quem chamou.
				if errors.Is(err, sql.ErrNoRows) {
					t.Error("sql.ErrNoRows vazou do repositório: traduza para o erro da padaria")
				}
			})

			t.Run("salvar sem nome", func(t *testing.T) {
				repositorio := montar(t)

				err := repositorio.Salvar(ctx, Produto{PrecoEmCentavos: 100})

				if !errors.Is(err, ErrProdutoInvalido) {
					t.Errorf("errors.Is não achou ErrProdutoInvalido em %v", err)
				}
			})

			t.Run("todos vem em ordem alfabética", func(t *testing.T) {
				repositorio := montar(t)

				for _, produto := range []Produto{
					{Nome: "Sonho", PrecoEmCentavos: 700, Disponivel: false},
					{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
					{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
				} {
					if err := repositorio.Salvar(ctx, produto); err != nil {
						t.Fatalf("erro inesperado: %v", err)
					}
				}

				todos, err := repositorio.Todos(ctx)
				if err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}

				esperado := []string{"Broa", "Pão francês", "Sonho"}
				if len(todos) != len(esperado) {
					t.Fatalf("vieram %d produtos, esperado %d", len(todos), len(esperado))
				}
				for i, nome := range esperado {
					if todos[i].Nome != nome {
						t.Errorf("posição %d = %q, esperado %q", i, todos[i].Nome, nome)
					}
				}
			})

			t.Run("todos num repositório vazio", func(t *testing.T) {
				repositorio := montar(t)

				todos, err := repositorio.Todos(ctx)

				if err != nil {
					t.Fatalf("erro inesperado: %v", err)
				}
				if len(todos) != 0 {
					t.Errorf("vieram %d produtos, esperado nenhum", len(todos))
				}
			})
		})
	}
}

// ----------------------------------------------------------------------
// O que só a implementação SQL faz.
// ----------------------------------------------------------------------

func padariaNoBanco(t *testing.T) *RepositorioSQL {
	t.Helper()

	ctx := context.Background()
	repositorio := repositorioSQLDeTeste(t)

	if err := repositorio.Salvar(ctx, Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := repositorio.Repor(ctx, "Broa", 5); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	return repositorio
}

func contarPedidos(t *testing.T, repositorio *RepositorioSQL) int {
	t.Helper()

	// O teste consulta o banco direto, sem passar pelo código sob teste. É
	// legítimo: aqui a pergunta é o que ficou GRAVADO, e não o que o método
	// diz ter feito.
	var total int
	if err := repositorio.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM pedidos`).Scan(&total); err != nil {
		t.Fatalf("contar pedidos: %v", err)
	}

	return total
}

func TestRegistrarVenda(t *testing.T) {
	ctx := context.Background()
	repositorio := padariaNoBanco(t)

	total, err := repositorio.RegistrarVenda(ctx, RelogioFixo{Momento: momentoDeTeste}, "Maria", "Broa", 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if total != 600 {
		t.Errorf("total = %d, esperado 600", total)
	}

	restante, err := repositorio.QuantidadeEm(ctx, "Broa")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if restante != 3 {
		t.Errorf("sobrou %d, esperado 3", restante)
	}

	if pedidos := contarPedidos(t, repositorio); pedidos != 1 {
		t.Errorf("gravou %d pedidos, esperado 1", pedidos)
	}
}

func TestRegistrarVendaGravaOsDadosCertos(t *testing.T) {
	ctx := context.Background()
	repositorio := padariaNoBanco(t)

	if _, err := repositorio.RegistrarVenda(ctx, RelogioFixo{Momento: momentoDeTeste}, "Maria", "Broa", 2); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var cliente, produto, em string
	var quantidade, total int
	err := repositorio.db.QueryRowContext(ctx,
		`SELECT cliente, produto, quantidade, total_em_centavos, em FROM pedidos`).
		Scan(&cliente, &produto, &quantidade, &total, &em)
	if err != nil {
		t.Fatalf("ler pedido: %v", err)
	}

	if cliente != "Maria" {
		t.Errorf("cliente = %q, esperado Maria", cliente)
	}
	if produto != "Broa" {
		t.Errorf("produto = %q, esperado Broa", produto)
	}
	if quantidade != 2 {
		t.Errorf("quantidade = %d, esperado 2", quantidade)
	}
	if total != 600 {
		t.Errorf("total = %d, esperado 600", total)
	}

	gravado, erroDeData := time.Parse(time.RFC3339, em)
	if erroDeData != nil {
		t.Fatalf("a hora gravada não está em RFC3339: %q", em)
	}
	if !gravado.Equal(momentoDeTeste) {
		t.Errorf("em = %v, esperado %v: o relógio injetado não foi usado", gravado, momentoDeTeste)
	}
}

func TestRegistrarVendaRecusaEDesfaz(t *testing.T) {
	casos := []struct {
		nome              string
		produto           string
		quantidade        int
		sentinelaEsperada error
	}{
		{"quantidade zero", "Broa", 0, ErrQuantidadeInvalida},
		{"quantidade negativa", "Broa", -2, ErrQuantidadeInvalida},
		{"produto que não existe", "Croissant", 1, ErrProdutoNaoEncontrado},
		{"mais do que tem em estoque", "Broa", 99, ErrEstoqueInsuficiente},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			ctx := context.Background()
			repositorio := padariaNoBanco(t)

			total, err := repositorio.RegistrarVenda(ctx, RelogioFixo{Momento: momentoDeTeste}, "Maria", caso.produto, caso.quantidade)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}
			if total != 0 {
				t.Errorf("com erro, o total deveria ser 0, veio %d", total)
			}

			// A transação precisa ter desfeito tudo: nem estoque baixado, nem
			// pedido gravado.
			restante, err := repositorio.QuantidadeEm(ctx, "Broa")
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if restante != 5 {
				t.Errorf("o estoque foi para %d numa venda recusada, esperado 5", restante)
			}
			if pedidos := contarPedidos(t, repositorio); pedidos != 0 {
				t.Errorf("gravou %d pedidos numa venda recusada", pedidos)
			}
		})
	}
}

func TestVendasSeguidasAcumulam(t *testing.T) {
	ctx := context.Background()
	repositorio := padariaNoBanco(t)
	relogio := RelogioFixo{Momento: momentoDeTeste}

	for range 3 {
		if _, err := repositorio.RegistrarVenda(ctx, relogio, "Maria", "Broa", 1); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
	}

	restante, err := repositorio.QuantidadeEm(ctx, "Broa")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if restante != 2 {
		t.Errorf("sobrou %d, esperado 2", restante)
	}
	if pedidos := contarPedidos(t, repositorio); pedidos != 3 {
		t.Errorf("gravou %d pedidos, esperado 3", pedidos)
	}
}

func TestDemonstracaoInjecaoDeSQLNaoFunciona(t *testing.T) {
	// Este teste já passa depois que Buscar existir. Ele mostra que um valor
	// entregue como parâmetro nunca é interpretado como comando, por mais
	// hostil que seja o texto.
	ctx := context.Background()
	repositorio := repositorioSQLDeTeste(t)

	if err := repositorio.Salvar(ctx, Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	malicioso := `Broa'; DROP TABLE produtos; --`
	if _, err := repositorio.Buscar(ctx, malicioso); !errors.Is(err, ErrProdutoNaoEncontrado) {
		t.Fatalf("esperava produto não encontrado, veio %v", err)
	}

	// A tabela continua lá, e a Broa também.
	if _, err := repositorio.Buscar(ctx, "Broa"); err != nil {
		t.Fatalf("a tabela sumiu: %v", err)
	}

	t.Log("o texto hostil foi tratado como nome de produto, não como comando")
}
