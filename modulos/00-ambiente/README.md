# Módulo 00 — Ambiente e o primeiro teste passando

Você já programa em alguma linguagem. Este módulo não vai te ensinar o que é
uma função. Ele vai te mostrar como um projeto Go se parece por dentro, e vai
te deixar com dois testes verdes.

## O que você vai fazer

Abrir `padaria/padaria.go`, implementar duas funções e fazer os testes
passarem. Só isso. O objetivo aqui é o ambiente, não o algoritmo.

```bash
make testar
```

Você vai ver os testes falharem. Isso é o esperado, e é assim que todo módulo
começa. O arquivo de teste já está escrito, e ele é o enunciado do exercício.

## Por que Go faz assim

**Cadê o `package.json`, o `requirements.txt`, o `composer.json`?**
É o `go.mod`, na raiz do repositório. Ele tem três linhas e uma delas é o nome
do módulo. A biblioteca padrão cobre HTTP, JSON, testes, criptografia e banco
de dados, então a lista de dependências de um projeto Go costuma ser curta o
suficiente para caber na tela. Você não vai instalar nada para terminar este
módulo.

**Cadê o Jest, o pytest, o PHPUnit?**
Não existe. Testar faz parte da linguagem. Qualquer arquivo terminado em
`_test.go` é um teste, qualquer função `TestAlgumaCoisa(t *testing.T)` é um
caso, e `go test` roda tudo. Não há framework para escolher, não há
configuração para escrever, e por isso projetos Go tendem a ter testes.

**Cadê o Prettier, o ESLint, o Black?**
É o `gofmt`, que vem junto e não tem opções. Ninguém discute chaves, aspas ou
tamanho de indentação em Go, porque não há o que discutir. Rode
`make formatar` e siga a vida.

**Por que dinheiro em centavos e não em `float64`?**
Porque `0.1 + 0.2` não dá `0.3` em ponto flutuante, em nenhuma linguagem. Um
número quebrado de centavo vira diferença de caixa no fim do mês. A padaria
guarda `450` e imprime `R$ 4,50`. Guardar dinheiro como inteiro é a regra em
qualquer sistema financeiro sério, e você vai carregar essa decisão pelo curso
inteiro.

## Pegadinhas

Estas duas vão te pegar hoje, então é melhor te pegarem agora.

**Import não usado é erro de compilação.** Não é aviso. O programa não compila.
Se você importar `fmt` e ainda não usar, o Go recusa o arquivo. A mesma coisa
vale para variável local declarada e não usada. Parece rigidez gratuita e é
uma escolha deliberada: código morto não entra no repositório porque não passa
pelo compilador.

**Só o que começa com letra maiúscula sai do pacote.** `Saudacao` é visível de
fora, `saudacao` não seria. Não existe `public`, `private` nem `export`. A
primeira letra é o modificador de visibilidade. Se um teste não enxerga sua
função, confira a maiúscula antes de qualquer outra coisa.

## O exercício

Em `padaria/padaria.go`:

1. `Saudacao(nome string) string` devolve
   `Bom dia, Maria! Bem-vindo à Padaria do Seu Zé.` e troca nome vazio por
   `cliente`.
2. `FormatarPreco(centavos int) string` transforma `450` em `R$ 4,50`.

Dica para a segunda: divisão inteira te dá os reais, o resto te dá os
centavos, e o verbo `%02d` do `fmt.Sprintf` coloca o zero à esquerda.

Quando os dois passarem:

```bash
make testar
```

## Desafio, sem teste pronto

Escreva `Troco(pagoEmCentavos, precoEmCentavos int) (int, error)` que devolve o
troco e um erro quando o pagamento não cobre o preço. Você ainda não viu erros
em Go, então resolva do jeito que achar. No módulo 03 a gente volta aqui e
refaz isso do jeito idiomático. Guardar a sua primeira tentativa para comparar
depois é metade do aprendizado.

## Se travar

A solução está em `solucoes/00-ambiente/`. Olhe depois de tentar. Se algo no
enunciado estiver ambíguo, abra uma issue: dúvida de aluno vira melhoria de
material.
