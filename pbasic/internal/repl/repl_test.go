package repl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeBas(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "prog.bas")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestRunFile(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantOut string
		wantErr string
	}{
		{
			name:    "hello",
			source:  "10 PRINT \"HELLO, WORLD!\"\n20 END\n",
			wantOut: "HELLO, WORLD!\n",
		},
		{
			name: "unnumbered lines ignored",
			source: "" +
				"10 PRINT \"A\"\n" +
				"LIST\n" +
				"20 PRINT \"B\"\n" +
				"RUN\n",
			wantOut: "A\nB\n",
		},
		{
			name:    "empty file",
			source:  "",
			wantErr: "NO PROGRAM",
		},
		{
			name:    "only unnumbered commands",
			source:  "LIST\nRUN\n",
			wantErr: "NO PROGRAM",
		},
		{
			name:    "blank lines only",
			source:  "\n\n  \n",
			wantErr: "NO PROGRAM",
		},
		{
			name:    "syntax error does not run",
			source:  "10 LET\n20 PRINT \"SHOULD NOT RUN\"\n",
			wantErr: "?SYNTAX ERROR:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			r := New(strings.NewReader(""), &out)
			err := r.RunFile(writeBas(t, tt.source))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RunFile() error = %v", err)
				}
				if got := out.String(); got != tt.wantOut {
					t.Errorf("stdout = %q, want %q", got, tt.wantOut)
				}
				return
			}
			if err == nil {
				t.Fatalf("RunFile() error = nil, want %q; stdout = %q", tt.wantErr, out.String())
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
			if strings.Contains(out.String(), "SHOULD NOT RUN") {
				t.Errorf("program ran after load error; stdout = %q", out.String())
			}
		})
	}
}

func TestLoadFileMissing(t *testing.T) {
	var out bytes.Buffer
	r := New(strings.NewReader(""), &out)
	err := r.LoadFile(filepath.Join(t.TempDir(), "missing.bas"))
	if err == nil {
		t.Fatal("LoadFile() error = nil, want missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("LoadFile() error = %v, want os.IsNotExist", err)
	}
}
