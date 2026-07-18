# PicoBasic

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev/)

Interpretador BASIC minimalista escrito em Go, portado da versão original em assembly MIPS. Executa como um REPL de linha de comando em qualquer plataforma com suporte a Go.

## Destaques

- Dialeto PicoBasic com 13 comandos: LET, PRINT, IF/THEN, GOTO, GOSUB/RETURN, INPUT, LIST, RUN, NEW, END, REM, EXIT
- Avaliador de expressões com análise de descida recursiva (+, -, *, /, parênteses, menos unário)
- 26 variáveis de letra única (A–Z) com aritmética inteira
- Funções embutidas: FREE (memória disponível), RND (aleatório), ABS (valor absoluto)
- Armazenamento de linhas tokenizadas com suporte a LIST para revisão do programa
- REPL interativo com prompt `>` e execução direta de comandos
- Roda em qualquer terminal — sem necessidade de simulador ou emulador

## Pré-requisitos

- **Go 1.23+** — necessário para compilar a partir do código-fonte; [download](https://go.dev/dl/)

## Instalação

### Compilar a partir do código-fonte

```bash
git clone https://github.com/carlosrabelo/pbasic-go.git
cd pbasic-go
make build
```

Instale como usuário em `~/.local/bin`, ou como root em `/usr/local/bin`:

```bash
make install
```

## Uso

Inicie o REPL:

```bash
make run
```

Ou diretamente:

```bash
./bin/pbasic
```

### Exemplo de sessão

```
PicoBasic
> 10 LET A = 42
> 20 PRINT A
> 30 PRINT A * 2 + 10
> RUN
42
94
> PRINT FREE
902
> LIST
10 LET A = 42
20 PRINT A
30 PRINT A * 2 + 10
> NEW
OK
> EXIT
```

## Estrutura do Projeto

```
pbasic/cmd/pbasic/       # Ponto de entrada Go
pbasic/internal/         # Núcleo do interpretador (token, lexer, ast, parser, evaluator, object, repl)
demos/                   # Programas BASIC de demonstração
bin/                     # Binários compilados (ignorados no git)
.make/                   # Scripts de build, teste e instalação
```

## Desenvolvimento

```bash
make build      # Compila o binário para bin/pbasic
make test       # Executa todos os testes
make run        # Compila e executa o REPL
make lint       # Executa o linter
make fmt        # Formata o código
make install    # Instala o binário (~/.local/bin ou /usr/local/bin)
make uninstall  # Remove o binário instalado
make clean      # Remove artefatos de build
```

## Licença

Este projeto está licenciado sob a Licença MIT — veja [LICENSE](LICENSE) para detalhes.
