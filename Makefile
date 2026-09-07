.PHONY: ajuda testar verificar formatar

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
