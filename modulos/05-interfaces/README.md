# Módulo 05 — Interfaces implícitas

A padaria vai parar de saber onde os produtos estão guardados. E o erro de
falta de estoque, prometido lá no fim do módulo 03, finalmente vai carregar os
números em vez de só uma frase.

```bash
go test ./modulos/05-interfaces/...
```

Duas coisas que você já usava sem saber ganham nome aqui. `error` é uma
interface, e sempre foi. O método `String` do módulo 04 satisfazia uma
interface da biblioteca padrão chamada `fmt.Stringer`, e é por isso que o
`fmt` achou ele sozinho.

## Por que Go faz assim

**Cadê o `implements`?**
Não existe essa palavra em Go. Um tipo satisfaz uma interface quando tem os
métodos que a interface pede, e ponto. Ninguém declara nada, ninguém registra
nada, e o tipo nem precisa saber que a interface existe.

Procure no código deste módulo: `*ErroDeEstoque` vira um `error` só por ter um
método `Error() string`. `*RepositorioEmMemoria` vira um `BuscadorDeProdutos`
só por ter um método `Buscar`. Os dublês de teste, escritos no arquivo de
teste, satisfazem a mesma interface sem uma linha de cerimônia.

**Quem é o dono da interface?**
Quem consome. Esta é a inversão que mais confunde quem vem de Java ou C#, e é
a coisa mais importante do módulo.

Lá você declara a interface junto da implementação, e a implementação anuncia
que a cumpre. Aqui a interface fica junto de **quem precisa do
comportamento**. `BuscadorDeProdutos` está declarada ao lado de `Vender`,
porque é `Vender` que precisa buscar produto. O repositório não sabe que ela
existe.

O que isso compra na prática: você consegue escrever uma interface para um
tipo que **já existe**, inclusive de um pacote de outra pessoa, sem tocar no
código dela. Precisa de um pedaço pequeno de uma biblioteca enorme? Declare
uma interface com os dois métodos que você usa, e pronto, o tipo dela já
satisfaz.

**Por que interfaces tão pequenas?**
Há um provérbio conhecido da comunidade, de Rob Pike, dizendo que quanto maior
a interface, mais fraca a abstração. `BuscadorDeProdutos` tem um método. As
interfaces mais usadas da biblioteca padrão, `io.Reader` e `io.Writer`, têm um
método cada, e são a razão de você conseguir ligar um arquivo, uma conexão de
rede, um buffer de memória e um compressor uns nos outros.

Interface grande é difícil de substituir, e o custo aparece justamente no
teste: escrever um dublê para uma interface de dez métodos é um castigo.
`Vender` pede o mínimo que precisa, e por isso `buscadorQuebrado`, no arquivo
de teste, cabe em quatro linhas.

**Aceite interface, devolva struct.**
Repare que `NovoRepositorioEmMemoria` devolve `*RepositorioEmMemoria`, o tipo
concreto, e não `BuscadorDeProdutos`. Quem chama fica com tudo que o tipo sabe
fazer, e ainda pode guardar numa variável de interface se quiser. Devolver
interface só tira opções de quem chama, e é um hábito importado de outras
linguagens que raramente ajuda em Go.

**Cadê a herança?**
Não tem, e não vai ter. O que existe é composição: você põe um tipo dentro de
outro e usa os métodos dele.

```go
type RepositorioComLog struct {
	BuscadorDeProdutos // embutido, sem nome de campo
	log *slog.Logger
}
```

Um tipo embutido assim empresta os métodos dele ao tipo de fora, o que faz
`RepositorioComLog` já satisfazer `BuscadorDeProdutos` de graça. Você
sobrescreve só o que quiser mudar. É o desafio deste módulo.

**E o `any`?**
`any` é apelido para `interface{}`, a interface sem método nenhum, que
portanto todo tipo satisfaz. Serve para os casos em que você realmente não
sabe o tipo, como o `fmt.Println`. Para tirar o valor de volta, existem
asserção de tipo e `switch` de tipo. Desde que Go ganhou genéricos, boa parte
do uso antigo de `any` virou parâmetro de tipo, que é o módulo 09.

## Pegadinhas

**Interface carregando ponteiro nil não é nil.**
Esta é a armadilha mais famosa da linguagem, e ela pega gente experiente.

```go
var erro *ErroDeEstoque // nil
var err error = erro    // err NÃO é nil
```

Uma interface guarda duas coisas: o tipo e o valor. O valor ali é nil, mas o
tipo é `*ErroDeEstoque`, e isso basta para `err == nil` dar false. Em código
real isso aparece assim:

```go
func Baixar(...) *ErroDeEstoque { ... }  // assinatura errada

if err := Baixar(...); err != nil {      // sempre entra aqui
```

A regra que evita o problema é simples: **funções devolvem `error`, nunca um
tipo de erro concreto**. `TestDemonstracaoPonteiroNilDentroDeInterfaceNaoEhNil`
já passa desde o começo e mostra isso funcionando.

**Método com receptor de ponteiro: só `*T` satisfaz, `T` não.**

```go
func (r *RepositorioEmMemoria) Buscar(...) (Produto, error)

var _ BuscadorDeProdutos = RepositorioEmMemoria{}  // não compila
```

```
RepositorioEmMemoria does not implement BuscadorDeProdutos
(method Buscar has pointer receiver)
```

É a mesma regra do módulo 04, vista de outro ângulo: um valor não carrega os
métodos de ponteiro do seu tipo. Aqui está a razão de a recomendação de não
misturar receptores existir, e é por isso que o `Error()` deste módulo tem
receptor de ponteiro e o `Baixar` devolve `&ErroDeEstoque{...}` com o e
comercial.

**A linha `var _ Interface = (*Tipo)(nil)` é sua amiga.**
Ela não guarda nada, não roda nada e não custa nada em tempo de execução.
Serve para o compilador reclamar **agora**, no arquivo onde o tipo mora, em
vez de reclamar lá longe, no ponto onde alguém tentou usar um pelo outro. Vale
para qualquer tipo seu que precise satisfazer uma interface importante.

**`errors.As` recebe o endereço da variável, não a variável.**

```go
var erroDeEstoque *ErroDeEstoque
errors.As(err, &erroDeEstoque)  // &, porque As precisa escrever ali dentro
```

Passar `erroDeEstoque` sem o `&` entra em pânico com uma mensagem sobre
ponteiro não-nil. `As` desce a cadeia de erros procurando algo do tipo certo e
guarda o achado na sua variável, então precisa do endereço dela.

**Não escreva interface antes de ter dois usos.**
A tentação de declarar `type ProdutoService interface` ao lado de cada struct,
por via das dúvidas, é forte para quem vem de Java. Em Go a interface custa
nada para ser criada depois, exatamente porque a satisfação é implícita: no
dia em que aparecer o segundo caso, você declara a interface e nenhum dos dois
tipos precisa mudar.

## O exercício

Sete coisas em `padaria/padaria.go`:

1. `Error()` e `Unwrap()` em `*ErroDeEstoque`. O `Baixar` que os usa já está
   escrito, no meio do arquivo, para você ver o consumidor antes.
2. `NovoRepositorioEmMemoria`, `Salvar`, `Buscar` e `Todos`.
3. `Vender`, que agora recebe `BuscadorDeProdutos` no lugar de `[]Produto`.

Dois testes merecem leitura antes de você começar.
`TestErroDeEstoqueContinuaSendoAchadoPorErrorsIs` é a dívida do módulo 03
sendo paga: uma linha de `Unwrap` faz todo o código antigo continuar
funcionando, e quem quiser os números usa `errors.As`. E
`TestVenderFuncionaComUmDubleDeTeste` mostra o retorno do investimento: a
mesma `Vender` roda contra o repositório real e contra um dublê de quatro
linhas, que ainda por cima consegue contar quantas vezes foi chamado.

## Desafio, sem teste pronto

Escreva um `RepositorioComLog` que embute um `BuscadorDeProdutos` e registra
cada busca antes de repassar:

```go
type RepositorioComLog struct {
	BuscadorDeProdutos
	linhas []string
}
```

Como o tipo embutido já traz o método `Buscar`, o seu tipo satisfaz a
interface antes mesmo de você escrever qualquer coisa. Sobrescreva `Buscar`
para anotar e depois chamar o de dentro.

Depois pense: quantas linhas você precisaria mudar em `Vender` para usar isso?
A resposta é a razão de a interface ter um método só.

## Se travar

A solução está em `solucoes/05-interfaces/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
