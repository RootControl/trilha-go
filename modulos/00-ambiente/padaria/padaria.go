// Package padaria guarda o código da Padaria do Seu Zé, o projeto que
// acompanha a Trilha Go do primeiro módulo até o último.
//
// Neste módulo a padaria ainda não vende nada. Ela só sabe receber quem
// entra e dizer quanto custa uma coisa.
package padaria

// Saudacao devolve a frase que a padaria usa para receber um cliente.
//
// Para o nome "Maria", o resultado esperado é:
//
//	Bom dia, Maria! Bem-vindo à Padaria do Seu Zé.
//
// Quando o nome chega vazio, use a palavra "cliente" no lugar dele.
func Saudacao(nome string) string {
	// TODO: implemente esta função.
	return ""
}

// FormatarPreco recebe um valor em centavos e devolve o preço escrito em
// reais, no formato brasileiro:
//
//	450    ->  R$ 4,50
//	1000   ->  R$ 10,00
//	5      ->  R$ 0,05
//
// A padaria guarda dinheiro em centavos, nunca em float64. O motivo está
// explicado no README deste módulo.
func FormatarPreco(centavos int) string {
	// TODO: implemente esta função.
	return ""
}
