# Referência da Linguagem PicoBasic

## Visão Geral da Sintaxe

Linhas no PicoBasic seguem esta estrutura:

```
<numero-da-linha> <comando>[: <comando> ...]
```

Números de linha vão de 1 a 63999. Múltiplos comandos podem ser encadeados na mesma linha com `:`.

Linhas sem número (modo direto) executam imediatamente.

Comentários usam `REM` (ou `'` como sinônimo):

```
10 REM Isto é um comentário
```

---

## Comandos (13)

### LET

Atribui um valor a uma variável.

```
LET <var> = <expr>
```

Nomes de variáveis são letras simples `A`–`Z`. Variáveis são inicializadas com 0.

```
10 LET X = 42
20 LET A = B + 3
```

### PRINT

Exibe valores no terminal.

```
PRINT <expr> [, <expr> ...]
```

Múltiplas expressões são separadas por `,` (dois espaços entre cada). Um `PRINT` sozinho exibe uma linha em branco.

```
10 PRINT "HELLO"
20 PRINT X
30 PRINT "A = "; A
40 PRINT
```

### IF / THEN

Execução condicional.

```
IF <expr-rel> THEN <comando>[: <comando> ...]
```

Executa o(s) comando(s) após `THEN` apenas se a expressão relacional for verdadeira (diferente de zero).

```
10 IF X > 5 THEN PRINT "BIG"
20 IF A = B THEN PRINT "EQ" : GOTO 100
```

### GOTO

Salto incondicional para um número de linha.

```
GOTO <numero-da-linha>
```

```
10 GOTO 50
```

### GOSUB / RETURN

Chamada de sub-rotina e retorno.

```
GOSUB <numero-da-linha>
...
RETURN
```

Sub-rotinas podem ser aninhadas.

```
10 GOSUB 100
20 END
100 PRINT "SUBROUTINE"
110 RETURN
```

### INPUT

Lê um número do usuário.

```
INPUT [<prompt> ;] <var>
```

Exibe um prompt opcional e aguarda entrada. Se o usuário digitar uma linha vazia, o input é solicitado novamente. Ctrl+C durante o input cancela o programa em execução.

```
10 INPUT A
20 INPUT "GUESS: "; G
```

### END

Interrompe a execução do programa.

```
END
```

### REM

Comentário (ignorado). `'` (apóstrofo) é sinônimo de REM.

```
10 REM ISTO É UM COMENTÁRIO
20 ' TAMBÉM UM COMENTÁRIO
```

### LIST

Lista as linhas do programa armazenado em ordem.

```
LIST
```

### RUN

Inicia a execução a partir do menor número de linha.

```
RUN
```

### NEW

Limpa o programa atual e as variáveis.

```
NEW
```

### EXIT

Sai do REPL.

```
EXIT
```

---

## Expressões

### Aritmética

Operadores infixos padrão com precedência:

| Operador | Associatividade | Precedência |
|----------|-----------------|-------------|
| `*` `/`  | esquerda        | alta        |
| `+` `-`  | esquerda        | baixa       |

`-` unário é suportado com a mesma precedência de `*`/`/`.

```
10 LET X = 3 + 4 * 2      -> 11
20 LET Y = (3 + 4) * 2    -> 14
30 LET Z = -5 + 3          -> -2
```

### Relacionais

Operadores relacionais retornam 1 (verdadeiro) ou 0 (falso):

`=`, `<>`, `<`, `>`, `<=`, `>=`

Têm a precedência mais baixa.

```
10 IF A <> 0 THEN PRINT "DIFERENTE DE ZERO"
20 LET X = 5 > 3            -> X = 1
```

### Funções Embutidas

| Função   | Descrição |
|----------|-----------|
| `FREE`   | Retorna a memória disponível (linhas restantes) |
| `RND(n)` | Retorna um inteiro aleatório 1 ≤ x ≤ n |
| `ABS(n)` | Retorna o valor absoluto de n |

```
10 PRINT FREE
20 LET R = RND(10)
30 PRINT ABS(-7.5)
```

### Números

Números são armazenados como `float64`. A saída é renderizada como inteiro quando o valor é um número inteiro, caso contrário como decimal.

```
PRINT 42       -> 42
PRINT 10 / 3   -> 3.333333
PRINT 2.5 * 4  -> 10
PRINT 2 / 4    -> 0.5
```

Números de linha 0–63999, valores de variáveis têm alcance completo de float64.

### Strings

Literais de string usam aspas duplas:

```
PRINT "HELLO, WORLD!"
```

Comparação de strings e variáveis de string não são suportadas — strings são apenas para saída.

### Variáveis

26 variáveis de letra única: `A` a `Z`. Todas numéricas (float64), inicializadas com 0.

---

## Dois-Pontos (Múltiplos Comandos)

Use `:` para colocar vários comandos em uma linha:

```
10 LET A = 5 : LET B = 10 : PRINT A + B
20 IF X = 0 THEN PRINT "ZERO" : GOTO 100
```

---

## Tratamento de Erros

- Divisão por zero retorna 0 (sem erro).
- GOTO/GOSUB para linha inexistente exibe uma mensagem e para.
- Estouro da pilha de sub-rotinas (>100 níveis) exibe erro e para.
- Erros de sintaxe exibem o erro com posição e número da linha.

---

## Memória

O `ProgramStore` armazena até 100 linhas por padrão (`FreeMem` informa quantos slots restam).
