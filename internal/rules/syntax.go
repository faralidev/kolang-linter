package rules

import (
	"regexp"
	"strconv"

	"github.com/faralidev/kolang/pkg/linter"

	"github.com/faralidev/kolang-linter/internal/diag"
)

// lineRe extracts the line number from parser error messages, which are
// produced either by errf («خطای نحو در خط N: ...») or, in a few places, as
// «خط N: ...».
var lineRe = regexp.MustCompile(`خط[^0-9]*(\d+)`)

// --- Rule 1: syntax-error ---

type syntaxErrorRule struct{}

func (syntaxErrorRule) ID() string { return "syntax-error" }

func (syntaxErrorRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	if parseErr == nil {
		return nil
	}
	line := 1
	if m := lineRe.FindStringSubmatch(parseErr.Error()); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			line = n
		}
	}
	return []diag.Diagnostic{
		diag.At(line, 1, diag.SeverityError, "خطای نحوی: "+parseErr.Error(), "syntax-error"),
	}
}

// --- Rule 2: unclosed-string ---

type unclosedStringRule struct{}

func (unclosedStringRule) ID() string { return "unclosed-string" }

// The lexer already flags an unterminated «...» literal as an ILLEGAL token
// with literal «متن بسته نشده»; this rule turns it into a proper diagnostic.
func (unclosedStringRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	var out []diag.Diagnostic
	for _, t := range toks {
		if t.Type == linter.ILLEGAL && t.Literal == "متن بسته نشده" {
			out = append(out, diag.At(t.Line, t.Col, diag.SeverityError,
				"متن بسته نشده است — « بدون » پایانی", "unclosed-string"))
		}
	}
	return out
}

// --- Rule 3: unclosed-comment ---

type unclosedCommentRule struct{}

func (unclosedCommentRule) ID() string { return "unclosed-comment" }

// A «//» block comment that never reaches its closing «//» is reported by the
// lexer as an ILLEGAL token with literal «کامنت بلوک بسته نشده».
func (unclosedCommentRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	var out []diag.Diagnostic
	for _, t := range toks {
		if t.Type == linter.ILLEGAL && t.Literal == "کامنت بلوک بسته نشده" {
			out = append(out, diag.At(t.Line, t.Col, diag.SeverityWarning,
				"کامنت بلوک // بسته نشده است — با // دوم ببندید", "unclosed-comment"))
		}
	}
	return out
}