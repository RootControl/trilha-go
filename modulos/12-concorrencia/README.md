# Módulo 12 — Concorrência

A padaria assa vários pães ao mesmo tempo, e o estoque aprende a sobreviver a
isso.

```bash
go test -race ./modulos/12-concorrencia/...
```

Use `-race` neste módulo, sempre. Sem ele, uma corrida pode passar mil vezes e
falhar exatamente na produção, na sexta-feira à noite.

## A resposta da pergunta do módulo 10

Lá no fim daquele módulo ficou uma pergunta: o estoque do servidor é um `map`
compartilhado, o `net/http` atende cada requisição numa goroutine, e o que
acontece?

Acontece isto:

```bash
MOSTRAR_CORRIDA=1 go test -race -run Corrida ./modulos/12-concorrencia/...
```

```
WARNING: DATA RACE
    esperado 100, veio 45
```

Cem goroutines somaram um cada, e o mapa terminou com quarenta e cinco.
Cinquenta e cinco reposições evaporaram. E repare que **não houve erro**: o
programa rodou, respondeu, e mentiu.

Esse teste é pulado por padrão justamente porque a corrida é de propósito.

## Por que Go faz assim

**Goroutine não é thread.**
`go f()` e pronto. Uma goroutine começa com poucos kilobytes de pilha, que
cresce conforme precisa, e o runtime multiplexa milhares delas sobre um
punhado de threads do sistema. Criar cem mil goroutines é normal; criar cem
mil threads derruba a máquina.

Não existe `async`, não existe `await`, não existe cor de função. Uma função
comum, chamada com `go` na frente, roda concorrente. Toda a biblioteca padrão
funciona nas duas situações sem duas versões de cada coisa.

**O programa termina quando a `main` termina.**
Goroutina pendente não segura o programa. É a primeira surpresa de todo mundo,
e o motivo de o `sync.WaitGroup` existir.

**`sync.WaitGroup` conta trabalho pendente.**
`Add` antes de disparar, `Done` adiado dentro, `Wait` depois do laço. O `Add`
tem que vir **antes** do `go`: chamar `Add` dentro da goroutine é uma corrida
com o próprio `Wait`, que pode voltar antes de a goroutine sequer existir.

**Canal é sincronização, não só um cano.**
Num canal sem buffer, o envio só termina quando alguém recebe. Os dois lados
se encontram. `TestDemonstracaoCanalSemBufferSincroniza` deixa um envio parado
por vinte milissegundos e mostra que ele só segue depois da recepção.

Com buffer, o envio só bloqueia quando o buffer enche. O buffer é uma decisão
de capacidade, não uma otimização automática.

**Quem escreve fecha; quem lê nunca fecha.**
Fechar é a forma de dizer "não vem mais nada", e por isso é responsabilidade
de quem produz. Fechar duas vezes, ou enviar em canal fechado, entra em
pânico. Ler de canal fechado, não: devolve na hora o valor zero com o segundo
retorno em `false`, e `TestDemonstracaoCanalFechadoDevolveOValorZero` mostra
isso, inclusive que fechar não descarta o que já estava lá dentro.

**`select` espera em várias coisas ao mesmo tempo.**
E segue com a primeira que acontecer. Com `time.After` numa das opções, você
tem prazo. Com um `default`, você tem tentativa que não bloqueia.

**Índices diferentes não são corrida.**
`AssarTudo` deixa várias goroutines escreverem na mesma fatia sem mutex
nenhum, e o `-race` concorda. Cada uma escreve num índice próprio, ou seja, em
endereços diferentes. O que seria corrida é `append` concorrente, porque o
`append` mexe no cabeçalho da fatia, que é compartilhado.

Este é o padrão mais barato de coleta paralela: dimensione a fatia antes,
escreva por índice, junte depois.

**A variável do laço deixou de ser uma armadilha.**
Até o Go 1.21, `for i, p := range` criava **uma** variável reutilizada, e
todas as goroutines lançadas dentro do laço viam o último valor. Era a
pegadinha mais famosa da linguagem, e a correção era copiar a variável na
primeira linha do corpo. Desde o Go 1.22 cada volta tem variáveis novas, e o
código deste módulo depende disso. Se você encontrar `p := p` no começo de um
laço em código antigo, agora sabe o que era.

## Mutex ou canal?

As duas coisas resolvem problemas diferentes, e escolher errado dá trabalho.

Use **mutex** quando várias goroutines precisam mexer no **mesmo estado**. É o
caso do `EstoqueSeguro`: existe um estoque, e todo mundo escreve nele. Canal
aqui só complicaria.

Use **canal** quando o dado precisa **passar** de uma goroutine para outra. É o
caso do `AssarEmCanal`: cada fornada é produzida num lugar e consumida em
outro.

O provérbio da comunidade diz para compartilhar memória comunicando, em vez de
comunicar compartilhando memória. É um bom conselho e virou dogma em alguns
lugares: um contador protegido por mutex é mais simples e mais rápido que um
contador atrás de um canal.

E, para contador puro e sinalizador, `sync/atomic` é mais leve que mutex. O
forno de teste deste módulo usa `atomic.Int64`.

## Pegadinhas

**Struct com mutex não pode ser copiada.**
Depois que um mutex foi usado, copiar a struct que o contém copia o estado da
trava, e as duas cópias passam a proteger coisas diferentes. É por isso que
todos os métodos de `EstoqueSeguro` têm receptor de ponteiro, **inclusive os
que só leem**. O `go vet` avisa quando você copia um mutex, e é um dos avisos
que nunca se ignora.

**Verificar e agir em travas separadas.**
Ler o estoque com `RLock`, soltar, e depois escrever com `Lock` parece
razoável e está errado: entre soltar e travar de novo, outra goroutine passa
pela mesma verificação. `TestEstoqueSeguroNaoDeixaOEstoqueNegativo` dispara
duzentas tentativas contra quarenta unidades e exige exatamente quarenta
vendas.

**Segurar o mutex mais tempo do que o necessário.**
A validação de `quantidade <= 0` fica **fora** da trava, porque não toca no
estado. Mutex segurado durante trabalho lento, ou durante uma chamada de rede,
transforma um programa concorrente num programa sequencial mais lento que o
original.

**Goroutine vazada.**
Uma goroutine parada num envio que ninguém vai receber fica lá para sempre,
segurando memória. `AssarEmCanal` usa canal sem buffer, então quem chama
**precisa** consumir o canal. Se o consumidor desistir no meio, as goroutines
restantes ficam penduradas. Resolver isso direito é contexto, no módulo 13.

**Esquecer de fechar o canal.**
O `for range` de quem lê espera para sempre, e o teste morre por prazo
esgotado em vez de falhar com mensagem. Se um teste seu ficar pendurado até o
`panic: test timed out`, procure um `close` faltando.

**Travar todo mundo de propósito.**
Se todas as goroutines ficarem esperando, o runtime percebe e derruba:

```
fatal error: all goroutines are asleep - deadlock!
```

É uma mensagem generosa. Ela só aparece quando **nenhuma** goroutine pode
progredir; um travamento parcial não é detectado e vira um programa que
simplesmente não responde.

**`t.Fatal` dentro de goroutine não funciona.**
Foi prometido no módulo 06 e é a hora. `t.Fatal` chama `runtime.Goexit`, que
encerra a **goroutine que a chamou**, não o teste. O teste segue e pode até
passar. Dentro de goroutine, use `t.Error`, ou mande o problema por canal e
falhe na goroutine do teste.

## O exercício

Sete implementações em `padaria/padaria.go`:

1. `EstoqueSeguro`, com `NovoEstoqueSeguro`, `Repor`, `Quantidade` e `Baixar`.
   É a versão do módulo 02 que sobrevive ao módulo 10.
2. `AssarTudo`, com `WaitGroup` e fatia por índice.
3. `AssarEmCanal`, devolvendo o canal na hora e fechando no fim.
4. `Coletar`, com `for range` sobre canal.
5. `EsperarComPrazo`, com `select`.
6. `ContarPorProduto`, que não tem concorrência nenhuma e é a resposta certa.
   Nem todo problema perto de goroutine vira goroutine.

Comece pelo `EstoqueSeguro`, e rode sempre com `-race`.

## Desafio, sem teste pronto

**Limite o forno.** Hoje `AssarTudo` dispara uma goroutine por pedido: com dez
mil pedidos, dez mil fornadas ao mesmo tempo. Faça uma versão que assa no
máximo N ao mesmo tempo, usando um canal com buffer como senha de entrada:

```go
senhas := make(chan struct{}, n)
// antes de assar: senhas <- struct{}{}
// depois de assar: <-senhas
```

`struct{}` ocupa zero byte, e é o valor convencional quando o que importa é o
sinal e não o dado.

**Meça.** Compare `AssarTudo` com a versão limitada, para mil pedidos e um
forno de um milissegundo, variando N. Existe um ponto em que aumentar N para
de ajudar. Descobrir onde ele fica, e por quê, é metade do que se sabe sobre
paralelismo.

**Quebre de propósito.** Tire o mutex do `Repor` e rode com `-race`. Depois
troque a trava única do `Baixar` por duas travas separadas e rode
`TestEstoqueSeguroNaoDeixaOEstoqueNegativo` umas vinte vezes com `-count=20`.
Ver o bug acontecer vale mais do que ler sobre ele.

## Se travar

A solução está em `solucoes/12-concorrencia/`. Dúvida sobre o enunciado é bug
do material: abra uma issue.
