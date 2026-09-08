# Trilha Go

Um curso de Go em português, para quem já programa em outra linguagem.

Você não vai ler sobre Go. Você vai abrir um repositório com testes falhando e
fazer eles passarem, módulo por módulo, construindo a mesma aplicação do começo
ao fim: a **Padaria do Seu Zé**, que começa como um programa de terminal, vira
uma API HTTP e termina com uma fila de pedidos processada em paralelo.

## Comece por aqui

Você precisa de Go 1.24 ou mais novo. Confira com `go version`.

Os módulos 00 a 10 usam só a biblioteca padrão. A partir do 11 há uma
dependência, um SQLite escrito em Go puro, que o `go test` baixa sozinho na
primeira vez. Não é preciso instalar banco de dados nenhum.

```bash
git clone https://github.com/RootControl/trilha-go
cd trilha-go
make testar
```

Os testes vão falhar. É esse o ponto de partida. Abra
[`modulos/00-ambiente/README.md`](modulos/00-ambiente/README.md) e siga daí.

## Como cada módulo funciona

Todo módulo tem a mesma forma:

- Um `README.md` com o conceito, uma seção **por que Go faz assim** que compara
  com JavaScript, Python e PHP, e uma seção **pegadinhas**.
- Um pacote com funções por implementar, marcadas com `TODO`.
- Um arquivo de teste **já escrito e falhando**. Ele é o enunciado.
- Um desafio final sem teste pronto, para quando a mão já estiver solta.

A tabela de casos no topo de cada teste é a especificação. Leia ela antes de
escrever qualquer código.

## A trilha

| # | Módulo | A padaria aprende a |
|---|--------|---------------------|
| 00 | [Ambiente e o primeiro teste](modulos/00-ambiente/) | receber cliente e mostrar preço |
| 01 | [Tipos e zero values](modulos/01-tipos/) | descrever produto e pedido |
| 02 | [Slices e maps](modulos/02-colecoes/) | ter cardápio e estoque |
| 03 | [Erros como valores](modulos/03-erros/) | dizer que o pão acabou |
| 04 | [Métodos e ponteiros](modulos/04-metodos/) | dar baixa no estoque |
| 05 | [Interfaces implícitas](modulos/05-interfaces/) | trocar onde guarda os dados |
| 06 | [Testes de verdade](modulos/06-testes/) | ser testada sem framework |
| 07 | [JSON e arquivos](modulos/07-json/) | não perder os pedidos ao desligar |
| 08 | [Um programa de terminal](modulos/08-cli/) | ser usada de verdade |
| 09 | [Genéricos](modulos/09-genericos/) | filtrar cardápio sem repetir código |
| 10 | [Servidor HTTP](modulos/10-http/) | atender pela internet |
| 11 | [Banco de dados](modulos/11-banco/) | guardar tudo em SQL |
| 12 | [Concorrência](modulos/12-concorrencia/) | assar vários pães ao mesmo tempo |
| 13 | Fila de pedidos | não derrubar tudo no horário de pico |
| 14 | Produção | rodar em contêiner com log e deploy |

Os módulos com link estão prontos. Os outros vêm nessa ordem.

## Comandos

```bash
make testar      # roda os testes dos exercícios: é aqui que você trabalha
go test -race ./modulos/...   # o mesmo, com o detector de corrida
make verificar   # confere que as soluções passam nos testes dos módulos
make formatar    # formata tudo com gofmt
```

`make verificar` é o que roda no CI. Ele copia os testes de cada módulo por
cima da solução correspondente e roda. Assim os testes têm uma fonte da verdade
só, e uma solução que sai do lugar quebra o build em vez de enganar alguém.

## Contribuindo

Dúvida sobre um enunciado é um bug do material, não do aluno. Abra uma issue.
Correções, módulos novos e traduções de exemplo são bem-vindos por pull request,
desde que `make verificar` passe.

## Licença

Código sob MIT, no arquivo [`LICENSE`](LICENSE). Texto dos módulos sob
Creative Commons BY 4.0, detalhes em
[`LICENSE-CONTEUDO.md`](LICENSE-CONTEUDO.md).
