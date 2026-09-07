# Módulo 01 — Tipos, structs e o valor zero

A padaria vai aprender a descrever o que vende. Para isso você precisa de dois
tipos, `Produto` e `Pedido`, e de quatro funções que trabalham com eles.

O assunto de verdade deste módulo é outro, e é o que mais derruba quem vem de
JavaScript, Python ou PHP: **em Go, passar uma struct para uma função entrega
uma cópia**. E **não existe null**.

```bash
go test ./modulos/01-tipos/...
```

## Por que Go faz assim

**Cadê a classe?**
Não tem. Tem `struct`, que é só um agrupamento de campos, e funções que
recebem essa struct. Comportamento e dado não vivem grudados por padrão. No
módulo 04 você vai aprender a pendurar métodos numa struct, mas mesmo lá o
método continua sendo uma função com um parâmetro a mais.

**Cadê o construtor?**
Também não tem. O que existe é a convenção de escrever `NovoProduto` e
devolver a struct já montada. Não é regra da linguagem, é hábito da
comunidade. Consequência prática: ninguém consegue impedir alguém de montar um
`Produto{}` na mão, então o valor zero do seu tipo precisa fazer sentido.

**Cadê o `null`, o `None`, o `undefined`?**
Não existe para struct. Todo tipo em Go tem um **valor zero**, que é o que
você recebe quando declara sem inicializar. String vira `""`, número vira `0`,
bool vira `false`, e uma struct vira uma struct com todos os campos no zero
deles. `var p Produto` é um Produto pronto para uso, não uma bomba.

Isso muda como você projeta. Em vez de checar `if (produto == null)` em toda
função, você escolhe campos cujo valor zero já seja o comportamento correto.
Olhe a solução de `TotalEmCentavos`: ela não tem nenhum `if`, porque preço
zero vezes quantidade zero já dá o total certo de um pedido vazio.

**Por que cópia em vez de referência?**
Porque você consegue ler uma função e saber que ela não estragou o seu dado.
Quando um valor é copiado, quem chamou fica protegido por construção, não por
disciplina. É por isso que Go não precisa de `const`, de `readonly`, nem de
biblioteca de imutabilidade para o caso comum. Existe passagem por referência
em Go, com ponteiro, e ela é explícita justamente para você ver no código
quando alguém pode mexer no seu dado. Isso é o módulo 04.

**Por que divisão de inteiro corta a parte quebrada?**
`333 * 10 / 100` dá `33`, não `33,3` nem `34`. Go não arredonda por você, e
não converte para ponto flutuante escondido. Num sistema que lida com
dinheiro isso é vantagem: o arredondamento vira uma decisão sua, escrita no
código, em vez de um comportamento que você descobre no fechamento do caixa.
Neste módulo a conta corta para baixo, o que favorece a padaria em um centavo.
Se favorecer o cliente for a política, isso precisa estar escrito.

## Pegadinhas

**A struct é copiada, e isso inclui a struct dentro da struct.**
`Pedido` tem um `Produto` inteiro dentro dele, não um endereço para um
Produto. Copiar o pedido copia o produto junto. Vindo de linguagem onde
objeto é sempre referência, essa é a diferença que mais causa bug de lógica
invertida: lá você muda sem querer, aqui você não muda achando que mudou.

**Só o que começa com maiúscula sai do pacote, e isso vale para campo.**
`Nome` é visível de fora do pacote. `nome` não seria. Você vai reencontrar
essa regra no módulo 07, quando um campo minúsculo simplesmente não aparecer
no JSON e ninguém avisar nada.

**Struct compara com `==`.**
Duas structs com os mesmos campos são iguais, sem você escrever nada. Vale
enquanto todos os campos forem comparáveis, o que deixa de valer quando
aparecerem slices e maps no módulo 02. O teste `TestProdutosSaoComparaveis`
está lá para você ver funcionando antes de perder isso.

## O exercício

Em `padaria/padaria.go`, quatro funções:

1. `NovoProduto` monta um produto já disponível para venda.
2. `Descrever` escreve `Pão de queijo: R$ 4,50`, ou `Pão de queijo: esgotado`,
   e troca nome vazio por `produto sem nome`.
3. `AplicarDesconto` devolve um produto novo, sem tocar no que recebeu.
4. `TotalEmCentavos` multiplica preço por quantidade.

`FormatarPreco` já vem pronta, do módulo 00. Cada módulo é autossuficiente,
então o que você resolveu antes chega implementado aqui.

O teste `TestAplicarDescontoNaoMexeNoOriginal` é o mais importante do módulo.
Se ele passar sem você ter feito esforço nenhum, ótimo: é exatamente esse o
ponto.

## Desafio, sem teste pronto

Crie um tipo próprio para categoria, em vez de usar string solta:

```go
type Categoria string

const (
	CategoriaPao     Categoria = "pão"
	CategoriaDoce    Categoria = "doce"
	CategoriaBebida  Categoria = "bebida"
)
```

Adicione o campo em `Produto` e faça `Descrever` mostrar a categoria. Depois
responda para você mesmo: qual é o valor zero de `Categoria`, e o que a
padaria deveria fazer com um produto que caiu nele? Não existe enum de
verdade em Go, e essa pergunta é o preço disso.

## Se travar

A solução está em `solucoes/01-tipos/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
