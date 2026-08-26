package rules

import (
	"strings"
	"testing"

	"github.com/faralidev/kolang/pkg/linter"

	"github.com/faralidev/kolang-linter/internal/diag"
)

// runRule lexes and parses src exactly as the registry would, then runs a
// single rule's Check.
func runRule(t *testing.T, r Rule, src string) []diag.Diagnostic {
	t.Helper()
	toks := linter.Lex(src)
	stmts, parseErr := linter.ParseProgram(src)
	return r.Check(src, toks, stmts, parseErr)
}

// hasRule reports whether any diagnostic in ds uses rule id.
func hasRule(ds []diag.Diagnostic, id string) bool {
	for _, d := range ds {
		if d.Rule == id {
			return true
		}
	}
	return false
}

// --- syntax-error ---

func TestSyntaxErrorRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"valid-verb-final", "«سلام» بنویس", 0},
		{"verb-first-invalid", "بنویس «سلام»\n", 1},
		{"incomplete-assignment", "x =", 1},
		{"if-without-copula", "اگر x:", 1},
		{"unexpected-token", "x = )", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, syntaxErrorRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
			for _, d := range got {
				if d.Severity != diag.SeverityError {
					t.Errorf("severity = %q, want error", d.Severity)
				}
				if !strings.HasPrefix(d.Message, "خطای نحوی: ") {
					t.Errorf("message should start with «خطای نحوی: », got %q", d.Message)
				}
			}
		})
	}
}

func TestSyntaxErrorLineExtraction(t *testing.T) {
	// First line valid, second line broken: the error must be reported on line 2.
	src := "«سلام» بنویس\nx ="
	got := runRule(t, syntaxErrorRule{}, src)
	if len(got) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %+v", len(got), got)
	}
	if got[0].Line != 2 {
		t.Errorf("diagnostic line = %d, want 2", got[0].Line)
	}
}

// --- no-implicit-truthiness ---

func TestNoImplicitTruthinessRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"bare-if", "اگر x:", 1},
		{"bare-while", "تاوقتی x:", 1},
		{"comparison-with-copula", "اگر x == 1 باشد:", 0},
		{"copula-only", "اگر x باشد:", 0},
		{"membership", "اگر x در لیست باشد:", 0},
		{"lt-copula", "اگر x < 5 باشد:\n\tx بنویس", 0},
		{"elif-bare", "اگر x == 1 باشد:\n\tمثل\nوگرنه اگر y:", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, noImplicitTruthinessRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
			for _, d := range got {
				if d.Severity != diag.SeverityError {
					t.Errorf("severity = %q, want error", d.Severity)
				}
				if d.Rule != "no-implicit-truthiness" {
					t.Errorf("rule = %q, want no-implicit-truthiness", d.Rule)
				}
			}
		})
	}
}

// --- line-too-long ---

func TestLineTooLongRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"exactly-100", strings.Repeat("ا", 100), 0},
		{"101", strings.Repeat("ا", 101), 1},
		{"long-second-line", "بنویس 1\n" + strings.Repeat("ا", 150), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, lineTooLongRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
			for _, d := range got {
				if d.Severity != diag.SeverityWarning {
					t.Errorf("severity = %q, want warning", d.Severity)
				}
			}
		})
	}
}

func TestLineTooLongPosition(t *testing.T) {
	src := "بنویس 1\n" + strings.Repeat("ا", 120)
	got := runRule(t, lineTooLongRule{}, src)
	if len(got) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(got))
	}
	if got[0].Line != 2 {
		t.Errorf("diagnostic line = %d, want 2", got[0].Line)
	}
}

// --- dot-access ---

func TestDotAccessRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"latin-dot", "a.b = 1", 1},
		{"persian-dot", "نام.فامیل بنویس", 1},
		{"decimal", "بنویس 3.14", 0},
		{"no-dot", "x = 1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, dotAccessRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- negation-no-bang-eq ---

func TestNegationNoBangEqRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"bang-eq", "اگر x != 1 باشد:", 1},
		{"eq-ok", "اگر x == 1 باشد:", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, negationNoBangEqRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- undefined-variable ---

func TestUndefinedVariableRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"defined", "x = 1\nx بنویس", 0},
		{"undefined", "y بنویس", 1},
		{"attribute-not-flagged", "zzzِ خود بنویس", 0},
		{"param-defined", "تعریف f(الف):\n\tالف بنویس", 0},
		{"parse-error-skipped", "اگر x:", 0},
		{"keyword-builtin", "بنویس 5", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, undefinedVariableRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- unused-variable ---

func TestUnusedVariableRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"unused", "x = 1", 1},
		{"used", "x = 1\nx بنویس", 0},
		{"one-unused-of-two", "x = 1\ny = 2\nx بنویس", 1},
		{"param-unused", "تعریف f(الف):\n\tمثل", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, unusedVariableRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- naming-convention ---

func TestNamingConventionRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"mixed", "myVarنام = 1", 1},
		{"latin-only", "myVar = 1", 0},
		{"persian-only", "متغیر = 1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, namingConventionRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- duplicate-import ---

func TestDuplicateImportRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"single", "ریاضی بیار", 0},
		{"duplicate", "ریاضی بیار\nریاضی بیار", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, duplicateImportRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
		})
	}
}

// --- mixed-indentation ---

func TestMixedIndentationRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"space-then-tab", " \tبنویس 1", 1},
		{"tab-then-space", "\t بنویس 1", 1},
		{"spaces-only", "    بنویس 1", 0},
		{"tab-only", "\tبنویس 1", 0},
		{"two-lines-one-mixed", "بنویس 1\n \tبنویس 2", 1},
		{"empty-line-no-crash", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, mixedIndentationRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
			for _, d := range got {
				if d.Severity != diag.SeverityWarning {
					t.Errorf("severity = %q, want warning", d.Severity)
				}
			}
		})
	}
}

// --- trailing-whitespace ---

func TestTrailingWhitespaceRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"trailing-spaces", "بنویس 1   ", 1},
		{"trailing-tab", "بنویس 1\t", 1},
		{"clean", "بنویس 1", 0},
		{"two-lines-one-dirty", "بنویس 1\nبنویس 2 \n", 1},
		{"whitespace-only-line", "   \n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runRule(t, trailingWhitespaceRule{}, tt.src)
			if len(got) != tt.want {
				t.Fatalf("got %d diagnostics, want %d: %+v", len(got), tt.want, got)
			}
			for _, d := range got {
				if d.Severity != diag.SeverityInfo {
					t.Errorf("severity = %q, want info", d.Severity)
				}
			}
		})
	}
}

// --- registry integration ---

func TestRegistryValidCode(t *testing.T) {
	got := NewRegistry().Run("«سلام» بنویس")
	if len(got) != 0 {
		t.Fatalf("valid code produced diagnostics: %+v", got)
	}
}

func TestRegistryIfWithoutCopula(t *testing.T) {
	got := NewRegistry().Run("اگر x:")
	if !hasRule(got, "no-implicit-truthiness") {
		t.Errorf("expected a no-implicit-truthiness diagnostic, got: %+v", got)
	}
	if !hasRule(got, "syntax-error") {
		t.Errorf("expected a syntax-error diagnostic too, got: %+v", got)
	}
}
