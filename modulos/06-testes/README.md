# Módulo 06 — Testes de verdade

Este módulo tem uma forma diferente. A padaria dos módulos anteriores chega
aqui **inteira e funcionando**: repositório, estoque, erros, venda. O que
falta são as peças que tornam esse código testável, e é isso que você escreve.

Você vem lendo testes há cinco módulos. Agora vale a pena entender o que
estava acontecendo neles.

```bash
go test ./modulos/06-testes/...
```

O arquivo `padaria/padaria_test.go` deste módulo é material de leitura, não só
de execução. Ele usa e comenta cada técnica que este README explica. Leia de
cima a baixo antes de implementar.

## Por que Go faz assim

**Cadê o Jest, o pytest, o PHPUnit, o JUnit?**
Não existem, e não fazem falta. Testar é parte da linguagem: arquivo terminado
em `_test.go`, função `TestAlgumaCoisa(t *testing.T)`, e `go test` acha e roda.
Sem instalar nada, sem arquivo de configuração, sem escolher entre três
bibliotecas concorrentes. É por isso que projetos Go tendem a ter testes: o
atrito para começar é zero.

**Cadê o `assertEquals`?**
Também não tem. Você escreve `if` e chama `t.Errorf`. Parece retrocesso e não
é: a asserção genérica produz mensagem genérica, e o `if` produz a mensagem
que **você** escolheu. Compare o que uma biblioteca diria com o que os testes
deste curso dizem:

```
expected 495 but got 450
preço = 450, esperado 495; o receptor precisa ser *Produto
```

A segunda foi escrita por alguém que sabia qual erro o leitor provavelmente
cometeu. Nenhuma biblioteca escreve isso por você.

**Cadê a biblioteca de mock?**
Você escreve o dublê à mão, como fez no módulo 05. `RelogioFixo` tem cinco
linhas. `buscadorQuebrado`, do módulo passado, tem quatro. Isso só é
confortável porque as interfaces em Go são pequenas, então o custo de escrever
uma implementação falsa é baixo. Framework de mock existe em outras linguagens
principalmente porque lá as interfaces são grandes.

**Por que tabela em todo teste?**
Porque separa o **que** está sendo testado do **como** o teste roda. A tabela
vira a especificação legível, o laço vira mecânica. Adicionar um caso é
acrescentar uma linha, e não copiar e colar vinte. A biblioteca padrão do Go é
escrita assim, e é de lá que o hábito veio.

**Por que o teste mora ao lado do código?**
`padaria_test.go` fica na mesma pasta e no mesmo pacote de `padaria.go`. Isso
dá acesso ao que não é exportado, o que é ótimo para testar detalhe interno.
Existe a variante `package padaria_test`, no mesmo diretório, que enxerga só a
API pública, e serve para você testar o pacote como um usuário dele testaria.
Use a primeira por padrão e a segunda quando quiser garantir que a API pública
basta.

## As ferramentas do `t`

| Chamada | Para quê |
|---------|----------|
| `t.Error` / `t.Errorf` | marca falha e **continua** o teste |
| `t.Fatal` / `t.Fatalf` | marca falha e **para** ali |
| `t.Run` | subteste com nome próprio, isolado e filtrável |
| `t.Helper` | a falha aponta para quem chamou, não para o auxiliar |
| `t.Cleanup` | roda no fim do teste, mesmo se ele falhar |
| `t.TempDir` | pasta temporária apagada sozinha, essencial no módulo 07 |
| `t.Parallel` | roda este teste junto com os outros marcados assim |
| `t.Skip` | pula com motivo registrado |
| `t.Log` | só aparece com `-v`, ou quando o teste falha |

A escolha entre `Error` e `Fatal` é a que mais importa no dia a dia. Use
`Fatal` quando as verificações seguintes não fariam sentido, como depois de um
erro no preparo do cenário. Use `Error` quando você quer ver **todas** as
diferenças de uma vez, e não descobrir uma por execução.

## Os comandos

```bash
go test ./...                      # tudo
go test -v ./...                   # mostra cada teste e cada t.Log
go test -run TestVender ./...      # só os que casam com o padrão
go test -run 'TestDiferenca/nome'  # desce até um subteste específico
go test -race ./...                # detector de corrida, essencial no módulo 12
go test -cover ./...               # percentual de linhas cobertas
go test -bench=. -benchmem ./...   # roda os benchmarks
go test -count=1 ./...             # ignora o cache
```

Duas coisas que valem saber agora. O `-race` é praticamente de graça e acha
uma classe inteira de bug que nenhum teste comum acha; ele volta a aparecer no
módulo 12. E o `go test` **guarda o resultado em cache**: rodar de novo sem
mudar nada imprime `(cached)` e não executa nada. Isso confunde muita gente
que está tentando reproduzir um teste instável. `-count=1` desliga o cache.

## Exemplo executável

`ExampleDiferencaEntreProdutos`, no arquivo de teste, é duas coisas ao mesmo
tempo: aparece na documentação do pacote e é executado pelo `go test`, que
compara a saída com o bloco `// Output:`. Documentação que mente quebra o
build. Nenhuma outra linguagem popular tem isso embutido, e é um dos melhores
motivos para escrever exemplo em vez de parágrafo.

## Sobre cobertura

Rodando `go test -cover` contra a solução deste módulo, o número dá em torno
de setenta e sete por cento. Isso é bom, e perseguir cem seria desperdício.

Cobertura mede quais linhas **executaram**, não quais estão **corretas**. Um
teste que chama todas as funções e não verifica nada dá cem por cento. Use o
número para achar áreas esquecidas, não como meta. O jeito honesto de saber se
o seu teste presta é outro, e está no desafio deste módulo.

## Pegadinhas

**Esquecer `t.Helper()` no auxiliar.**
Sem ele, toda falha aponta para a linha de dentro do auxiliar, e você fica sem
saber qual dos vinte casos quebrou. É uma linha, e ela transforma a saída do
teste de inútil em útil.

**Teste que depende do relógio, da rede ou da ordem de um map.**
São as três fontes clássicas de teste instável, aquele que passa dez vezes e
falha na décima primeira, sempre no CI e nunca na sua máquina. Este módulo
inteiro é sobre a primeira: `time.Now` sai da regra de negócio e entra como
parâmetro. Ordem de map você já resolveu no módulo 02.

**O cache do `go test` fingindo que rodou.**
Você muda alguma coisa fora do pacote, roda de novo, vê `(cached)` e conclui
que o comportamento não mudou. Rode com `-count=1` antes de tirar conclusão.

**Nome errado, teste invisível.**
O arquivo precisa terminar em `_test.go`. A função precisa começar com `Test`
maiúsculo, seguido de outra maiúscula ou de nada, e receber `*testing.T`.
`testVender` ou `TestvenderPaes` simplesmente não rodam, e o `go test` não
avisa que existe uma função ali sendo ignorada.

**`t.Fatal` dentro de goroutine não faz o que você espera.**
Ele encerra a goroutine, não o teste. Isso vira problema de verdade no módulo
12; por enquanto, guarde a informação.

## O exercício

Sete implementações em `padaria/padaria.go`, todas a serviço da testabilidade:

1. `RelogioDoSistema.Agora` e `RelogioFixo.Agora`, a interface que tira
   `time.Now` da regra de negócio.
2. `NovoPedido`, que carimba a hora que o relógio disser.
3. `RegistradorEmMemoria.Registrar` e `.Linhas`, um espião que guarda o que
   recebeu. `Linhas` devolve cópia, e o teste confere isso.
4. `VenderComRegistro`, que dá ao espião algo para observar.
5. `DiferencaEntreProdutos`, que é, em miniatura, o que uma biblioteca de
   asserção faz. Depois de escrever, você entende por que muita gente em Go
   decide que não precisa de uma.

## Desafio, sem teste pronto

Agora inverte. Volte em `modulos/04-metodos/`, apague mentalmente o arquivo de
teste, e escreva do zero o seu próprio `caixa_test.go` para o tipo `Caixa`:
`Registrar`, `TicketMedio` e `String`.

Quando terminar, faça o teste de mutação, que é a única forma honesta de saber
se um teste presta: **quebre o código de propósito** e veja se o seu teste
percebe. Troque `c.Vendas++` por `c.Vendas += 2`. Troque a divisão do ticket
médio por multiplicação. Tire o `if c.Vendas == 0`. Cada sabotagem dessas
precisa deixar pelo menos um teste vermelho.

Sabotagem que não quebra nenhum teste é um buraco na sua suíte, e nenhum
percentual de cobertura vai te contar isso.

## Se travar

A solução está em `solucoes/06-testes/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
