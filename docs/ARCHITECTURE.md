# Architecture

## Package Dependency Graph

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

All internal packages are under `pbasic/internal/`.

## Data Flow

```
source text → lexer → tokens → parser → AST → evaluator → output
                              ↑                      ↓
                          lexer_test             object types
                          parser_test            evaluator_test
```

### Token (`pbasic/internal/token`)

Defines token types (identifiers, keywords, operators, delimiters) and `token.Position` for error locations.

Key types:
- `TokenType` — string enum (e.g., `LET`, `PRINT`, `PLUS`, `NUMBER`)
- `Token` — a single token with type, literal, and position
- `Position` — line, column, and filename for error messages

### Lexer (`pbasic/internal/lexer`)

Reads source text character-by-character and produces a `Token` stream. Handles numbers (integers and floats), identifiers, operators, strings, and line numbers.

Number lexing: digits including `.` → float64. Single-character lookahead via `ch` and `peek()`.

### Parser (`pbasic/internal/parser`)

Pratt parser (top-down operator precedence / precedence climbing). Produces an AST from the token stream via `ParseLine()`.

#### Precedence Levels

| Level | Operators |
|-------|-----------|
| LOWEST | relational (=, <>, <, >, <=, >=) |
| SUM    | +, - |
| PRODUCT| *, / |
| PREFIX | unary - |

#### How It Works

- `parseExpression(precedence)` dispatches prefix parsers then infix parsers
- Prefix parsers: identifiers → `VarExpr`, numbers → `NumberExpr`, `(` → grouped, `-` → unary prefix, `RND`/`ABS` → function call
- Infix parser: only binary operators (`+`, `-`, `*`, `/`, relational)
- Relational operators consume only up to SUM-level precedence on the right, preventing `AND`/`OR` (not implemented) from binding incorrectly
- After a prefix parser advances the token, the loop checks `curPrecedence()` (not `peekPrecedence()`) to decide whether to continue

#### Statement Parsing

`ParseLine()` handles:
- `REM` / `'` → `RemStmt`
- `LET` → `LetStmt` (variable, `=`, expression)
- `PRINT` → `PrintStmt` (list of expressions)
- `IF` → `IfStmt` (condition, `THEN`, block of statements)
- `GOTO` → `GotoStmt` (target line number)
- `GOSUB` → `GosubStmt` (target line number)
- `RETURN` → `ReturnStmt`
- `INPUT` → `InputStmt` (optional string prompt, variable)
- `END` → `EndStmt`
- `LIST` → `ListStmt`
- `RUN` → `RunStmt`
- `NEW` → `NewStmt`
- `EXIT` → `ExitStmt`
- bare expression → `ExprStmt` (evaluated for side effects, value discarded)

Colon `:` separates statements: `parseStatements()` returns `[]Stmt`.

### AST (`pbasic/internal/ast`)

All AST nodes implement the `Node` interface.

| Node | Meaning |
|------|---------|
| `NumberExpr` | numeric literal (float64) |
| `VarExpr` | variable reference (A–Z) |
| `PrefixExpr` | unary `-` |
| `InfixExpr` | binary operator |
| `FuncExpr` | built-in function call (RND, ABS, FREE) |
| `LetStmt` | variable assignment |
| `PrintStmt` | list of expressions to print |
| `IfStmt` | condition + `BlockStmt` (THEN body) |
| `GotoStmt` | jump to line number |
| `GosubStmt` | subroutine call |
| `ReturnStmt` | subroutine return |
| `InputStmt` | prompt string + target variable |
| `EndStmt` | program termination |
| `RemStmt` | comment |
| `ListStmt` | list program |
| `RunStmt` | run program |
| `NewStmt` | clear program |
| `ExitStmt` | exit REPL |
| `ExprStmt` | expression used as statement |
| `BlockStmt` | sequence of statements (for THEN or direct mode) |

### Evaluator (`pbasic/internal/evaluator`)

Tree-walking evaluator. The main entry point is `Eval(node, ctx)`, which type-switches on the AST node:

- Expressions → produce `object.Object` (always `*object.Number` for numerics)
- Statements → produce `object.Object` (return values used for control flow)
- `BlockStmt` → evaluates each statement in sequence (previously had a `Running` flag guard — removed because it prevented direct-mode multi-statement execution)

#### ProgramStore Interface

`EvalContext.Program` implements:

```go
type ProgramStore interface {
    Find(n int) (ast.Stmt, bool)
    Lines() []int
    Len() int
    FreeMem() int
    Clear()
}
```

This decouples the evaluator from the concrete program storage implementation in `repl`.

#### Control Flow

- `GotoStmt` / `GosubStmt`: return a `*object.GotoValue` containing the target line number
- `ReturnStmt`: return a `*object.ReturnValue`
- The `execLoop()` in the REPL checks the returned object and dispatches accordingly
- Subroutine calls are tracked via a Go `[]int` stack in `EvalContext.SubStack`
- `Running` flag on `EvalContext` gates the execution loop

#### Signal Handling

- `EvalContext.SigCh` is a `chan os.Signal` (buffered, 1)
- The REPL checks `sigCh` non-blockingly between statement executions and before each `ReadString`
- On `SIGINT` during execution: sets `Running = false` (program stops, stays in REPL)
- On `SIGINT` in REPL prompt: `ReadString` returns empty string (prints new prompt)

### Object (`pbasic/internal/object`)

Defines runtime values:

| Type | Meaning |
|------|---------|
| `Number` | float64, wraps a single `Value` field. `Inspect()` formats as integer (no `.` suffix) when value is whole and < 1e15 |
| `Error` | runtime error with message and `token.Position` |
| `Nil` | null placeholder (stub for future use) |
| `GotoValue` | control flow — target line number (int) |
| `ReturnValue` | control flow — subroutine return sentinel |

### REPL (`pbasic/internal/repl`)

Orchestrates the read-eval-print loop:

1. Display `PicoBasic` banner
2. Loop:
   - Print `> ` prompt
   - Read a line of input
   - Parse via `parser.ParseLine()`
   - If direct mode (no line number): evaluate immediately
   - If numbered line: store in `ProgramStore`
   - If `RUN`: call `execLoop()`
   - If `EXIT`: break

`ProgramStore` is an in-memory map of `map[int]ast.Stmt` managed by `REPL`.

#### execLoop()

1. Evaluates each program line sequentially via `Eval()`
2. Checks returned object for `GotoValue` / `ReturnValue` / `Error`
3. Checks `sigCh` non-blockingly — if signal received, sets `Running = false` and breaks
4. After loop, resets `Running` to `false`

### Entry Point (`pbasic/cmd/pbasic/main.go`)

Minimal: calls `repl.Start()` with `os.Stdin`/`os.Stdout`/`os.Stderr`.

## Design Decisions

- **No `panic`/`recover`**: Control flow (exit, goto, return) uses return values. A single `ShouldExit` flag signals clean shutdown to the REPL.
- **Single numeric type (`float64`)**: Matches classic TinyBASIC behavior. Output renders whole numbers without decimal.
- **Pratt parser instead of recursive descent**: Easier to maintain operator precedence; same approach used by Go itself.
- **Tree-walking evaluator**: Simple, correct, adequate for educational interpreter. Not optimized for speed.
- **`EvalContext` as single threaded context**: Holds variables, program store, statement index, subroutine stack, signal channel, and running flag. Passed through all evaluations.
- **ProgramStore interface**: Allows testing without real storage; keeps evaluator independent of REPL internals.
