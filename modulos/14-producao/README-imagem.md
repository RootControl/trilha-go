# Construindo e rodando a imagem

Da raiz do repositório:

```bash
docker build -f modulos/14-producao/Dockerfile -t padaria:local .
docker run --rm -p 8080:8080 padaria:local
```

Com configuração diferente:

```bash
docker run --rm -p 9000:9000 \
  -e PADARIA_ENDERECO=:9000 \
  -e PADARIA_LOG_NIVEL=debug \
  padaria:local
```

Para ver o desligamento com graça funcionando dentro do contêiner, rode em um
terminal e, em outro:

```bash
docker stop $(docker ps -q --filter ancestor=padaria:local)
```

O `docker stop` manda `SIGTERM` e espera dez segundos antes de matar. O
servidor usa esse tempo para terminar as requisições em andamento.

Para conferir o tamanho:

```bash
docker images padaria:local
```
