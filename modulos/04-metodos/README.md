# Módulo 04 — Métodos e ponteiros

A padaria não ganha nenhuma capacidade nova neste módulo. Ela ganha uma forma
nova: o que era função solta vira método pendurado no tipo, e aparece um
`Caixa` para acumular as vendas do dia.

O assunto de verdade é a escolha do receptor. Errar essa escolha produz o bug
mais silencioso de Go: um método que roda, não reclama de nada, e não muda
nada.

```bash
go test ./modulos/04-metodos/...
```

Comece pelo exemplo já resolvido no topo de `padaria/padaria.go`. A função
`Baixar` do módulo 03 está lá, convertida em método, com o antes e o depois
lado a lado.

## Por que Go faz assim

**Cadê a classe? Cadê o `this`?**
Um método em Go é uma função com um parâmetro a mais, escrito num lugar
especial:

```go
func Baixar(estoque Estoque, nome string, quantidade int) error  // função
func (e Estoque) Baixar(nome string, quantidade int) error       // método
```

O primeiro parâmetro saiu da lista e virou o **receptor**. É só isso. Não há
classe, não há `this` implícito, e o nome do receptor é escolha sua. A
convenção é uma ou duas letras, sempre a mesma para todos os métodos daquele
tipo, e nunca `this` nem `self`. Como o receptor é um parâmetro comum, você
enxerga na assinatura se ele é cópia ou endereço, o que em outras linguagens é
invisível.

**Por que existem dois tipos de receptor?**
Porque o módulo 01 continua valendo dentro do método. Receptor de valor recebe
uma cópia, e escrever nessa cópia não alcança ninguém. Receptor de ponteiro
recebe o endereço, e escrever alcança o original.

A regra prática cabe em duas linhas:

- O método **escreve** em algum campo? Receptor de ponteiro.
- O método só **lê**? Receptor de valor.

Quem chama nem percebe a diferença na hora da chamada. `produto.Reajustar(10)`
tem exatamente a mesma cara nos dois casos, porque o Go monta o `&produto`
sozinho quando precisa. A diferença aparece só no resultado, e é por isso que
o erro é silencioso.

**Por que `Repor` muda o estoque com receptor de valor, então?**
Porque o que foi copiado ali não é a tabela, é a referência para ela. Um
`map` é um cabeçalho apontando para uma tabela de hash, do mesmo jeito que uma
fatia aponta para um array. Copiar o cabeçalho não copia a tabela, e os dois
lados escrevem no mesmo lugar. É a mesma lição do módulo 02, agora aplicada ao
receptor.

Isso vale para map, fatia e canal. Não vale para struct.

**Método não é só para struct.**
`Estoque` é um `map[string]int` com nome próprio, e tem métodos. Você pode
pendurar método em qualquer tipo que **você** declarou, inclusive
`type Centavos int`. O que não dá é pendurar método num tipo de outro pacote,
e essa restrição é o que impede o "monkey patching" que existe em Ruby,
Python e JavaScript.

**Cadê os getters e setters?**
Não tem cultura disso em Go. Campo que precisa ser lido de fora fica
exportado, com letra maiúscula, e pronto. Escrever `GetNome()` para devolver
`p.nome` é considerado ruído. Encapsulamento em Go acontece na fronteira do
pacote, não na fronteira do tipo.

**Ponteiro em Go assusta menos do que parece.**
Não existe aritmética de ponteiro, não existe `free`, e o coletor de lixo
cuida da memória. Um ponteiro em Go serve para duas coisas: deixar uma função
alterar o seu dado, e evitar copiar uma struct grande. Você quase nunca
escreve `*p` explicitamente, porque `p.Campo` já desreferencia sozinho.

## Pegadinhas

**Receptor de valor num método que deveria escrever.**
Compila. Roda. Não avisa nada. Não muda nada. É o bug que
`TestDemonstracaoReceptorDeValorNaoAlteraNada` reproduz, com um método errado
de propósito escrito dentro do arquivo de teste. Rode e leia:

```bash
go test -v -run Demonstracao ./modulos/04-metodos/...
```

Se algum teste deste módulo falhar dizendo que o valor não mudou, olhe o
receptor antes de olhar o corpo.

**Elemento de map não tem endereço.**
Esta é a diferença que pega todo mundo. Elemento de fatia é endereçável, então
o Go monta o `&` sozinho:

```go
cardapio := []Produto{{PrecoEmCentavos: 100}}
cardapio[0].Reajustar(10) // funciona
```

Elemento de map não é, e o compilador recusa:

```go
cardapio := map[string]Produto{"pão": {PrecoEmCentavos: 100}}
cardapio["pão"].Reajustar(10) // cannot call pointer method Reajustar on Produto
```

O motivo é que a tabela de hash pode mover os valores de lugar quando cresce,
então um endereço para dentro dela não seria confiável. A saída é tirar,
mexer e devolver:

```go
produto := cardapio["pão"]
produto.Reajustar(10)
cardapio["pão"] = produto
```

Ou guardar `map[string]*Produto`, o que resolve isso e cria outros problemas.

**`String` que chama `%v` em si mesma nunca termina.**

```go
func (c Caixa) String() string {
	return fmt.Sprintf("caixa com %v", c) // chama String de novo, para sempre
}
```

O programa consome a pilha inteira e morre. Dentro de `String`, formate os
campos, nunca o receptor.

**`String` com receptor de ponteiro some quando você imprime um valor.**
Se `String` tivesse receptor `*Caixa`, `fmt.Printf("%v", caixa)` sobre um
`Caixa` comum voltaria a imprimir os campos crus, sem erro nenhum e sem
aviso. Um valor não carrega os métodos de ponteiro do seu tipo; um ponteiro
carrega os dois. Por isso `String` fica no receptor de valor, mesmo o `Caixa`
tendo `Registrar` no ponteiro.

**Misturar receptores tem custo, e este módulo mistura de propósito.**
A recomendação comum é usar o mesmo tipo de receptor em todos os métodos de um
tipo. `Caixa` desobedece, e a justificativa está no parágrafo acima. A
recomendação existe principalmente por causa de interface, e é lá, no módulo
05, que a mistura passa a doer de verdade.

## O exercício

Oito métodos em `padaria/padaria.go`. Três deles têm o receptor escrito
errado de propósito, e faz parte do exercício descobrir quais:

| Tipo | Método | O que faz |
|------|--------|-----------|
| `Produto` | `Descrever` | monta a linha da placa |
| `Produto` | `Reajustar` | aumenta o preço |
| `Produto` | `Esgotar` | tira de venda |
| `Estoque` | `Quantidade` | lê o estoque |
| `Estoque` | `Repor` | soma unidades |
| `Caixa` | `Registrar` | anota uma venda |
| `Caixa` | `TicketMedio` | média por venda |
| `Caixa` | `String` | a linha de resumo |

`String` merece atenção. Você não vai registrar esse método em lugar nenhum,
não vai declarar que implementa coisa alguma, e mesmo assim `fmt.Printf` vai
achar e usar. Essa é a primeira aparição de interface implícita no curso, sem
o nome. O nome vem no próximo módulo.

## Desafio, sem teste pronto

Dê ao `Caixa` a capacidade de fechar o dia: um método `Fechar()` que devolve o
total e zera o caixa para o dia seguinte.

Depois responda duas perguntas. Qual receptor `Fechar` precisa ter, e por quê?
E o que acontece se alguém chamar `Fechar` a partir de um `*Caixa` que é nil?
Teste, veja a mensagem, e guarde: chamar método em ponteiro nil é permitido, e
o que derruba o programa é o acesso ao campo lá dentro.

## Se travar

A solução está em `solucoes/04-metodos/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
