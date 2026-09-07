# Módulo 02 — Fatias e maps

A padaria ganha cardápio e estoque. O cardápio é uma fatia de produtos, o
estoque é um map de nome para quantidade.

Este módulo tem um par montado com o anterior. No módulo 01 você provou que
passar uma struct para uma função **protege** o seu dado. Aqui você vai provar
que passar uma fatia ou um map **não protege nada**. Entender por que as duas
coisas são verdade ao mesmo tempo é o objetivo real.

```bash
go test ./modulos/02-colecoes/...
```

## Por que Go faz assim

**Cadê a lista, o array associativo, o dicionário?**
São dois tipos e só dois: `[]T` para sequência e `map[K]V` para associação.
Não há `List`, `ArrayList`, `Vector`, `Set` nem `OrderedDict` para escolher.
Set, quando você precisar, é um `map[T]struct{}` e ninguém acha isso estranho.

**Por que a fatia não protege quem chamou?**
Porque uma fatia não é o array. Ela é um valor pequeno com três campos: um
ponteiro para o começo dos dados, o comprimento e a capacidade. Quando você
passa a fatia para uma função, esses três campos são copiados, fielmente ao
que o módulo 01 ensinou. Só que o ponteiro copiado aponta para o mesmo array
de sempre. A função recebeu um mapa do tesouro novo, apontando para o mesmo
tesouro.

Consequência prática, e ela é assimétrica:

- Alterar um **elemento** dentro da função alcança quem chamou.
- Dar `append` dentro da função **pode ou não** alcançar quem chamou, e é por
  isso que funções que crescem uma fatia sempre devolvem a fatia nova.

Map funciona parecido, com uma diferença: um map já é uma referência a uma
tabela, então `Repor` consegue mudar o estoque de quem chamou sem devolver
nada. Repare que a função não tem retorno nenhum e mesmo assim funciona.

**Por que a ordem de um map é aleatória?**
Não é acaso, é sabotagem deliberada. O runtime de Go começa o percurso de um
map num ponto aleatório, e muda esse ponto a cada execução, justamente para
você **não conseguir** depender da ordem. A alternativa seria você escrever
código que funciona por acidente hoje e quebra quando a tabela crescer. Aqui
o acidente é impossível: ou você ordena, ou você percebe no primeiro teste.
`TestNomesEmOrdemEhEstavel` roda cem vezes por esse motivo.

**Cadê `filter`, `map`, `reduce`?**
Escreva o laço. Sério. Um `for range` de quatro linhas é mais fácil de ler do
que uma cadeia de funções, e Go só ganhou genéricos em 2022, então a
biblioteca padrão nunca teve esse vocabulário. O que existe hoje são os
pacotes `slices` e `maps`, com ferramentas de propósito específico como
`slices.Sort` e `slices.Contains`. No módulo 09 você escreve as suas com
genéricos, e vai perceber que raramente precisa.

## Pegadinhas

**`for _, x := range` te dá uma cópia, e mexer nela não faz nada.**
Esta compila, roda, não avisa nada e não muda preço nenhum:

```go
for _, produto := range cardapio {
	produto.PrecoEmCentavos += 100 // some no fim da volta do laço
}
```

Para alcançar o elemento de verdade, percorra índices: `for i := range cardapio`
e escreva em `cardapio[i]`. É a mesma regra de cópia do módulo 01, aplicada à
variável do laço. Esse é o bug silencioso mais comum de quem está começando.

**Ler de um map nil é seguro, escrever derruba o programa.**

```go
var estoque Estoque      // nil
_ = estoque["Sonho"]     // devolve 0, tudo bem
estoque["Sonho"] = 3     // panic: assignment to entry in nil map
```

É por isso que `NovoEstoque` existe, e é por isso que o teste dela checa
explicitamente que o retorno não é nil. Toda struct sua que tiver um campo de
map precisa de uma função `NovoAlgumaCoisa` que inicialize esse map, ou o
valor zero do seu tipo vira uma armadilha.

**`append` pode escrever por cima de dados que você achava seus.**
Quando ainda cabe na capacidade, `append` não cria array novo: ele escreve na
posição seguinte do array que já existe. Se outra fatia estiver olhando para
aquela posição, ela muda também. O teste `TestDemonstracaoFatiaCompartilhaMemoria`
já passa desde o começo e existe só para você ver isso acontecer. Rode com
`go test -v -run Demonstracao ./modulos/02-colecoes/...` e leia a saída.

Quando precisar de uma cópia independente de verdade, peça: `slices.Clone`.

**Fatia e map dentro de uma struct tiram o `==` dela.**
No módulo 01 você comparou dois produtos com `==`. Isso para de funcionar no
instante em que a struct ganhar um campo de fatia ou de map:

```go
type Combo struct {
	Nome  string
	Itens []Produto
}

a == b // não compila: invalid operation, Combo cannot be compared
```

Fatias não são comparáveis porque não existe resposta óbvia: comparar os
endereços ou comparar o conteúdo? Go se recusa a escolher por você. Para
conteúdo, use `slices.Equal`, que é o que os testes deste módulo usam.

**Ordenar string ordena bytes, não português.**
`slices.Sort` compara byte a byte. Funciona no estoque deste módulo, mas
`"Água"` vai parar depois de `"Zebra"`, porque a letra acentuada ocupa bytes
mais altos em UTF-8. Ordenação com regra de idioma existe em
`golang.org/x/text/collate`, fora da biblioteca padrão. Saber que a diferença
existe já é o suficiente por agora.

## O exercício

Seis funções em `padaria/padaria.go`:

1. `Disponiveis` filtra o cardápio, preservando a ordem.
2. `MaisBarato` devolve o produto e um `bool` dizendo se achou. Esse par
   valor-e-encontrou é o degrau antes dos erros do módulo 03.
3. `Reajustar` aumenta os preços no lugar, sem devolver nada.
4. `NovoEstoque` devolve um map inicializado.
5. `Repor` soma quantidade.
6. `NomesEmOrdem` devolve as chaves ordenadas.

Comece por `Reajustar`. Se você escrever o laço da forma intuitiva, ele vai
falhar, e essa falha vale mais do que o resto do módulo junto.

## Depois que passar

`NomesEmOrdem` tem uma versão de uma linha só, com o que a biblioteca padrão
ganhou nas versões recentes:

```go
return slices.Sorted(maps.Keys(estoque))
```

`maps.Keys` devolve uma sequência preguiçosa e `slices.Sorted` coleta e
ordena. Escreva o laço primeiro, entenda o que ele faz, e só depois troque
pela linha curta. Na ordem inversa você decora em vez de aprender.

## Desafio, sem teste pronto

Escreva `Baixar(estoque Estoque, nome string, quantidade int)` para dar baixa
quando um produto é vendido. O estoque não pode ficar negativo, e a padaria
precisa saber que a baixa não aconteceu.

Você vai bater na mesma parede do desafio do módulo 00: dá para devolver um
`bool`, dá para devolver a quantidade que sobrou, dá para não devolver nada e
torcer. Escolha uma, escreva, e guarde. O módulo 03 é sobre exatamente essa
pergunta, e a resposta idiomática não é nenhuma das três.

## Se travar

A solução está em `solucoes/02-colecoes/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
