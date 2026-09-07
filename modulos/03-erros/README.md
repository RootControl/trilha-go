# Módulo 03 — Erros são valores

A padaria vai aprender a dizer que o pão acabou. Você já tentou isso duas
vezes: no desafio do módulo 00, com o troco, e no desafio do módulo 02, com a
baixa no estoque. Nas duas você teve que improvisar. Aqui está a resposta.

```bash
go test ./modulos/03-erros/...
```

## Por que Go faz assim

**Cadê o `try`, o `catch`, o `throw`?**
Não existem. Um erro em Go é um valor comum, devolvido junto com o resultado,
e a assinatura da função avisa isso na cara:

```go
func BuscarProduto(cardapio []Produto, nome string) (Produto, error)
```

Você olha para a assinatura e já sabe que essa chamada pode falhar. Com
exceção, qualquer linha pode desviar o fluxo para um lugar que você não
escreveu, e a assinatura não te conta nada. Em Go o caminho do erro é escrito
no código, na mesma indentação do caminho de sucesso, e passa pela revisão de
código junto com o resto.

**Então tem `if err != nil` em toda linha.**
Tem. Essa é a crítica mais repetida sobre a linguagem, e ela é justa em uma
coisa: é verboso mesmo. O que você compra com essa verbosidade é que **todo
ponto de falha fica visível no lugar onde ele acontece**, e que tratar erro
vira trabalho normal em vez de exceção. Ninguém escreve `catch (Exception e) {}`
em Go por acidente, porque ignorar erro em Go é uma linha explícita que
qualquer revisor enxerga.

Existe uma disciplina que torna isso confortável: **trate o erro e saia cedo**.
O caminho de erro entra no `if` e retorna. O caminho feliz continua na margem
esquerda, sem `else`, sem aninhamento. Olhe a solução de `Baixar`: três
saídas antecipadas e, no fim, uma linha de sucesso sem nenhuma indentação
extra.

**Por que erro sentinela e não código de erro?**
Um sentinela é um valor criado uma vez e comparado depois:

```go
var ErrEstoqueInsuficiente = errors.New("estoque insuficiente")
```

`errors.New` devolve um valor novo a cada chamada, então dois sentinelas com o
mesmo texto continuam sendo coisas diferentes. Isso é proposital: quem
identifica o erro é a **identidade do valor**, nunca a mensagem. Comparar
`err.Error() == "estoque insuficiente"` funciona hoje e quebra no dia em que
alguém melhorar o texto ou traduzir a mensagem.

**Por que a mensagem é minúscula e sem ponto final?**
Porque ela quase nunca aparece sozinha. Os erros se encaixam uns dentro dos
outros e formam uma frase. Rode o teste de demonstração deste módulo e olhe o
que sai:

```
vender 99 de "Pão de queijo": baixar 99 de "Pão de queijo", só tem 10: estoque insuficiente
```

Três camadas, separadas por dois-pontos, do contexto mais externo até a causa
raiz. Se cada pedaço começasse com maiúscula e terminasse com ponto, essa
linha viraria lixo. A convenção existe para a composição funcionar.

**O que é `%w`?**
É o verbo do `fmt.Errorf` que **envolve** um erro dentro de outro, guardando
uma referência ao original. `%v` só imprime o texto e joga a identidade fora.
Os dois compilam, os dois produzem uma mensagem parecida, e só um deixa
`errors.Is` funcionar.

**Então `errors.Is` em vez de `==`.**
`errors.Is` desembrulha o erro camada por camada e pergunta em cada nível se
é aquele sentinela. `==` só compara a camada de fora. O teste
`TestDemonstracaoIgualdadeDiretaNaoEnxergaErroEmbrulhado` já passa desde o
começo e existe para você ver os dois lado a lado.

## Pegadinhas

**Trocar `%w` por `%v` quebra tudo em silêncio.**
Compila, roda, a mensagem sai idêntica na tela, e `errors.Is` passa a
devolver `false`. Nenhuma ferramenta reclama por padrão. Se um teste de
`errors.Is` falhar sem motivo aparente, o primeiro lugar para olhar é o verbo
do `fmt.Errorf`.

**Erro ignorado não é erro de compilação.**
Isto é uma ironia genuína da linguagem: variável declarada e não usada
**impede** o programa de compilar, mas erro descartado passa liso.

```go
_ = Baixar(estoque, "Pão de queijo", 3) // compila, e você acabou de perder a venda
```

O compilador não te protege aqui. Quem protege é o `go vet` em alguns casos,
o revisor de código, e ferramentas de fora da biblioteca padrão como
`errcheck`. Ignorar erro de propósito é legítimo às vezes, e nesses casos
escreva um comentário dizendo por quê.

**Comparar mensagem de erro é bug esperando data para acontecer.**
`strings.Contains(err.Error(), "insuficiente")` funciona até alguém corrigir
uma vírgula. Os testes deste módulo usam `errors.Is` para identificar a causa,
e só usam o texto para checar se o **contexto** foi acrescentado, que é outra
coisa.

**Envolver sem acrescentar nada é ruído.**
`fmt.Errorf("erro: %w", err)` não ajuda ninguém. Envolva quando você tem
contexto que quem está embaixo não tinha: qual produto, qual quantidade,
qual arquivo. Quando você não tem nada a acrescentar, `return err` já está
certo.

## O exercício

Três sentinelas e três funções em `padaria/padaria.go`:

1. Crie `ErrProdutoNaoEncontrado`, `ErrEstoqueInsuficiente` e
   `ErrQuantidadeInvalida` com `errors.New`.
2. `BuscarProduto` devolve o produto ou envolve `ErrProdutoNaoEncontrado`
   com o nome procurado.
3. `Baixar` é o desafio do módulo 02, agora com erro de verdade. Verifique na
   ordem que o comentário da função descreve, e só escreva no estoque depois
   que todas as verificações passarem.
4. `Vender` chama as duas e envolve o que vier, empilhando contexto.

Preste atenção em dois testes. `TestBaixarRecusa` confere que o estoque não
mudou quando a operação falhou, porque uma função que erra no meio não pode
deixar sujeira. E `TestVenderPropagaOSentinelaDeDuasCamadasAbaixo` é o que
justifica o módulo: o erro nasce lá no fundo, é envolvido duas vezes, e quem
chamou ainda consegue perguntar qual foi a causa.

## O que ficou para depois

Você pode querer um erro que **carrega dados**, não só uma mensagem. Algo do
tipo "faltaram 89 unidades", para a padaria conseguir sugerir uma quantidade
menor em vez de só recusar. Isso se faz com um tipo próprio que tenha um
método `Error() string`, e se lê de volta com `errors.As`.

Métodos são o módulo 04. Interface, que é o que `error` realmente é, é o
módulo 05. A gente volta aqui.

Também existe `errors.Join`, para quando várias coisas dão errado ao mesmo
tempo e você quer reportar todas. Fica anotado.

## Desafio, sem teste pronto

Escreva `VenderVarios(cardapio []Produto, estoque Estoque, pedido map[string]int) (int, error)`
que vende vários produtos de uma vez e devolve o total.

A pergunta difícil não é escrever o laço. É decidir o que fazer quando o
terceiro item falha depois de os dois primeiros já terem saído do estoque.
Você para no primeiro erro e deixa o estoque pela metade? Você verifica tudo
antes de mexer em qualquer coisa? Você devolve todos os erros de uma vez?

Não existe resposta única, existe decisão de projeto, e ela precisa estar
escrita. Escolha uma, escreva num comentário por que escolheu, e siga.

## Se travar

A solução está em `solucoes/03-erros/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
