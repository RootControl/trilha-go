# Módulo 09 — Genéricos

A padaria para de repetir o mesmo laço para cada tipo.

Repare que os genéricos só aparecem no nono módulo. A ordem é essa de
propósito, e é a mesma ordem que a linguagem seguiu: Go passou treze anos sem
eles, e ganhou parâmetros de tipo só na versão 1.18, em 2022. Você precisa ter
sentido a repetição doer antes de removê-la.

```bash
go test ./modulos/09-genericos/...
```

Antes de implementar qualquer coisa, leia as três funções na seção **A
repetição** do arquivo. Elas já funcionam. Leia as três seguidas e repare que
o corpo é idêntico: muda o tipo e muda a condição, o laço é o mesmo copiado
três vezes.

## Por que Go faz assim

**Como se escreve.**

```go
func Filtrar[T any](itens []T, manter func(T) bool) []T
```

O que está entre colchetes, antes dos parênteses, é a lista de **parâmetros de
tipo**. `T` é o nome, `any` é a **restrição**. Daí para a frente, `T` funciona
como um tipo comum dentro da função.

**Restrição é conjunto de tipos, não conjunto de métodos.**
Esta é a novidade conceitual em relação ao módulo 05. Uma interface comum
descreve o que um tipo sabe **fazer**. Uma restrição descreve **quais tipos**
são aceitos, e o compilador usa isso para saber quais operações ele pode
permitir dentro da função.

| Restrição | Aceita | Permite |
|-----------|--------|---------|
| `any` | todo tipo | nada além de copiar e repassar |
| `comparable` | tipos com igualdade | `==`, `!=`, e ser chave de map |
| `cmp.Ordered` | números e string | `<`, `>`, e comparação |
| `Numero` (sua) | os tipos da união | os operadores aritméticos |

É por isso que `Filtrar` usa `any`, `AgruparPor` usa `comparable` na chave,
`MaiorPor` usa `cmp.Ordered` no valor e `Somar` usa uma união própria. Cada
uma pede o mínimo que precisa, exatamente como as interfaces do módulo 05.

**O til.**

```go
type Numero interface {
	~int | ~int64 | ~float64
}
```

A barra é união. O til quer dizer "qualquer tipo cujo tipo **subjacente** seja
este". Sem ele, a restrição aceitaria `int` e recusaria `Centavos`, porque
`Centavos` é um tipo próprio feito de `int`, e não um `int`.

Faça o experimento e desfaça depois: apague os três tis e rode o teste. O
compilador é generoso nesse erro específico:

```
Centavos does not satisfy Numero (possibly missing ~ for int in Numero)
```

Como restrição é conferida em tempo de compilação, esse tipo de coisa nunca
vira teste vermelho: vira build quebrado. É também o motivo de `Numero` já vir
pronta no exercício em vez de ser um `TODO`.

**Inferência de tipo.**
Você quase nunca escreve os colchetes na chamada. `Filtrar(numeros, ehPar)` e
`Filtrar[int](numeros, ehPar)` são a mesma coisa, e
`TestDemonstracaoInferenciaDeTipo` mostra as duas dando no mesmo. Escreva o
tipo explicitamente só quando o compilador não conseguir deduzir, o que
acontece principalmente quando o parâmetro de tipo aparece só no retorno.

**Tipo genérico.**
`Pilha[T any]` é um tipo, não uma função. O parâmetro entra na declaração e
reaparece no receptor dos métodos, escrito `*Pilha[T]`. Na hora de usar, o
tipo concreto vem junto: `var pilha Pilha[Produto]`.

E repare que não existe `NovaPilha`. O valor zero já funciona, porque `append`
em fatia nil funciona. A regra do módulo 01 continua valendo em código
genérico.

**Você já usava genéricos desde o módulo 02.**
`slices.Sort`, `slices.Equal`, `slices.Clone`, `maps.Keys`. Todos os pacotes
`slices` e `maps` são genéricos, e existem porque essa era a repetição que
mais doía na biblioteca padrão.

## Quando NÃO usar

Esta seção é a mais importante do módulo, porque o erro comum de quem vem de
Java ou TypeScript é generalizar cedo demais.

A recomendação do próprio time do Go é usar parâmetro de tipo em três
situações:

1. Funções que operam sobre fatia, map ou canal **sem depender do tipo do
   elemento**.
2. Tipos de contêiner de propósito geral, como uma pilha ou uma árvore.
3. Casos em que a implementação seria **letra por letra idêntica** para vários
   tipos.

E a regra negativa, que vale mais:

> Se o comportamento muda conforme o tipo, use interface. Se o código é
> idêntico e só o tipo muda, use parâmetro de tipo.

Um genérico com um `switch` de tipo por dentro é quase sempre uma interface
mal disfarçada. E, para um caso só, o laço escrito à mão continua sendo a
resposta certa: quatro linhas legíveis valem mais do que uma abstração que
existe para ser usada uma vez.

## Pegadinhas

**Faltou o til.**
Já visto acima. O erro é claro e diz o que fazer.

**Método não pode declarar parâmetro de tipo.**

```go
func (p *Pilha[T]) Mapear[R any](f func(T) R) []R  // não compila
```

```
syntax error: method must have no type parameters
```

Quem declara parâmetro de tipo é o **tipo**, não o método. É por isso que
`Mapear` neste módulo é uma função de pacote e não um método de fatia. A
limitação existe porque um método com tipo próprio quebraria a forma como Go
decide, em tempo de compilação, quais métodos um tipo tem.

**`var zero T` é o único jeito de escrever o valor zero de `T`.**
Você não pode escrever `T{}`, nem `0`, nem `nil`, porque não sabe o que `T`
vai ser. Declare uma variável e devolva ela. Aparece em `MaiorPor` e em
`Desempilhar`.

**`any` como restrição não deixa você fazer nada.**
Nem comparar com `==`. Se você precisa comparar, a restrição é `comparable`.
Se precisa ordenar, é `cmp.Ordered`. Se precisa somar, é uma união. A
restrição não é burocracia: é o que o compilador consulta para liberar cada
operador.

**Genérico não substitui interface, e vice-versa.**
As duas coisas resolvem problemas diferentes, e o módulo 05 continua valendo
inteiro. `BuscadorDeProdutos` não vira genérico: o comportamento de buscar
muda conforme a implementação, e é disso que interface trata.

## O exercício

Cinco funções e um tipo em `padaria/padaria.go`:

1. `Filtrar`, que substitui duas das funções repetidas.
2. `Mapear`, que substitui a terceira e ainda troca de tipo no caminho.
3. `Somar`, com restrição de união.
4. `MaiorPor`, com dois parâmetros de tipo e `cmp.Ordered`.
5. `AgruparPor`, com `comparable` na chave.
6. `Pilha[T]`, um tipo genérico com três métodos.

`TestFiltrarSubstituiAsFuncoesAntigas` é o argumento do módulo: ele roda a sua
função genérica e as funções escritas à mão sobre os mesmos dados, e exige o
mesmo resultado.

Depois que passar, olhe `TestSomarOTotalDeUmaLista`. Ele compõe `Mapear` com
`Somar` para calcular o faturamento do dia em uma linha, usando um problema
que a padaria tem de verdade.

## Desafio, sem teste pronto

**Escreva `Reduzir`**, a terceira função clássica dessa família:

```go
func Reduzir[T, A any](itens []T, inicial A, junta func(A, T) A) A
```

Depois reescreva `Somar` chamando `Reduzir`. Funciona, e o código fica mais
curto.

Agora a pergunta que importa, e ela não tem resposta única: ficou **melhor**?
Compare as duas versões de `Somar` lado a lado e decida qual você preferiria
encontrar num código que não é seu. A comunidade de Go, em geral, prefere a
primeira, e vale você entender por quê antes de concordar ou discordar.

**Se quiser ir além**, escreva a versão preguiçosa de `Filtrar`, devolvendo um
`iter.Seq[T]` em vez de uma fatia. Iteradores entraram na linguagem no Go 1.23
e permitem encadear transformações sem alocar uma fatia intermediária a cada
etapa.

## Se travar

A solução está em `solucoes/09-genericos/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
