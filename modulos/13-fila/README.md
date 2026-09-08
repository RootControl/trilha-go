# Módulo 13 — Fila de pedidos, contexto e desligamento

No módulo 12 a padaria aprendeu a assar em paralelo. Aqui ela aprende a
**parar**: a desistir de um trabalho que demorou demais, e a fechar a loja sem
jogar fora o pão que já está no forno.

```bash
go test -race ./modulos/13-fila/...
```

## O que é contexto, afinal

`context.Context` é uma interface com quatro métodos, e ela carrega três
coisas:

- um **sinal de cancelamento**, no canal devolvido por `Done()`;
- um **prazo**, opcional;
- alguns **valores**, também opcionais.

É só isso. Contexto não é injeção de dependência, não é sessão de usuário e
não é um lugar conveniente para guardar coisas.

O que ele resolve é um problema real: quando uma requisição HTTP chega, ela
pode disparar uma consulta ao banco, que dispara uma chamada a outro serviço,
que dispara três goroutines. Se o cliente fecha o navegador, tudo isso vira
trabalho jogado fora, e sem um sinal comum ninguém tem como avisar ninguém.

## Por que Go faz assim

**As convenções, todas de uma vez.**

- Contexto é sempre o **primeiro** parâmetro, chamado `ctx`.
- Nunca vai guardado dentro de uma struct.
- Nunca é `nil`. Se não tem um, use `context.Background()`.
- `context.TODO()` existe para marcar "ainda não sei qual contexto vai aqui",
  e é diferente de `Background()` só na intenção.

**Cancelamento desce a árvore e nunca sobe.**
Cada contexto derivado é filho de outro. Cancelar um cancela ele e todos os
descendentes, e não toca no pai.
`TestDemonstracaoCancelarFilhoNaoAfetaOPai` monta pai, filho e neto e mostra
isso funcionando.

É essa propriedade que torna `AssarComPrazo` seguro: ele impõe um prazo à sua
própria chamada sem afetar o contexto de quem chamou.

**`defer cancelar()`, sempre.**
`WithCancel`, `WithTimeout` e `WithDeadline` devolvem dois valores, e o
segundo é obrigatório chamar, **mesmo no caminho de sucesso**. Ele libera o
temporizador interno e desliga o filho do pai. Esquecer é vazamento, e o
`go vet` reclama.

**Tornar uma espera cancelável é sempre o mesmo `select`.**

```go
select {
case <-temporizador.C:
	// terminou
case <-ctx.Done():
	return ctx.Err()
}
```

Um `time.Sleep` comum é surdo: começou, vai até o fim.
`TestAssarDesisteQuandoOContextoEhCancelado` põe uma espera de dez segundos e
exige que ela morra em menos de um.

**Os dois erros de contexto são diferentes.**
`context.Canceled` significa que alguém desistiu. `context.DeadlineExceeded`
significa que demorou demais. Um serviço reage diferente a cada um: o primeiro
não é problema seu, o segundo talvez seja. Devolva sempre `ctx.Err()` em vez
de um erro próprio, para quem chamou poder usar `errors.Is` contra os
sentinelas da biblioteca padrão.

**Fila com limite não é o mesmo que paralelismo solto.**
O `AssarTudo` do módulo 12 disparava uma goroutine por pedido. Com dez mil
pedidos, dez mil fornadas ao mesmo tempo, e o recurso escasso vira gargalo ou
estoura. `ProcessarFila` usa um número fixo de trabalhadores lendo de um canal
comum: quem termina antes pega a próxima tarefa, e não é preciso dividir o
trabalho de antemão.

`TestProcessarFilaRespeitaOLimiteDeTrabalhadores` manda vinte pedidos com três
trabalhadores e verifica que nunca houve quatro fornadas no forno.

**`r.Context()` já vem cancelado quando o cliente vai embora.**
No servidor do módulo 10, cada requisição carrega um contexto que o `net/http`
cancela se a conexão cair. Repassar esse contexto adiante é o que faz o
servidor parar de trabalhar por alguém que já fechou o navegador.

## Desligamento com graça

Rode e veja, que vale mais do que ler:

```bash
go run ./modulos/13-fila/cmd/servidor
```

Em outro terminal, dispare uma fornada de cinco segundos e, com ela no ar,
aperte Ctrl+C no primeiro:

```bash
curl "localhost:8080/assar?produto=Broa&quantidade=5"
```

O que acontece: o servidor para de aceitar conexões novas na hora, a
requisição em andamento **termina normalmente**, com status 200, e só então o
processo sai. Uma conexão nova tentada nesse meio-tempo é recusada.

O `main.go` já vem pronto e comentado. Três peças:

`signal.NotifyContext` devolve um contexto cancelado quando chega um sinal.
Trate `os.Interrupt`, que é o Ctrl+C, e `syscall.SIGTERM`, que é o que um
orquestrador de contêiner manda antes de matar o processo. Ignorar o segundo
significa perder requisições a cada implantação.

`servidor.Shutdown(ctx)` para de aceitar conexões e espera as em andamento.
Se o prazo do contexto estourar antes, ele desiste e corta o resto. A
diferença para `servidor.Close()` é exatamente essa: `Close` corta tudo na
hora, no meio da resposta de quem estava sendo atendido.

E o contexto do desligamento nasce de `context.Background()`, **não** do
contexto do sinal. Aquele já está cancelado, e um filho dele daria zero tempo
para terminar as requisições.

## Pegadinhas

**`time.After` dentro de laço vaza.**
O canal que ele devolve não é coletado enquanto o prazo não vencer. Numa
função que desiste rápido e é chamada muito, isso deixa um rastro de
temporizadores vivos. Use `time.NewTimer` com `defer temporizador.Stop()`,
como na solução deste módulo. Para um `select` único e curto, `time.After`
é aceitável, e é por isso que ele aparece no módulo 12.

**Esquecer `defer cancelar()`.**
Vazamento silencioso. O `go vet` avisa, e é mais um motivo para rodar `go vet`
junto com os testes.

**Guardar contexto numa struct.**
A tentação aparece quando a função tem muitos parâmetros. O problema é que um
contexto vale para **uma operação**, e uma struct costuma viver mais que isso:
você acaba usando um contexto já cancelado, ou um que nunca cancela. A exceção
conhecida é uma struct que representa a própria operação, e mesmo aí o campo
se chama `ctx` e é documentado.

**Aceitar o contexto e não olhar para ele.**
Receber `ctx` e nunca consultar `ctx.Done()` é o mesmo que não recebê-lo. Se
`ProcessarFila` não tivesse o braço de `ctx.Done()` no alimentador, o
cancelamento não teria efeito nenhum sobre o enfileiramento, e o alimentador
ainda ficaria pendurado tentando entregar tarefa a trabalhadores que já foram
embora. Goroutine vazada é quase sempre isso.

**Abusar de `context.WithValue`.**
Serve para dado que acompanha a requisição e atravessa camadas que não se
conhecem, como identificador de rastreamento ou usuário autenticado. Não serve
para passar dependência: conexão de banco, configuração e relógio entram como
parâmetro ou como campo, onde o compilador confere.

E a chave **precisa** ser de um tipo próprio e não exportado. Com uma string
comum, outro pacote pode usar a mesma string e sobrescrever o seu valor. Com
`type chaveDeContexto string` privado, ninguém de fora consegue nem construir
a chave.

**Confundir os dois erros.**
`errors.Is(err, context.Canceled)` é falso quando o prazo estourou.
`TestDemonstracaoOsDoisErrosDeContextoSaoDiferentes` deixa isso explícito.

## O exercício

Cinco implementações em `padaria/padaria.go`:

1. `FornoLento.Assar`, a espera cancelável com `select`.
2. `AssarComPrazo`, com `WithTimeout` e `defer cancelar()`.
3. `ProcessarFila`, o pool de trabalhadores.
4. `ContextoComCliente` e `ClienteDoContexto`, com chave de tipo privado.

Comece por `Assar`: ela é curta e todo o resto do módulo depende dela.

`ProcessarFila` é a maior função do curso até aqui, e o comentário dela
descreve as cinco peças em ordem. Escreva uma de cada vez e rode os testes no
meio, começando por `TestProcessarFilaPreservaAOrdem` com `-run Ordem`.

## Desafio, sem teste pronto

**Devolva o que deu tempo de assar.** Hoje `ProcessarFila` devolve `nil` no
cancelamento. Numa fila de verdade, as fornadas já prontas costumam valer
alguma coisa. Mude a assinatura para devolver as parciais junto com o erro, e
decida como quem chama sabe quais posições ficaram vazias.

**Junte a fila com o servidor.** Faça o endereço `/assar` enfileirar em vez de
assar direto, com um pool de três trabalhadores compartilhado por todas as
requisições. Repare que agora existem dois contextos em jogo, o da requisição e
o do servidor, e pense em qual deles cada trabalho deve obedecer.

**Meça o desligamento.** Suba o servidor, dispare vinte requisições lentas,
mande `SIGTERM` e cronometre. Depois troque `Shutdown` por `Close` e faça o
mesmo. A diferença entre os dois é a diferença entre uma implantação que
ninguém percebe e uma que aparece no gráfico de erros.

## Se travar

A solução está em `solucoes/13-fila/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
