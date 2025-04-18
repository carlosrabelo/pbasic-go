# PicoBasic

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev/)

Minimal BASIC interpreter written in Go, ported from the original MIPS assembly version. Runs as a command-line REPL on any platform with Go support.

## Highlights

- PicoBasic dialect with 13 commands: LET, PRINT, IF/THEN, GOTO, GOSUB/RETURN, INPUT, LIST, RUN, NEW, END, REM, EXIT
- Expression evaluator with recursive descent parsing (+, -, *, /, parentheses, unary minus)
- 26 single-letter variables (A–Z) with integer arithmetic
- Built-in functions: FREE (available memory), RND (random), ABS (absolute value)
- Tokenized line storage with LIST support for program review
- Interactive REPL with `>` prompt and direct command execution
- Runs in any terminal — no simulator or emulator required

## Prerequisites

- **Go 1.23+** — required to build from source; [download](https://go.dev/dl/)

## Installation

### Build from Source

```bash
git clone https://github.com/carlosrabelo/pbasic-go.git
cd pbasic-go
make build
```

Install to `~/.local/bin` (default), or system-wide to `/usr/local/bin` (sudo only for the copy):

```bash
make install
make install-system
make uninstall
make uninstall-system
```

## Usage

Start the REPL:

```bash
make run
```

Or directly:

```bash
./bin/pbasic
```

### Example session

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

## Project Layout

```
pbasic/cmd/pbasic/       # Go entry point
pbasic/internal/         # Interpreter core (token, lexer, ast, parser, evaluator, object, repl)
demos/                   # BASIC demonstration programs
bin/                     # Compiled binaries (git-ignored)
.make/                   # Build, test, and install scripts
```

## Development

```bash
make build             # Compile binary to bin/pbasic
make test              # Run all tests
make quality           # Format, vet, and lint
make run               # Build and run the REPL
make install           # Install binary to ~/.local/bin
make install-system    # Install binary to /usr/local/bin
make uninstall         # Remove from ~/.local/bin
make uninstall-system  # Remove from /usr/local/bin
```

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.
