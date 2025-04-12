package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/evaluator"
	"github.com/carlosrabelo/pbasic/pbasic/internal/lexer"
	"github.com/carlosrabelo/pbasic/pbasic/internal/object"
	"github.com/carlosrabelo/pbasic/pbasic/internal/parser"
)

type ProgramStore struct {
	lines []ast.LineNode
}

func NewProgramStore() *ProgramStore {
	return &ProgramStore{}
}

func (ps *ProgramStore) Add(num int, stmt ast.Statement) {
	for i, l := range ps.lines {
		if l.Number == num {
			if stmt == nil {
				ps.lines = append(ps.lines[:i], ps.lines[i+1:]...)
			} else {
				ps.lines[i].Statement = stmt
			}
			return
		}
	}
	if stmt == nil {
		return
	}
	ps.lines = append(ps.lines, ast.LineNode{Number: num, Statement: stmt})
	sort.Slice(ps.lines, func(i, j int) bool {
		return ps.lines[i].Number < ps.lines[j].Number
	})
}

func (ps *ProgramStore) Clear() {
	ps.lines = nil
}

func (ps *ProgramStore) Find(num int) (*ast.LineNode, bool) {
	idx := sort.Search(len(ps.lines), func(i int) bool {
		return ps.lines[i].Number >= num
	})
	if idx < len(ps.lines) && ps.lines[idx].Number == num {
		return &ps.lines[idx], true
	}
	return nil, false
}

func (ps *ProgramStore) Lines() []ast.LineNode {
	return ps.lines
}

func (ps *ProgramStore) Len() int {
	return len(ps.lines)
}

const progCapacity = 1024

func (ps *ProgramStore) FreeMem() int {
	used := 0
	for _, l := range ps.lines {
		used += 8 + len(l.Statement.String())
	}
	free := progCapacity - used
	if free < 0 {
		free = 0
	}
	return free
}

type REPL struct {
	prog  *ProgramStore
	ctx   *evaluator.EvalContext
	in    *bufio.Reader
	sigCh chan os.Signal
}

func New(in io.Reader, out io.Writer) *REPL {
	ps := NewProgramStore()
	rd := bufio.NewReader(in)
	ctx := evaluator.NewEvalContext(ps, out, rd)
	r := &REPL{prog: ps, ctx: ctx, in: rd}
	r.setupSignals()
	return r
}

func (r *REPL) setupSignals() {
	r.sigCh = make(chan os.Signal, 1)
	signal.Notify(r.sigCh, syscall.SIGINT)
}

func (r *REPL) Run() {
	defer signal.Stop(r.sigCh)
	fmt.Fprintln(r.ctx.Wr, "PicoBasic")

	for {
		if r.interrupted() {
			if !r.ctx.Running {
				fmt.Fprintln(r.ctx.Wr)
			}
			r.ctx.Running = false
		}

		if r.ctx.ShouldExit {
			return
		}
		if !r.ctx.Running {
			fmt.Fprint(r.ctx.Wr, "> ")
		}

		line, err := r.in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(r.ctx.Wr)
			}
			return
		}

		if r.interrupted() {
			r.ctx.Running = false
			continue
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		r.processLine(strings.ToUpper(line))

		if r.ctx.ShouldExit {
			return
		}

		if r.ctx.Running {
			r.execLoop()
		}
	}
}

func (r *REPL) interrupted() bool {
	select {
	case <-r.sigCh:
		return true
	default:
		return false
	}
}

func (r *REPL) processLine(line string) {
	if num, body, ok := parseNumberedLine(line); ok {
		r.handleNumberedLine(num, body)
		return
	}

	stmt := r.parseStmt(line)
	if stmt == nil {
		return
	}

	result := evaluator.Eval(stmt, r.ctx)
	if result != nil && result.Type() == object.ErrorType {
		fmt.Fprintf(r.ctx.Wr, "%s\n", result.Inspect())
		if r.ctx.Running {
			r.ctx.Running = false
		}
	}
}

func (r *REPL) execLoop() {
	for r.ctx.Running {
		if r.interrupted() {
			fmt.Fprintln(r.ctx.Wr)
			r.ctx.Running = false
			return
		}

		prevPC := r.ctx.PC
		line, found := r.prog.Find(prevPC)
		if !found {
			fmt.Fprintf(r.ctx.Wr, "?LINE %d NOT FOUND\n", prevPC)
			r.ctx.Running = false
			return
		}

		result := evaluator.Eval(line.Statement, r.ctx)
		if result != nil && result.Type() == object.ErrorType {
			fmt.Fprintf(r.ctx.Wr, "%s\n", result.Inspect())
			r.ctx.Running = false
			return
		}

		if !r.ctx.Running || r.ctx.ShouldExit {
			return
		}

		if r.ctx.PC == prevPC {
			lines := r.prog.Lines()
			for i, l := range lines {
				if l.Number == prevPC && i+1 < len(lines) {
					r.ctx.PC = lines[i+1].Number
					break
				}
			}
			if r.ctx.PC == prevPC {
				r.ctx.Running = false
			}
		}
	}
}

func (r *REPL) parseStmt(input string) ast.Statement {
	l := lexer.New(input)
	p := parser.New(l)
	stmt := p.ParseLine()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(r.ctx.Wr, "?SYNTAX ERROR: %s\n", err)
		}
		return nil
	}
	return stmt
}

func (r *REPL) handleNumberedLine(num int, body string) {
	if body == "" {
		r.prog.Add(num, nil)
		return
	}
	l := lexer.New(body)
	p := parser.New(l)
	stmt := p.ParseLine()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(r.ctx.Wr, "?SYNTAX ERROR: %s\n", err)
		}
		return
	}
	if stmt != nil {
		r.prog.Add(num, stmt)
	}
}

func parseNumberedLine(line string) (num int, body string, ok bool) {
	if line == "" {
		return 0, "", false
	}
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, "", false
	}
	n := 0
	for j := 0; j < i; j++ {
		n = n*10 + int(line[j]-'0')
	}
	rest := ""
	if i < len(line) {
		rest = strings.TrimSpace(line[i:])
	}
	return n, rest, true
}
