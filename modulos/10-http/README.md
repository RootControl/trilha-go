# Módulo 10 — Servidor HTTP

A padaria passa a atender pela rede. Sem framework: tudo que aparece neste
módulo vem da biblioteca padrão.

```bash
go test ./modulos/10-http/...
go run ./modulos/10-http/cmd/servidor
```

Com o servidor no ar, em outro terminal:

```bash
curl -s localhost:8080/produtos
curl -s "localhost:8080/produtos/P%C3%A3o%20de%20queijo"
curl -s -X POST localhost:8080/vendas -d '{"cliente":"Maria","produto":"Broa","quantidade":2}'
curl -s -i localhost:8080/produtos/Croissant
```

## Por que Go faz assim

**Cadê o Express, o Flask, o Laravel, o Spring?**
`net/http`, da biblioteca padrão, e ele é usado em produção do jeito que vem.
Até a versão 1.21 havia um argumento honesto a favor de uma biblioteca de
roteamento, porque o roteador embutido não sabia distinguir método nem extrair
pedaço de caminho. Isso mudou no Go 1.22:

```go
mux.HandleFunc("GET /produtos/{nome}", handler)
```

Método, caminho e curinga no mesmo texto, e `r.PathValue("nome")` para ler o
curinga. O mux ainda devolve **405** sozinho quando o caminho casa com algum
padrão mas o método não. Para a maioria das APIs, isso já é o suficiente.

**`http.Handler` é uma interface de um método.**

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

O módulo 05 outra vez, e é essa escolha que sustenta todo o ecossistema HTTP
de Go. Um `*ServeMux` é um Handler. O seu `*Servidor` é um Handler. Um
middleware devolve um Handler. Todos se encaixam porque todos são a mesma
coisa.

**`http.HandlerFunc` é um tipo função que tem um método.**

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```

Leia devagar, porque é um truque bonito: um **tipo função** com um método
pendurado nele. Converter a sua função para esse tipo é tudo que falta para
ela virar um Handler. Não há registro, não há herança, não há decorador.

**Middleware não é conceito de framework.**

```go
func ComRegistro(proximo http.Handler, registrador Registrador) http.Handler
```

A forma `func(http.Handler) http.Handler` é apenas o que sai naturalmente
quando as duas pontas são a mesma interface. Middleware empilha porque a saída
de um serve de entrada do outro, e `TestDemonstracaoServidorEhHandler` empilha
dois de propósito para você ver.

**Nada de estado global.**
O `*Servidor` guarda o cardápio, o estoque e o relógio. Cada teste monta o seu
próprio, com estoque próprio, e por isso os testes deste módulo não interferem
uns nos outros. Handler que lê variável global é handler que você não consegue
testar duas vezes com dados diferentes.

**Teste sem abrir porta.**
`httptest.NewRequest` monta a requisição, `httptest.NewRecorder` recebe a
resposta, e você chama `ServeHTTP` direto. Nenhum teste deste módulo escuta
numa porta. É rápido, não conflita quando dois testes rodam ao mesmo tempo, e
não depende de a máquina ter a porta livre. Quando você precisar de um servidor
de verdade, por exemplo para testar um cliente HTTP, aí existe
`httptest.NewServer`.

**A tradução de erro mora na borda.**
Procure `http` no código de domínio da padaria: não existe. O pacote não sabe
o que é status code, e não deve saber. Quem traduz é `StatusParaErro`, aqui na
camada de rede, usando `errors.Is` porque os erros chegam embrulhados em
camadas de contexto, como no módulo 03.

Sobre a escolha de **409** para estoque insuficiente: o pedido está bem
formado, então não é 400. O que falha é o estado atual do servidor, e depois
de uma reposição o mesmo pedido funciona. Isso é conflito, não erro de
sintaxe.

## Pegadinhas

**Cabeçalho depois do `WriteHeader` não faz nada.**
A pegadinha número um do `net/http`. Assim que `WriteHeader` roda, os
cabeçalhos foram enviados, e qualquer `Header().Set` depois disso é descartado
em silêncio. A ordem é sempre: cabeçalhos, depois status, depois corpo.
`TestDemonstracaoCabecalhoDepoisDoWriteHeaderNaoFazNada` mostra rodando.

**`gravador.Header()` mente; `gravador.Result().Header` não.**
Esta pega quem escreve o teste, e é sutil. O `ResponseRecorder` guarda um mapa
vivo de cabeçalhos que continua aceitando escrita depois do `WriteHeader`.
Quem congela a foto do que realmente teria ido para o cliente é `Result()`. Um
teste que confere `gravador.Header()` passa mesmo com a ordem invertida, e
portanto não testa nada. Os testes deste módulo usam `Result().Header`.

**Escrever no corpo já chama `WriteHeader(200)`.**
Um `fmt.Fprint(w, ...)` sem status explícito envia 200 sozinho. Chamar
`WriteHeader` depois disso não muda nada e ainda registra um aviso no log do
servidor:

```
http: superfluous response.WriteHeader call from main.main.func1 (main.go:16)
```

Se você ver essa linha, tem um caminho no seu handler que responde duas vezes.

**Caminho com espaço e acento precisa ser escapado.**
`/produtos/Pão de queijo` não é uma URL válida. Use `url.PathEscape`, e o
roteador desescapa antes de entregar em `PathValue`. Na hora de montar o
teste, um espaço cru derruba o `httptest.NewRequest` com um pânico sobre
versão de HTTP malformada, porque ele monta a linha de requisição por texto.

**`http.ListenAndServe` não tem prazo nenhum.**
A função de uma linha que aparece em todo tutorial sobe um servidor sem
`ReadTimeout`, sem `WriteTimeout` e sem `IdleTimeout`. Uma conexão lenta fica
pendurada para sempre, e um punhado delas segura o processo. Em produção,
sempre um `http.Server` com prazos, como em `cmd/servidor/main.go`.

**Não devolva a mensagem de erro interna ao cliente.**
Neste módulo o corpo do erro vem da mensagem crua porque o material é
didático. Num serviço de verdade, um erro 500 deve registrar tudo no log e
responder ao cliente algo genérico. Mensagem de erro é uma das fontes mais
comuns de vazamento de detalhe interno.

**`r.Body` não tem limite.**
Um cliente pode mandar um corpo de vários gigabytes e o seu decodificador vai
tentar ler. Em serviço exposto, embrulhe com `http.MaxBytesReader` antes de
decodificar.

## O exercício

Sete implementações em `padaria/padaria.go`:

1. `NovoServidor`, que monta o mux e registra as quatro rotas.
2. `ServeHTTP`, delegando ao mux, que é o que faz o servidor ser um Handler.
3. `ResponderJSON`, onde a ordem das três linhas é o exercício.
4. `ResponderErro` e `StatusParaErro`, a tradução de domínio para HTTP.
5. `listarProdutos`, `buscarProduto` e `criarVenda`.
6. `ComRegistro`, o middleware, usando a `respostaComStatus` que já vem
   pronta no arquivo.

Comece por `StatusParaErro`. Ela não depende de nada, é um `switch` com
`errors.Is`, e `TestStatusParaErro` testa ela isolada, inclusive com um
sentinela embrulhado duas vezes.

## Desafio, sem teste pronto

**Devolva 400 para campo desconhecido.** Hoje um corpo com
`{"produtoo":"Broa"}` é aceito e vira uma venda de produto vazio, porque o
decodificador ignora o que não conhece, como você viu no módulo 07. Use
`DisallowUnknownFields` e decida o que responder.

**Limite o tamanho do corpo** com `http.MaxBytesReader` e responda 413 quando
estourar.

**Escreva um middleware de tempo** que meça a duração de cada requisição e a
inclua na linha de log. Depois empilhe os dois e repare que a ordem em que
você empilha muda o que cada um enxerga.

E uma pergunta para pensar antes do módulo 12: o estoque deste servidor é um
`map` compartilhado entre todas as requisições, e o `net/http` atende cada uma
numa goroutine própria. Rode os testes com `go test -race`. O que acontece?

## Se travar

A solução está em `solucoes/10-http/`. Dúvida sobre o enunciado é bug do
material: abra uma issue.
