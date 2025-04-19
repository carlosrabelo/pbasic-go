# PicoBasic Language Reference

## Invoking a program file

This is a CLI feature, not a language command. There is no `LOAD` statement.

```
pbasic                  # interactive REPL
pbasic program.bas      # load numbered lines, RUN, then exit
```

Unnumbered lines in a file (for example a trailing `RUN`) are ignored. `INPUT` still reads from standard input.

## Syntax Overview

Lines in PicoBasic follow this structure:

```
<line-number> <statement>[: <statement> ...]
```

Line numbers are 1–63999. Multiple statements can be chained on one line with `:`.

Lines without a number (direct mode) execute immediately.

Comments use `REM` (or `'` as a synonym):

```
10 REM This is a comment
```

---

## Statements (13)

### LET

Assigns a value to a variable.

```
LET <var> = <expr>
```

Variable names are single letters `A`–`Z`. Variables are initialized to 0.

```
10 LET X = 42
20 LET A = B + 3
```

### PRINT

Outputs values to the terminal.

```
PRINT <expr> [, <expr> ...]
```

Multiple expressions are separated by `,` (two spaces between each). A bare `PRINT` outputs a blank line.

```
10 PRINT "HELLO"
20 PRINT X
30 PRINT "A = "; A
40 PRINT
```

### IF / THEN

Conditional execution.

```
IF <rel-expr> THEN <statement>[: <statement> ...]
```

Runs the statement(s) after `THEN` only if the relational expression is true (non-zero).

```
10 IF X > 5 THEN PRINT "BIG"
20 IF A = B THEN PRINT "EQ" : GOTO 100
```

### GOTO

Unconditional jump to a line number.

```
GOTO <line-number>
```

```
10 GOTO 50
```

### GOSUB / RETURN

Subroutine call and return.

```
GOSUB <line-number>
...
RETURN
```

Subroutines may be nested.

```
10 GOSUB 100
20 END
100 PRINT "SUBROUTINE"
110 RETURN
```

### INPUT

Reads a number from the user.

```
INPUT [<prompt> ;] <var>
```

Displays an optional prompt, then waits for input. If the user enters an empty line, input is re-prompted. Ctrl+C during input cancels the running program.

```
10 INPUT A
20 INPUT "GUESS: "; G
```

### END

Stops program execution.

```
END
```

### REM

Comment (ignored). `'` (apostrophe) is a synonym for REM.

```
10 REM THIS IS A COMMENT
20 ' ALSO A COMMENT
```

### LIST

Lists the stored program lines in order.

```
LIST
```

### RUN

Starts execution at the lowest line number.

```
RUN
```

### NEW

Clears the current program and variables.

```
NEW
```

### EXIT

Exits the REPL.

```
EXIT
```

---

## Expressions

### Arithmetic

Standard infix operators with precedence:

| Operator | Associativity | Precedence |
|----------|---------------|------------|
| `*` `/`  | left          | high       |
| `+` `-`  | left          | low        |

Unary `-` is supported with the same precedence as `*`/`/`.

```
10 LET X = 3 + 4 * 2      -> 11
20 LET Y = (3 + 4) * 2    -> 14
30 LET Z = -5 + 3          -> -2
```

### Relational

Relational operators return 1 (true) or 0 (false):

`=`, `<>`, `<`, `>`, `<=`, `>=`

They have the lowest precedence.

```
10 IF A <> 0 THEN PRINT "NONZERO"
20 LET X = 5 > 3            -> X = 1
```

### Built-in Functions

| Function | Description |
|----------|-------------|
| `FREE`   | Returns available memory (remaining lines) |
| `RND(n)` | Returns random integer 1 ≤ x ≤ n |
| `ABS(n)` | Returns absolute value of n |

```
10 PRINT FREE
20 LET R = RND(10)
30 PRINT ABS(-7.5)
```

### Numbers

Numbers are stored as `float64`. Output is rendered as integer when the value is a whole number, otherwise as decimal.

```
PRINT 42       -> 42
PRINT 10 / 3   -> 3.333333
PRINT 2.5 * 4  -> 10
PRINT 2 / 4    -> 0.5
```

Line numbers 0–63999, variable values are full float64 range.

### Strings

String literals use double quotes:

```
PRINT "HELLO, WORLD!"
```

String comparison and string variables are not supported — strings are output-only.

### Variables

26 single-letter variables: `A` through `Z`. All numeric (float64), initialized to 0.

---

## Colons (Multiple Statements)

Use `:` to put multiple statements on one line:

```
10 LET A = 5 : LET B = 10 : PRINT A + B
20 IF X = 0 THEN PRINT "ZERO" : GOTO 100
```

---

## Error Handling

- Division by zero returns 0 (no error).
- GOTO/GOSUB to non-existent line prints a message and stops.
- Subroutine stack overflow (>100 levels) prints error and stops.
- Syntax errors print the error with position and line number.

---

## Memory

The `ProgramStore` holds up to 100 lines by default (`FreeMem` reports how many slots remain).
