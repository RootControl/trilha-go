# Módulo 07 — JSON, arquivos e `defer`

Até aqui, tudo que a padaria sabia morria quando o programa terminava. Neste
módulo os pedidos passam a sobreviver ao desligamento.

O assunto secundário é mais importante que o principal: repare que
`SalvarPedidos` recebe um `io.Writer`, e não um nome de arquivo.

```bash
go test ./modulos/07-json/...
```

## Por que Go faz assim

**Cadê o `JSON.stringify`, o `json.dumps`, o `json_encode`?**
É o pacote `encoding/json`, da biblioteca padrão. A diferença de fundo é que
Go não tem objeto genérico: você decodifica **para dentro de uma struct**, com
tipos definidos por você. Isso obriga a declarar o formato antes, e em troca o
compilador passa a te ajudar com o conteúdo do arquivo.

**Cadê o decorador, o atributo, a anotação?**
São as **etiquetas de struct**, aquela string entre crases depois do tipo do
campo:

```go
PrecoEmCentavos int `json:"preco_em_centavos"`
```

Etiqueta é só texto pendurado no campo. O `encoding/json` lê aquilo em tempo
de execução e decide o nome. Outros pacotes usam a mesma mecânica com chaves
diferentes, como `db:` ou `yaml:`, e por isso você vê campos com três
etiquetas em projetos grandes.

O vocabulário que você usa neste módulo:

| Etiqueta | Efeito |
|----------|--------|
| `json:"cliente"` | grava e lê com esse nome |
| `json:"observacao,omitempty"` | some do arquivo quando está no valor zero |
| `json:"-"` | nunca grava e nunca lê |
| sem etiqueta | usa o nome do campo em Go, com maiúscula |

**Campo minúsculo não vai para o JSON, e ninguém avisa.**
A regra de visibilidade dos módulos 01 e 04 volta a cobrar aqui, e é o erro
mais frequente com `encoding/json`. Um campo `nome string` simplesmente não
aparece no arquivo, sem erro, sem aviso. O `encoding/json` mora fora do seu
pacote e enxerga só o que é exportado.

**Por que `io.Writer` e não um nome de arquivo?**
Porque `io.Writer` é uma interface de um método, e um monte de coisa a
satisfaz: arquivo, conexão de rede, resposta HTTP, buffer de memória,
compressor, a saída do terminal. Escrevendo contra a interface, a mesma
função serve para todas.

O ganho aparece no teste deste módulo. `TestIdaEVolta` salva e recarrega os
pedidos usando um `bytes.Buffer` como destino **e** como origem, sem tocar em
disco. É rápido, não deixa sujeira, e não falha porque a máquina de CI está
com o disco cheio.

Essa é a regra geral: **funções de biblioteca recebem `io.Reader` e
`io.Writer`; quem lida com nome de arquivo é uma casca fina por cima**. É por
isso que existem quatro funções neste módulo, e não duas.

**`defer` faz o quê?**
Adia a chamada para quando a função ao redor terminar, aconteça o que
acontecer, inclusive em caso de pânico. Serve para não esquecer de fechar,
destravar ou limpar. Os defers empilham: o último registrado roda primeiro, o
que é o comportamento certo quando você abre coisas em sequência e precisa
fechar na ordem inversa. `TestDemonstracaoOrdemDoDefer` mostra isso rodando.

**Por que o retorno de `SalvarPedidosEmArquivo` tem nome?**
Porque `defer arquivo.Close()`, sozinho, **joga fora o erro de fechar**. Numa
leitura isso é aceitável. Numa escrita não: fechar é quando o buffer do
sistema é descarregado, e um erro ali significa que os dados podem não ter
chegado ao disco. Disco cheio aparece exatamente assim.

O padrão completo é este:

```go
func SalvarPedidosEmArquivo(caminho string, pedidos []Pedido) (err error) {
	arquivo, err := os.Create(caminho)
	if err != nil {
		return fmt.Errorf("criar %q: %w", caminho, err)
	}

	defer func() {
		if erroAoFechar := arquivo.Close(); erroAoFechar != nil && err == nil {
			err = fmt.Errorf("fechar %q: %w", caminho, erroAoFechar)
		}
	}()

	return SalvarPedidos(arquivo, pedidos)
}
```

O retorno nomeado é o que permite ao defer alterar o erro **depois** do
return. E o `err == nil` garante que um erro de escrita, que diz mais, não
seja atropelado por um erro de fechamento.

## Pegadinhas

**Campo ausente e campo com zero são a mesma coisa.**
`{"quantidade": 0}` e `{}` produzem o mesmo `Pedido`, porque o valor zero de
`int` é `0` nos dois casos.
`TestDemonstracaoCampoAusenteEZeroSaoIndistinguiveis` mostra isso passando.

Quando a diferença importar, e às vezes importa muito, as saídas são usar
ponteiro no campo, `*int`, onde `nil` significa ausente, ou guardar o JSON
cru em `json.RawMessage` e olhar depois. Ponteiro resolve, e cobra: todo
acesso ao campo passa a precisar de verificação.

**Campo desconhecido some sem reclamar.**
Por padrão, o decodificador ignora o que não conhece.
`TestDemonstracaoCampoDesconhecidoEhIgnorado` mostra. Quando você quer que um
campo escrito errado seja erro, e num arquivo de configuração você quase
sempre quer, use `decodificador.DisallowUnknownFields()`.

**Não compare `time.Time` com `==`.**
Um `time.Time` carrega fuso e, às vezes, um relógio monotônico. Dois valores
que representam o mesmo instante podem ser diferentes byte a byte. Use
`.Equal()`. É por esse motivo que o auxiliar `exigirPedidosIguais` compara
campo a campo em vez de usar `==` no `Pedido` inteiro.

**`os.ReadFile` devolve um erro embrulhado.**
Ele não devolve o sentinela pelado: vem um `*fs.PathError` com o caminho
dentro. Por isso a checagem é `errors.Is(err, fs.ErrNotExist)`, e não
comparação direta. É o módulo 03 cobrando de novo.

**Nunca escreva em caminho fixo dentro de teste.**
Use `t.TempDir()`. Ele devolve uma pasta nova, exclusiva daquele teste, e
apaga tudo sozinho no fim, mesmo se o teste falhar. Teste que escreve em
`/tmp/pedidos.json` funciona sozinho e quebra quando dois testes rodam em
paralelo.

**Sobrescrever não é a mesma coisa que truncar.**
`os.Create` trunca o arquivo, e é por isso que `TestSalvarSobrescreveOArquivo`
passa. Se você abrir com `os.OpenFile` sem `O_TRUNC` e gravar um conteúdo
menor, sobra o rabo do conteúdo antigo no fim do arquivo, e o JSON fica
inválido de um jeito difícil de enxergar.

## Arquivo dourado

`testdata/pedidos.json` é um **golden file**: a saída esperada mora num
arquivo em vez de dentro do código do teste. Vale quando a saída é grande
demais para ficar legível numa string, e tem a vantagem de você poder abrir e
ler como um humano lê.

O Go trata a pasta `testdata` de forma especial: as ferramentas a ignoram, e
ela nunca vira pacote. É a convenção da linguagem para arquivo de apoio de
teste.

Projetos maiores costumam acrescentar uma bandeira para regravar os arquivos
dourados quando a mudança é intencional:

```go
var atualizar = flag.Bool("update", false, "regrava os arquivos dourados")
```

Aí `go test -update` regenera, e o `git diff` vira a revisão da mudança.

## O exercício

Duas structs para etiquetar e quatro funções em `padaria/padaria.go`:

1. Etiquetas de JSON em `Produto` e `Pedido`, incluindo `omitempty` e `-`.
2. `SalvarPedidos` e `CarregarPedidos`, contra `io.Writer` e `io.Reader`.
3. `SalvarPedidosEmArquivo` e `CarregarPedidosDeArquivo`, a casca fina por
   cima das duas anteriores.

Comece pelas etiquetas e rode `TestSalvarPedidosOmiteOsCamposCertos`: ele
falha com mensagens que dizem exatamente qual etiqueta está faltando.

Uma decisão de produto já vem tomada, e ela está escrita no comentário da
função: arquivo que ainda não existe **não é erro**, é uma padaria que abriu
hoje. Qualquer outro problema de leitura é erro de verdade.

## Desafio, sem teste pronto

Faça a gravação ser atômica. Do jeito que está, se a máquina desligar no meio
de `SalvarPedidosEmArquivo`, o arquivo fica pela metade e a padaria perde
todos os pedidos do dia, inclusive os que já estavam gravados antes.

O padrão é: grave num arquivo temporário na **mesma pasta** do destino, feche,
e só então use `os.Rename` por cima do arquivo final. Renomear dentro do mesmo
sistema de arquivos é atômico, então em qualquer instante o destino contém ou
o conteúdo antigo inteiro ou o novo inteiro, nunca metade.

`os.CreateTemp` te dá o arquivo temporário. Pense também no que fazer com ele
quando a gravação falha no meio: um `defer os.Remove` no caminho temporário
evita encher a pasta de restos.

## Se travar

A solução está em `solucoes/07-json/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
