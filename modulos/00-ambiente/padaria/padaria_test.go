package padaria

import "testing"

// Este é o formato de teste que você vai ver o curso inteiro: uma tabela de
// casos e um laço que roda cada caso como um subteste. Leia a tabela como a
// especificação da função. Ela é o enunciado do exercício.

func TestSaudacao(t *testing.T) {
	casos := []struct {
		nome     string
		entrada  string
		esperado string
	}{
		{
			nome:     "cliente com nome",
			entrada:  "Maria",
			esperado: "Bom dia, Maria! Bem-vindo à Padaria do Seu Zé.",
		},
		{
			nome:     "cliente sem nome vira cliente",
			entrada:  "",
			esperado: "Bom dia, cliente! Bem-vindo à Padaria do Seu Zé.",
		},
		{
			nome:     "nome com acento é preservado",
			entrada:  "Antônio",
			esperado: "Bom dia, Antônio! Bem-vindo à Padaria do Seu Zé.",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := Saudacao(caso.entrada)
			if recebido != caso.esperado {
				t.Errorf("Saudacao(%q) = %q, esperado %q", caso.entrada, recebido, caso.esperado)
			}
		})
	}
}

func TestFormatarPreco(t *testing.T) {
	casos := []struct {
		nome     string
		centavos int
		esperado string
	}{
		{nome: "preço com centavos", centavos: 450, esperado: "R$ 4,50"},
		{nome: "preço redondo", centavos: 1000, esperado: "R$ 10,00"},
		{nome: "centavos precisam de zero à esquerda", centavos: 5, esperado: "R$ 0,05"},
		{nome: "de graça", centavos: 0, esperado: "R$ 0,00"},
		{nome: "encomenda cara", centavos: 123456, esperado: "R$ 1234,56"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			recebido := FormatarPreco(caso.centavos)
			if recebido != caso.esperado {
				t.Errorf("FormatarPreco(%d) = %q, esperado %q", caso.centavos, recebido, caso.esperado)
			}
		})
	}
}
