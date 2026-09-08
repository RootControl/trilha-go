# Módulo 08 — Um programa de terminal

A padaria vira um programa que alguém consegue rodar. É a primeira entrega
usável do curso.

```bash
go run ./modulos/08-cli/cmd/padaria cardapio
go run ./modulos/08-cli/cmd/padaria -arquivo /tmp/pedidos.json vender "Pão de queijo" 3
go run ./modulos/08-cli/cmd/padaria -arquivo /tmp/pedidos.json pedidos
go run ./modulos/08-cli/cmd/padaria -h
```

Abra `cmd/padaria/main.go` antes de qualquer coisa. Ele já está pronto, tem
nove linhas de código e não decide nada. Esse é o assunto do módulo.

```bash
go test ./modulos/08-cli/...
```

## Por que Go faz assim

**Cadê o ponto de entrada?**
Um pacote chamado `main` com uma função `main()` sem parâmetros e sem retorno.
`go run` compila e executa; `go build` deixa um **binário único e estático**,
sem interpretador, sem `node_modules`, sem ambiente virtual. Você copia o
arquivo para o servidor e ele roda. Compilar para outro sistema é mudar duas
variáveis:

```bash
GOOS=linux GOARCH=amd64 go build -o padaria ./modulos/08-cli/cmd/padaria
```

**Por que o programa mora em `cmd/padaria/` e não na raiz?**
É a convenção da comunidade. Uma pasta por binário dentro de `cmd/`, e a
lógica em pacotes de biblioteca fora dela. Assim o mesmo código serve para
mais de um programa, e principalmente: **o pacote `main` é o único lugar que
não dá para testar direito**, então ele precisa ser minúsculo.

**O padrão do main fino.**
Compare as duas formas. A primeira é a que aparece em tutorial:

```go
func main() {
	flag.Parse()
	produtos := carregar()
	fmt.Println(produtos)
}
```

Nada disso é testável. `flag.Parse` lê o `os.Args` do processo, `fmt.Println`
escreve no `os.Stdout` do processo, e um teste não tem como interferir em
nenhum dos dois sem truque.

A segunda forma, que é a deste módulo, empurra tudo para uma função comum:

```go
func Executar(args []string, saida io.Writer, erros io.Writer, relogio Relogio) error
```

Agora o teste passa os argumentos que quiser, lê a saída de um `bytes.Buffer`,
e o `main` fica só com a tradução entre esse mundo e o sistema operacional:
pegar `os.Args[1:]`, entregar `os.Stdout` e `os.Stderr`, e transformar erro em
código de saída.

Olhe o auxiliar `executar` no arquivo de teste. Ele roda o **programa inteiro**
dentro do teste, sem criar processo, sem arquivo em caminho fixo e sem relógio
real. Nenhum teste deste módulo executa um binário.

**Por que dois canais de saída?**
`os.Stdout` é para **dados**, `os.Stderr` é para **recado**. A regra existe
porque a saída padrão é o que atravessa um cano:

```bash
padaria -json pedidos | jq '.[].cliente'
```

Se o texto de ajuda, os avisos ou as mensagens de erro fossem para a saída
padrão, esse comando quebraria. É por isso que os testes deste módulo checam
que nada foi para a saída quando o comando falhou.

**Por que código de saída?**
É como um script em volta descobre o que aconteceu, sem ler texto. Zero é
sucesso, qualquer outro valor é falha. Repare que `-h` devolve zero: pedir
ajuda é uso correto do programa, e um script que roda `padaria -h` não pode
achar que deu errado.

**Cadê o Cobra, o Click, o Commander?**
O pacote `flag` da biblioteca padrão resolve a maior parte dos casos, e é o
que este módulo usa. Bibliotecas como o Cobra existem para árvores grandes de
subcomandos, com completação de shell e ajuda hierárquica. Comece com o
`flag`; troque quando o `flag` doer, não antes.

**Por que a data de formatação é 2006-01-02?**
Porque Go não usa uma gramática tipo `yyyy-MM-dd`. Você escreve **a data de
referência**, formatada do jeito que você quer a sua. E a data de referência é
uma sequência fácil de lembrar:

```
mês 1, dia 2, hora 3, minuto 4, segundo 5, ano 6, fuso 7
```

Ou seja, `01/02 03:04:05PM '06 -0700`. `TestDemonstracaoDataDeReferencia` já
passa desde o começo e mostra quatro layouts funcionando.

## Pegadinhas

**`fmt.Println` no lugar de `fmt.Fprintln`.**
`Println` escreve no `os.Stdout` do processo, sempre, e torna a função
impossível de testar. `Fprintln` escreve onde mandarem. A diferença é uma
letra e ela decide se o seu código é testável.

**`flag.Parse()` do pacote é uma variável global disfarçada.**
Ela lê o `os.Args` do processo e escreve no `os.Stderr` do processo. Além de
intestável, ela só pode ser chamada uma vez por processo, então dois testes
que quisessem argumentos diferentes brigariam. `flag.NewFlagSet` recebe o que
você der.

**As opções vêm antes do comando.**
O `flag` para de interpretar no primeiro argumento que não começa com traço.
Então isto funciona:

```bash
padaria -json pedidos
```

e isto **não**, porque `-json` vira um argumento posicional comum:

```bash
padaria pedidos -json
```

Ferramentas com subcomando de verdade resolvem isso criando um `FlagSet` por
subcomando. É parte do desafio.

**`conjunto.String` devolve um ponteiro.**
`caminho := conjunto.String(...)` te dá um `*string`, que só tem valor
**depois** do `Parse`. Usar `caminho` onde você queria `*caminho` costuma
render um erro de compilação claro, mas em `Fprintf` com `%v` passa liso e
imprime um endereço.

**`os.Exit` não roda os `defer` pendentes.**
Arquivo não fechado, trava não liberada, buffer não descarregado. É por isso
que `os.Exit` aparece uma vez só, na última linha do `main`, e nunca no meio
da lógica. Funções devolvem erro; quem encerra o processo é o `main`.

**Erro de uso e erro de negócio não são a mesma coisa.**
`ErrUsoInvalido` significa que quem chamou o programa errou a linha de
comando. `ErrEstoqueInsuficiente` significa que a padaria recusou a venda. Um
script em volta reage de formas diferentes a cada um, e por isso os dois
sentinelas existem separados.

## O exercício

Três funções em `padaria/padaria.go`:

1. `EscreverCardapio`, em texto ou em JSON.
2. `EscreverPedidos`, idem, com a linha formatada e o caso de lista vazia.
3. `Executar`, que monta o `FlagSet`, despacha os três comandos e devolve
   erro em vez de encerrar o processo.

Tudo que vem dos módulos anteriores já está pronto no arquivo, inclusive a
persistência do módulo 07 e o relógio do módulo 06.

Comece por `EscreverCardapio`, que é a mais curta, e rode
`go test -run Cardapio ./modulos/08-cli/...` para trabalhar num teste só.

## Uma limitação de propósito

O estoque não sobrevive entre execuções. Cada chamada monta um
`EstoquePadrao()` novo, então você pode vender dez pães de queijo, rodar de
novo e vender mais dez. Os pedidos persistem, o estoque não.

Isso está errado, e está errado de propósito: é o desafio. Repare que o
programa parece funcionar perfeitamente até alguém pensar no assunto, o que é
uma boa lição sobre o que um teste verde não garante.

## Desafio, sem teste pronto

Duas coisas, em ordem de dificuldade.

**Persista o estoque.** Grave num arquivo, como você já faz com os pedidos, e
carregue no começo de cada execução. Decida o que acontece quando o arquivo
não existe, e o que acontece quando ele existe mas está desatualizado em
relação ao cardápio.

**Dê subcomandos de verdade ao programa.** Hoje as opções precisam vir antes
do comando. Crie um `flag.NewFlagSet` por subcomando, com opções próprias, e
faça o `Executar` escolher o conjunto certo depois de ler o primeiro
argumento. Assim `padaria vender -cliente Maria "Broa" 2` passa a funcionar, e
cada comando ganha a sua própria linha de ajuda.

## Se travar

A solução está em `solucoes/08-cli/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
