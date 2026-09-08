#!/usr/bin/env bash
#
# verificar.sh garante que a Trilha Go continua honesta.
#
# Os testes de cada módulo são a única fonte da verdade. Este script copia
# cada arquivo *_test.go de modulos/ para a pasta espelhada em solucoes/,
# roda os testes contra a solução e apaga as cópias no fim.
#
# Se uma solução deixar de satisfazer o enunciado do próprio exercício, o
# build quebra aqui em vez de confundir um aluno daqui a seis meses.

set -euo pipefail

cd "$(dirname "$0")/.."

lista_de_copias="$(mktemp)"

limpar() {
  while IFS= read -r arquivo; do
    [ -n "$arquivo" ] && rm -f "$arquivo"
  done < "$lista_de_copias"
  rm -f "$lista_de_copias"
  # -empty garante que só pastas vazias somem, nunca uma com conteúdo.
  find solucoes -type d -name testdata -empty -delete 2>/dev/null || true
}
trap limpar EXIT

echo "==> compilando os exercícios"
# Os stubs precisam compilar. A primeira coisa que o aluno vê tem que ser um
# teste falhando, nunca um erro de compilação do material.
go build ./modulos/...
go vet ./modulos/...

echo "==> verificando a formatação"
nao_formatados="$(gofmt -l .)"
if [ -n "$nao_formatados" ]; then
  echo "os arquivos abaixo não estão formatados. rode: make formatar" >&2
  echo "$nao_formatados" >&2
  exit 1
fi

echo "==> conferindo que os exercícios falham sem entrar em pânico"
# Com os stubs por implementar, os testes DEVEM falhar. O que eles não podem
# fazer é entrar em pânico: um índice fora da faixa ou um nil desreferenciado
# aborta o binário de teste inteiro e esconde as mensagens que deveriam
# orientar o aluno. Cada teste precisa parar com t.Fatal antes de tocar em
# resultado que o stub não produziu.
saida_dos_exercicios="$(go test ./modulos/... 2>&1 || true)"
if printf '%s' "$saida_dos_exercicios" | grep -q '^panic:'; then
  echo "os testes dos módulos entraram em pânico com os stubs:" >&2
  printf '%s\n' "$saida_dos_exercicios" | grep -A3 '^panic:' >&2
  exit 1
fi

echo "==> copiando os testes dos módulos para as soluções"
while IFS= read -r teste; do
  destino="solucoes/${teste#modulos/}"
  pasta_destino="$(dirname "$destino")"

  if [ ! -d "$pasta_destino" ]; then
    echo "não existe solução para $teste (esperava a pasta $pasta_destino)" >&2
    exit 1
  fi

  cp "$teste" "$destino"
  echo "$destino" >> "$lista_de_copias"
  echo "    $teste -> $destino"
done < <(find modulos -name '*_test.go' | sort)

echo "==> copiando os testdata dos módulos para as soluções"
while IFS= read -r fixture; do
  destino="solucoes/${fixture#modulos/}"
  mkdir -p "$(dirname "$destino")"
  cp "$fixture" "$destino"
  echo "$destino" >> "$lista_de_copias"
  echo "    $fixture -> $destino"
done < <(find modulos -type d -name testdata -exec find {} -type f -print \; | sort)

echo "==> rodando os testes contra as soluções"
go test ./solucoes/...

echo
echo "tudo certo: cada solução passa nos testes do seu próprio módulo."
