.PHONY: ajuda testar verificar formatar binario imagem

ajuda:
	@echo "make testar      roda os testes dos exercícios (é aqui que você trabalha)"
	@echo "make verificar   confere que as soluções passam nos testes dos módulos"
	@echo "make formatar    formata todo o código com gofmt"

testar:
	go test ./modulos/...

verificar:
	./scripts/verificar.sh

formatar:
	gofmt -w .

binario:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/padaria ./modulos/14-producao/cmd/servidor

imagem:
	docker build -f modulos/14-producao/Dockerfile -t padaria:local .
