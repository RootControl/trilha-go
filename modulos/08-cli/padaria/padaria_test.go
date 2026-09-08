package padaria

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var momentoDeTeste = time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)

// executar roda o programa inteiro e devolve o que saiu em cada canal.
//
// É este auxiliar que torna o módulo possível: o programa completo roda dentro
// do teste, sem processo novo, sem arquivo em caminho fixo, sem relógio real.
func executar(t *testing.T, args ...string) (saida string, erros string, err error) {
	t.Helper()

	var bufferSaida, bufferErros bytes.Buffer
	err = Executar(args, &bufferSaida, &bufferErros, RelogioFixo{Momento: momentoDeTeste})

	return bufferSaida.String(), bufferErros.String(), err
}

func arquivoTemporario(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "pedidos.json")
}

func TestExecutarCardapioEmTexto(t *testing.T) {
	saida, _, err := executar(t, "cardapio")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	esperado := "Broa: R$ 3,00\n" +
		"Pão de queijo: R$ 4,50\n" +
		"Pão francês: R$ 1,00\n" +
		"Sonho: esgotado\n"

	if saida != esperado {
		t.Errorf("saída:\n%s\nesperado:\n%s", saida, esperado)
	}
}

func TestExecutarCardapioEmJSON(t *testing.T) {
	saida, _, err := executar(t, "-json", "cardapio")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// As etiquetas do módulo 07 continuam valendo aqui.
	if !strings.Contains(saida, `"preco_em_centavos"`) {
		t.Errorf("esperava JSON com as etiquetas do módulo 07, veio:\n%s", saida)
	}
	if !strings.HasPrefix(strings.TrimSpace(saida), "[") {
		t.Errorf("esperava um array JSON, veio:\n%s", saida)
	}
}

func TestExecutarSemComando(t *testing.T) {
	saida, erros, err := executar(t)

	if !errors.Is(err, ErrUsoInvalido) {
		t.Errorf("errors.Is não achou ErrUsoInvalido em %v", err)
	}
	// Mensagem para gente vai no canal de erros, nunca na saída.
	if !strings.Contains(erros, "uso:") {
		t.Errorf("esperava o texto de uso no canal de erros, veio:\n%s", erros)
	}
	if saida != "" {
		t.Errorf("nada deveria ter ido para a saída, veio:\n%s", saida)
	}
}

func TestExecutarComandoDesconhecido(t *testing.T) {
	saida, _, err := executar(t, "assar")

	if err == nil {
		t.Fatal("esperava erro para comando desconhecido")
	}
	if !errors.Is(err, ErrUsoInvalido) {
		t.Errorf("errors.Is não achou ErrUsoInvalido em %v", err)
	}
	if !strings.Contains(err.Error(), "assar") {
		t.Errorf("o erro deveria citar o comando, veio %q", err.Error())
	}
	if saida != "" {
		t.Errorf("nada deveria ter ido para a saída, veio:\n%s", saida)
	}
}

func TestExecutarAjudaNaoEhErro(t *testing.T) {
	// Pedir ajuda com -h é uso correto do programa. O código de saída
	// precisa ser zero, senão todo script que chama o programa com -h
	// acredita que deu errado.
	_, erros, err := executar(t, "-h")

	if err != nil {
		t.Errorf("-h não deveria devolver erro, veio: %v", err)
	}
	if !strings.Contains(erros, "uso:") {
		t.Errorf("esperava o texto de uso, veio:\n%s", erros)
	}
	if !strings.Contains(erros, "-arquivo") {
		t.Errorf("esperava as opções listadas por PrintDefaults, veio:\n%s", erros)
	}
}

func TestExecutarVenderGravaEConfirma(t *testing.T) {
	caminho := arquivoTemporario(t)

	saida, _, err := executar(t, "-arquivo", caminho, "vender", "Pão de queijo", "3")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	esperado := `vendido: 3 de "Pão de queijo" por R$ 13,50` + "\n"
	if saida != esperado {
		t.Errorf("saída = %q, esperado %q", saida, esperado)
	}

	// E o pedido precisa ter sobrevivido no arquivo.
	pedidos, err := CarregarPedidosDeArquivo(caminho)
	if err != nil {
		t.Fatalf("erro ao recarregar: %v", err)
	}
	if len(pedidos) != 1 {
		t.Fatalf("gravou %d pedidos, esperado 1", len(pedidos))
	}
	if pedidos[0].Cliente != "balcão" {
		t.Errorf("Cliente = %q, esperado o padrão %q", pedidos[0].Cliente, "balcão")
	}
	if !pedidos[0].Em.Equal(momentoDeTeste) {
		t.Errorf("Em = %v, esperado %v: o relógio injetado não foi usado", pedidos[0].Em, momentoDeTeste)
	}
}

func TestExecutarVenderAcumulaNoArquivo(t *testing.T) {
	caminho := arquivoTemporario(t)

	if _, _, err := executar(t, "-arquivo", caminho, "vender", "Broa", "1"); err != nil {
		t.Fatalf("primeira venda: %v", err)
	}
	if _, _, err := executar(t, "-arquivo", caminho, "-cliente", "Maria", "vender", "Pão francês", "2"); err != nil {
		t.Fatalf("segunda venda: %v", err)
	}

	pedidos, err := CarregarPedidosDeArquivo(caminho)
	if err != nil {
		t.Fatalf("erro ao recarregar: %v", err)
	}
	if len(pedidos) != 2 {
		t.Fatalf("gravou %d pedidos, esperado 2: a segunda venda apagou a primeira", len(pedidos))
	}
	if pedidos[1].Cliente != "Maria" {
		t.Errorf("Cliente do segundo = %q, esperado Maria: a opção -cliente não foi usada", pedidos[1].Cliente)
	}
}

func TestExecutarVenderRecusa(t *testing.T) {
	casos := []struct {
		nome              string
		args              []string
		sentinelaEsperada error
	}{
		{
			nome:              "quantidade que não é número",
			args:              []string{"vender", "Pão de queijo", "três"},
			sentinelaEsperada: ErrQuantidadeInvalida,
		},
		{
			nome:              "produto que não está no cardápio",
			args:              []string{"vender", "Croissant", "1"},
			sentinelaEsperada: ErrProdutoNaoEncontrado,
		},
		{
			nome:              "mais do que tem em estoque",
			args:              []string{"vender", "Broa", "99"},
			sentinelaEsperada: ErrEstoqueInsuficiente,
		},
		{
			nome:              "faltou a quantidade",
			args:              []string{"vender", "Broa"},
			sentinelaEsperada: ErrUsoInvalido,
		},
		{
			nome:              "argumento sobrando",
			args:              []string{"vender", "Broa", "1", "agora"},
			sentinelaEsperada: ErrUsoInvalido,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			caminho := arquivoTemporario(t)
			args := append([]string{"-arquivo", caminho}, caso.args...)

			saida, _, err := executar(t, args...)

			if err == nil {
				t.Fatal("esperava erro")
			}
			if !errors.Is(err, caso.sentinelaEsperada) {
				t.Errorf("errors.Is não achou %v em %v", caso.sentinelaEsperada, err)
			}
			if saida != "" {
				t.Errorf("venda recusada não pode escrever na saída, veio:\n%s", saida)
			}

			// Recusou, não gravou.
			pedidos, err := CarregarPedidosDeArquivo(caminho)
			if err != nil {
				t.Fatalf("erro ao recarregar: %v", err)
			}
			if len(pedidos) != 0 {
				t.Errorf("gravou %d pedidos mesmo recusando a venda", len(pedidos))
			}
		})
	}
}

func TestExecutarPedidos(t *testing.T) {
	caminho := arquivoTemporario(t)

	if _, _, err := executar(t, "-arquivo", caminho, "vender", "Pão de queijo", "3"); err != nil {
		t.Fatalf("venda: %v", err)
	}

	saida, _, err := executar(t, "-arquivo", caminho, "pedidos")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	esperado := "2026-03-15 07:30 | balcão | 3 x Pão de queijo | R$ 13,50\n"
	if saida != esperado {
		t.Errorf("saída = %q, esperado %q", saida, esperado)
	}
}

func TestExecutarPedidosVazio(t *testing.T) {
	saida, _, err := executar(t, "-arquivo", arquivoTemporario(t), "pedidos")
	if err != nil {
		t.Fatalf("arquivo inexistente não deveria ser erro: %v", err)
	}

	if saida != "nenhum pedido registrado\n" {
		t.Errorf("saída = %q, esperado %q", saida, "nenhum pedido registrado\n")
	}
}

func TestExecutarPedidosVazioEmJSON(t *testing.T) {
	// Em JSON, lista vazia é [] e não uma frase em português. Quem consome
	// JSON é outro programa, e ele não sabe ler recado.
	saida, _, err := executar(t, "-arquivo", arquivoTemporario(t), "-json", "pedidos")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if strings.TrimSpace(saida) != "[]" {
		t.Errorf("saída = %q, esperado %q", strings.TrimSpace(saida), "[]")
	}
}

func TestEscreverPedidosFormatoDaLinha(t *testing.T) {
	pedidos := []Pedido{{
		Cliente:    "Maria",
		Produto:    Produto{Nome: "Broa", PrecoEmCentavos: 300, Disponivel: true},
		Quantidade: 2,
		Em:         time.Date(2026, time.December, 24, 18, 5, 0, 0, time.UTC),
	}}

	var buffer bytes.Buffer
	if err := EscreverPedidos(&buffer, pedidos, false); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	esperado := "2026-12-24 18:05 | Maria | 2 x Broa | R$ 6,00\n"
	if buffer.String() != esperado {
		t.Errorf("linha = %q, esperado %q", buffer.String(), esperado)
	}
}

func TestDemonstracaoDataDeReferencia(t *testing.T) {
	// Este teste já passa. Go não usa yyyy-MM-dd: você escreve a data de
	// referência, formatada do jeito que você quer. E a data de referência é
	// uma sequência: mês 1, dia 2, hora 3, minuto 4, segundo 5, ano 6,
	// fuso 7.
	momento := time.Date(2026, time.March, 15, 7, 30, 45, 0, time.UTC)

	casos := map[string]string{
		"2006-01-02":          "2026-03-15",
		"02/01/2006":          "15/03/2026",
		"15:04:05":            "07:30:45",
		"02 de January, 2006": "15 de March, 2026",
	}

	for layout, esperado := range casos {
		if recebido := momento.Format(layout); recebido != esperado {
			t.Errorf("Format(%q) = %q, esperado %q", layout, recebido, esperado)
		}
	}

	t.Log("o layout é um exemplo, não uma gramática: 1 2 3 4 5 6 7 = mês dia hora minuto segundo ano fuso")
}
