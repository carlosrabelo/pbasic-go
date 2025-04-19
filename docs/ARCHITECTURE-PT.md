# Arquitetura

## Grafo de Dependência dos Pacotes

```
cmd/pbasic
    ^
    |
  repl
    ^
    |
evaluator
    ^
    |
 parser
    ^
    |
 lexer
    ^
    |
  token
```

Todos os pacotes internos estão em `pbasic/internal/`.

## Fluxo de Dados

```
texto fonte → lexer → tokens → parser → AST → evaluator → saída
                              ↑                      ↓
                          lexer_test             tipos object
                          parser_test            evaluator_test
```

### Token (`pbasic/internal/token`)

Define os tipos de token (identificadores, palavras-chave, operadores, delimitadores) e `token.Position` para localização de erros.

Tipos principais:
- `TokenType` — enum de string (ex.: `LET`, `PRINT`, `PLUS`, `NUMBER`)
- `Token` — um único token com tipo, literal e posição
- `Position` — linha, coluna e nome do arquivo para mensagens de erro

### Lexer (`pbasic/internal/lexer`)

Lê o texto fonte caractere por caractere e produz um fluxo de `Token`. Trata números (inteiros e floats), identificadores, operadores, strings e números de linha.

Lexagem de números: dígitos incluindo `.` → float64. Casal de caracteres lookahead via `ch` e `peek()`.

### Parser (`pbasic/internal/parser`)

Parser Pratt (precedência do operador top-down / precedence climbing). Produz uma AST a partir do fluxo de tokens via `ParseLine()`.

#### Níveis de Precedência

| Nível   | Operadores |
|---------|------------|
| LOWEST  | relacionais (=, <>, <, >, <=, >=) |
| SUM     | +, - |
| PRODUCT | *, / |
| PREFIX  | - unário |

#### Como Funciona

- `parseExpression(precedence)` despacha parsers prefixo e depois parsers infixo
- Parsers prefixo: identificadores → `VarExpr`, números → `NumberExpr`, `(` → agrupado, `-` → prefixo unário, `RND`/`ABS` → chamada de função
- Parser infixo: apenas operadores binários (`+`, `-`, `*`, `/`, relacionais)
- Operadores relacionais consomem apenas até o nível SUM à direita, evitando que `AND`/`OR` (não implementados) liguem incorretamente
- Após um parser prefixo avançar o token, o laço verifica `curPrecedence()` (não `peekPrecedence()`) para decidir se deve continuar

#### Parsing de Comandos

`ParseLine()` trata:
- `REM` / `'` → `RemStmt`
- `LET` → `LetStmt` (variável, `=`, expressão)
- `PRINT` → `PrintStmt` (lista de expressões)
- `IF` → `IfStmt` (condição, `THEN`, bloco de comandos)
- `GOTO` → `GotoStmt` (linha alvo)
- `GOSUB` → `GosubStmt` (linha alvo)
- `RETURN` → `ReturnStmt`
- `INPUT` → `InputStmt` (string de prompt opcional, variável)
- `END` → `EndStmt`
- `LIST` → `ListStmt`
- `RUN` → `RunStmt`
- `NEW` → `NewStmt`
- `EXIT` → `ExitStmt`
- expressão simples → `ExprStmt` (avaliada por efeitos colaterais, valor descartado)

Dois-pontos `:` separa comandos: `parseStatements()` retorna `[]Stmt`.

### AST (`pbasic/internal/ast`)

Todos os nós da AST implementam a interface `Node`.

| Nó | Significado |
|----|-------------|
| `NumberExpr` | literal numérico (float64) |
| `VarExpr` | referência a variável (A–Z) |
| `PrefixExpr` | `-` unário |
| `InfixExpr` | operador binário |
| `FuncExpr` | chamada de função embutida (RND, ABS, FREE) |
| `LetStmt` | atribuição de variável |
| `PrintStmt` | lista de expressões para imprimir |
| `IfStmt` | condição + `BlockStmt` (corpo do THEN) |
| `GotoStmt` | salto para número de linha |
| `GosubStmt` | chamada de sub-rotina |
| `ReturnStmt` | retorno de sub-rotina |
| `InputStmt` | string de prompt + variável alvo |
| `EndStmt` | terminação do programa |
| `RemStmt` | comentário |
| `ListStmt` | listar programa |
| `RunStmt` | executar programa |
| `NewStmt` | limpar programa |
| `ExitStmt` | sair do REPL |
| `ExprStmt` | expressão usada como comando |
| `BlockStmt` | sequência de comandos (para THEN ou modo direto) |

### Evaluator (`pbasic/internal/evaluator`)

Avaliador por percurso em árvore (tree-walking). O ponto de entrada principal é `Eval(node, ctx)`, que usa type-switch no nó da AST:

- Expressões → produzem `object.Object` (sempre `*object.Number` para numéricos)
- Comandos → produzem `object.Object` (valores de retorno usados para fluxo de controle)
- `BlockStmt` → avalia cada comando em sequência

#### Interface ProgramStore

`EvalContext.Program` implementa:

```go
type ProgramStore interface {
    Find(n int) (ast.Stmt, bool)
    Lines() []int
    Len() int
    FreeMem() int
    Clear()
}
```

Isso desacopla o evaluator da implementação concreta de armazenamento do programa no `repl`.

#### Fluxo de Controle

- `GotoStmt` / `GosubStmt`: retornam um `*object.GotoValue` contendo o número da linha alvo
- `ReturnStmt`: retorna um `*object.ReturnValue`
- O `execLoop()` no REPL verifica o objeto retornado e despacha adequadamente
- Chamadas de sub-rotina são rastreadas via uma pilha Go `[]int` em `EvalContext.SubStack`
- A flag `Running` no `EvalContext` controla o laço de execução

#### Tratamento de Sinal

- `EvalContext.SigCh` é um `chan os.Signal` (buffered, 1)
- O REPL verifica `sigCh` de forma não-bloqueante entre execuções de comando e antes de cada `ReadString`
- Ao receber `SIGINT` durante execução: define `Running = false` (programa para, permanece no REPL)
- Ao receber `SIGINT` no prompt do REPL: `ReadString` retorna string vazia (exibe novo prompt)

### Object (`pbasic/internal/object`)

Define valores em tempo de execução:

| Tipo | Significado |
|------|-------------|
| `Number` | float64, encapsula um único campo `Value`. `Inspect()` formata como inteiro (sem sufixo `.`) quando o valor é inteiro e < 1e15 |
| `Error` | erro em tempo de execução com mensagem e `token.Position` |
| `Nil` | placeholder nulo (stub para uso futuro) |
| `GotoValue` | fluxo de controle — número da linha alvo (int) |
| `ReturnValue` | fluxo de controle — sentinela de retorno de sub-rotina |

### REPL (`pbasic/internal/repl`)

Orquestra o loop read-eval-print:

1. Exibe banner `PicoBasic`
2. Loop:
   - Imprime prompt `> `
   - Lê uma linha de entrada
   - Faz parse via `parser.ParseLine()`
   - Se modo direto (sem número de linha): avalia imediatamente
   - Se linha numerada: armazena no `ProgramStore`
   - Se `RUN`: chama `execLoop()`
   - Se `EXIT`: encerra

`ProgramStore` é um slice ordenado de `LineNode` gerenciado pelo `REPL`.

Modo arquivo (`LoadFile` + `RunProgram` / `RunFile`): linhas numeradas são armazenadas, linhas vazias e sem número são ignoradas, e o programa executa automaticamente. Sem banner nem prompt. Erros retornam ao `main` para exit code diferente de zero.

#### execLoop()

1. Avalia cada linha do programa sequencialmente via `Eval()`
2. Verifica o objeto retornado para `GotoValue` / `ReturnValue` / `Error`
3. Verifica `sigCh` de forma não-bloqueante — se sinal recebido, define `Running = false` e interrompe
4. Após o loop, redefine `Running` para `false`
5. Devolve erro em falha de runtime para o modo arquivo sair com status 1

### Ponto de Entrada (`pbasic/cmd/pbasic/main.go`)

Mínimo: sem args → `repl.Run()` (interativo). Um arg → `repl.RunFile(path)` (carrega, executa, sai). Caso contrário imprime o usage e sai com 1.

## Decisões de Design

- **Sem `panic`/`recover`**: Fluxo de controle (exit, goto, return) usa valores de retorno. Uma única flag `ShouldExit` sinaliza shutdown limpo para o REPL.
- **Tipo numérico único (`float64`)**: Compatível com o comportamento clássico do TinyBASIC. A saída renderiza números inteiros sem decimal.
- **Parser Pratt em vez de descida recursiva**: Mais fácil de manter a precedência de operadores; mesma abordagem usada pelo Go.
- **Avaliador tree-walking**: Simples, correto, adequado para um interpretador educacional. Não otimizado para velocidade.
- **`EvalContext` como contexto single-threaded**: Contém variáveis, armazenamento do programa, índice de comandos, pilha de sub-rotinas, canal de sinal e flag de execução. Passado por todas as avaliações.
- **Interface ProgramStore**: Permite teste sem armazenamento real; mantém o evaluator independente dos internos do REPL.
