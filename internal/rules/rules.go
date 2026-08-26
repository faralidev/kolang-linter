// Package rules implements the Kolang linter rules.
//
// All rules share a single Check signature. The registry lexes and parses the
// source exactly once and hands the results to every rule; a rule only uses
// the inputs it needs.
package rules

import (
	"sort"

	"github.com/faralidev/kolang/pkg/linter"

	"github.com/faralidev/kolang-linter/internal/diag"
)

// Rule is a single lint rule. src is the raw source, toks its token stream,
// ast_ the parsed program (nil when parseErr != nil), and parseErr the parser
// error (nil on success).
type Rule interface {
	ID() string
	Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic
}

// Registry owns the rule set and runs all rules over one source string.
type Registry struct {
	rules []Rule
}

// NewRegistry builds the default rule set: syntax, then style, then semantic.
func NewRegistry() *Registry {
	return &Registry{
		rules: []Rule{
			syntaxErrorRule{},
			unclosedStringRule{},
			unclosedCommentRule{},

			noImplicitTruthinessRule{},
			negationNoBangEqRule{},
			dotAccessRule{},
			lineTooLongRule{},
			mixedIndentationRule{},
			trailingWhitespaceRule{},

			undefinedVariableRule{},
			unusedVariableRule{},
			namingConventionRule{},
			duplicateImportRule{},
		},
	}
}

// Run lexes and parses src once, runs every rule, and returns all diagnostics
// sorted by (line, col, rule) for deterministic output.
func (r *Registry) Run(src string) []diag.Diagnostic {
	toks := linter.Lex(src)
	stmts, parseErr := linter.ParseProgram(src)

	var out []diag.Diagnostic
	for _, rule := range r.rules {
		out = append(out, rule.Check(src, toks, stmts, parseErr)...)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		if out[i].Col != out[j].Col {
			return out[i].Col < out[j].Col
		}
		return out[i].Rule < out[j].Rule
	})
	return out
}

// posOf returns the 1-based line/column of rune index idx in rs.
func posOf(rs []rune, idx int) (line, col int) {
	if idx > len(rs) {
		idx = len(rs)
	}
	line, col = 1, 1
	for i := 0; i < idx; i++ {
		if rs[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// isLetterRune reports whether r is a Latin or Arabic/Persian letter
// (combining diacritics such as the kasra/ezafe are excluded), mirroring the
// lexer's notion of an identifier-start rune.
func isLetterRune(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	if r >= 0x0600 && r <= 0x06FF {
		if r >= 0x064B && r <= 0x065F {
			return false
		}
		return true
	}
	return (r >= 0xFB50 && r <= 0xFDFF) || (r >= 0xFE70 && r <= 0xFEFF)
}

// isIdentPartRune reports whether r may appear inside an identifier: letters,
// digits, underscore, or ZWNJ.
func isIdentPartRune(r rune) bool {
	if r == '_' || r == 0x200C {
		return true
	}
	if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	return isLetterRune(r)
}