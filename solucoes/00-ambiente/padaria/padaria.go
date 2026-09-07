// Package padaria guarda o código da Padaria do Seu Zé, o projeto que
// acompanha a Trilha Go do primeiro módulo até o último.
//
// Esta é a solução do módulo 00. Olhe depois de tentar, não antes.
package padaria

import "fmt"

// Saudacao devolve a frase que a padaria usa para receber um cliente.
func Saudacao(nome string) string {
	if nome == "" {
		nome = "cliente"
	}
	return fmt.Sprintf("Bom dia, %s! Bem-vindo à Padaria do Seu Zé.", nome)
}

// FormatarPreco recebe um valor em centavos e devolve o preço escrito em
// reais, no formato brasileiro.
//
// O verbo %02d preenche com zero à esquerda até ter dois dígitos. É ele que
// transforma 5 centavos em "0,05" em vez de "0,5".
func FormatarPreco(centavos int) string {
	return fmt.Sprintf("R$ %d,%02d", centavos/100, centavos%100)
}
