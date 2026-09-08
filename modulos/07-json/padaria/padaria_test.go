package padaria

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func pedidosDeTeste() []Pedido {
	return []Pedido{
		{
			Cliente:    "Maria",
			Produto:    Produto{Nome: "Pão de queijo", PrecoEmCentavos: 450, Disponivel: true},
			Quantidade: 3,
			Em:         time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC),
		},
		{
			Cliente:    "João",
			Produto:    Produto{Nome: "Pão francês", PrecoEmCentavos: 100, Disponivel: true},
			Quantidade: 10,
			Em:         time.Date(2026, time.March, 15, 8, 15, 0, 0, time.UTC),
			Observacao: "sem sal",
		},
	}
}

// exigirPedidosIguais compara pedido a pedido.
//
// Não dá para comparar Pedido com == por causa do time.Time lá dentro. Um
// time.Time carrega o fuso e, às vezes, um relógio monotônico, e dois
// instantes que representam o mesmo momento podem ser diferentes byte a byte.
// Para tempo, sempre Equal.
func exigirPedidosIguais(t *testing.T, recebidos, esperados []Pedido) {
	t.Helper()

	if len(recebidos) != len(esperados) {
		t.Fatalf("vieram %d pedidos, esperado %d", len(recebidos), len(esperados))
	}

	for i := range esperados {
		recebido, esperado := recebidos[i], esperados[i]

		if recebido.Cliente != esperado.Cliente {
			t.Errorf("pedido %d: Cliente = %q, esperado %q", i, recebido.Cliente, esperado.Cliente)
		}
		if recebido.Produto != esperado.Produto {
			t.Errorf("pedido %d: Produto = %+v, esperado %+v", i, recebido.Produto, esperado.Produto)
		}
		if recebido.Quantidade != esperado.Quantidade {
			t.Errorf("pedido %d: Quantidade = %d, esperado %d", i, recebido.Quantidade, esperado.Quantidade)
		}
		if !recebido.Em.Equal(esperado.Em) {
			t.Errorf("pedido %d: Em = %v, esperado %v", i, recebido.Em, esperado.Em)
		}
		if recebido.Observacao != esperado.Observacao {
			t.Errorf("pedido %d: Observacao = %q, esperado %q", i, recebido.Observacao, esperado.Observacao)
		}
	}
}

func TestSalvarPedidosProduzOArquivoDourado(t *testing.T) {
	// Arquivo dourado, ou golden file: a saída esperada mora em testdata/ em
	// vez de dentro do código do teste. Serve quando a saída é grande demais
	// para caber legível numa string, e tem a vantagem de você conseguir
	// abrir e ler o arquivo como um humano leria.
	dourado, err := os.ReadFile(filepath.Join("testdata", "pedidos.json"))
	if err != nil {
		t.Fatalf("não consegui ler o arquivo dourado: %v", err)
	}

	var buffer bytes.Buffer
	if err := SalvarPedidos(&buffer, pedidosDeTeste()); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if buffer.String() != string(dourado) {
		t.Errorf("a saída não bate com testdata/pedidos.json\n\nrecebido:\n%s\nesperado:\n%s",
			buffer.String(), dourado)
	}
}

func TestSalvarPedidosOmiteOsCamposCertos(t *testing.T) {
	pedido := Pedido{
		Cliente:    "Maria",
		Produto:    Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
		Quantidade: 1,
		Em:         time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC),
		Cancelado:  true, // não pode aparecer no arquivo
	}

	var buffer bytes.Buffer
	if err := SalvarPedidos(&buffer, []Pedido{pedido}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	saida := buffer.String()

	if strings.Contains(saida, "observacao") {
		t.Error(`"observacao" está vazia e não deveria aparecer: falta omitempty`)
	}
	if strings.Contains(strings.ToLower(saida), "cancelado") {
		t.Error(`"cancelado" nunca deveria aparecer: falta a etiqueta "-"`)
	}
	if !strings.Contains(saida, `"preco_em_centavos"`) {
		t.Errorf("esperava o campo preco_em_centavos na saída, veio:\n%s", saida)
	}
	if strings.Contains(saida, `"Nome"`) {
		t.Error("o campo saiu com o nome do Go: faltam as etiquetas de JSON")
	}
}

func TestCarregarPedidosDoArquivoDourado(t *testing.T) {
	arquivo, err := os.Open(filepath.Join("testdata", "pedidos.json"))
	if err != nil {
		t.Fatalf("não consegui abrir o arquivo dourado: %v", err)
	}
	defer arquivo.Close()

	// os.Open devolve um *os.File, e um *os.File é um io.Reader. Nenhuma
	// conversão, nenhum adaptador: ele simplesmente tem o método Read.
	pedidos, err := CarregarPedidos(arquivo)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	exigirPedidosIguais(t, pedidos, pedidosDeTeste())
}

func TestIdaEVolta(t *testing.T) {
	originais := pedidosDeTeste()

	var buffer bytes.Buffer
	if err := SalvarPedidos(&buffer, originais); err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}

	// O mesmo buffer que serviu de io.Writer na ida serve de io.Reader na
	// volta. É por isso que as duas funções recebem interface em vez de nome
	// de arquivo: o teste roda inteiro na memória, sem tocar em disco.
	recarregados, err := CarregarPedidos(&buffer)
	if err != nil {
		t.Fatalf("erro ao carregar: %v", err)
	}

	exigirPedidosIguais(t, recarregados, originais)
}

func TestSalvarECarregarEmArquivo(t *testing.T) {
	// t.TempDir devolve uma pasta nova, exclusiva deste teste, e apaga tudo
	// sozinho no fim, mesmo se o teste falhar. Nunca escreva num caminho fixo
	// dentro de teste.
	caminho := filepath.Join(t.TempDir(), "pedidos.json")

	if err := SalvarPedidosEmArquivo(caminho, pedidosDeTeste()); err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}

	recarregados, err := CarregarPedidosDeArquivo(caminho)
	if err != nil {
		t.Fatalf("erro ao carregar: %v", err)
	}

	exigirPedidosIguais(t, recarregados, pedidosDeTeste())
}

func TestSalvarSobrescreveOArquivo(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "pedidos.json")

	if err := SalvarPedidosEmArquivo(caminho, pedidosDeTeste()); err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}
	if err := SalvarPedidosEmArquivo(caminho, pedidosDeTeste()[:1]); err != nil {
		t.Fatalf("erro ao salvar de novo: %v", err)
	}

	recarregados, err := CarregarPedidosDeArquivo(caminho)
	if err != nil {
		t.Fatalf("erro ao carregar: %v", err)
	}
	if len(recarregados) != 1 {
		t.Errorf("vieram %d pedidos, esperado 1: o arquivo não foi truncado", len(recarregados))
	}
}

func TestCarregarArquivoInexistenteNaoEhErro(t *testing.T) {
	// Decisão de produto: padaria que abriu hoje ainda não tem arquivo.
	caminho := filepath.Join(t.TempDir(), "ainda-nao-existe.json")

	pedidos, err := CarregarPedidosDeArquivo(caminho)

	if err != nil {
		t.Fatalf("arquivo inexistente não deveria ser erro, veio: %v", err)
	}
	if len(pedidos) != 0 {
		t.Errorf("vieram %d pedidos, esperado nenhum", len(pedidos))
	}
}

func TestCarregarArquivoQuebradoEhErro(t *testing.T) {
	_, err := CarregarPedidosDeArquivo(filepath.Join("testdata", "quebrado.json"))

	if err == nil {
		t.Fatal("esperava erro para JSON inválido")
	}
	if !strings.Contains(err.Error(), "quebrado.json") {
		t.Errorf("o erro deveria citar o caminho, veio %q", err.Error())
	}
	// Um JSON inválido não pode ser confundido com arquivo ausente.
	if errors.Is(err, fs.ErrNotExist) {
		t.Error("JSON inválido virou erro de arquivo ausente")
	}
}

func TestDemonstracaoCampoAusenteEZeroSaoIndistinguiveis(t *testing.T) {
	// Este teste já passa. Ele mostra a limitação mais importante do
	// encoding/json com structs.
	semQuantidade := `[{"cliente": "Maria"}]`
	comZero := `[{"cliente": "Maria", "quantidade": 0}]`

	a, err := CarregarPedidos(strings.NewReader(semQuantidade))
	if err != nil || len(a) == 0 {
		t.Skip("implemente CarregarPedidos para ver esta demonstração")
	}
	b, err := CarregarPedidos(strings.NewReader(comZero))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if a[0].Quantidade != b[0].Quantidade {
		t.Fatal("era para os dois darem no mesmo")
	}

	t.Log("campo ausente e campo com zero produzem o mesmo Pedido: o zero value não distingue os dois")
}

func TestDemonstracaoCampoDesconhecidoEhIgnorado(t *testing.T) {
	// Este teste já passa. Por padrão, o decodificador ignora em silêncio
	// qualquer campo que ele não conhece.
	comLixo := `[{"cliente": "Maria", "campo_que_nao_existe": 42, "outro": {"a": 1}}]`

	pedidos, err := CarregarPedidos(strings.NewReader(comLixo))
	if err != nil || len(pedidos) == 0 {
		t.Skip("implemente CarregarPedidos para ver esta demonstração")
	}

	if pedidos[0].Cliente != "Maria" {
		t.Fatalf("Cliente = %q, esperado Maria", pedidos[0].Cliente)
	}

	t.Log("o campo desconhecido sumiu sem reclamação: use DisallowUnknownFields quando isso importar")
}

func TestDemonstracaoOrdemDoDefer(t *testing.T) {
	// Este teste já passa. Defers empilham: o último registrado é o primeiro
	// a rodar. É o que faz o padrão abrir-e-adiar-o-fechamento funcionar
	// quando você abre várias coisas em sequência.
	var ordem []string

	func() {
		defer func() { ordem = append(ordem, "primeiro defer") }()
		defer func() { ordem = append(ordem, "segundo defer") }()
		ordem = append(ordem, "corpo da função")
	}()

	esperado := []string{"corpo da função", "segundo defer", "primeiro defer"}
	for i := range esperado {
		if ordem[i] != esperado[i] {
			t.Fatalf("ordem = %v, esperado %v", ordem, esperado)
		}
	}

	t.Logf("ordem de execução: %v", ordem)
}
