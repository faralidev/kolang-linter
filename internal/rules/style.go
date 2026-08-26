package rules

import (
	"strings"

	"github.com/faralidev/kolang/pkg/linter"

	"github.com/faralidev/kolang-linter/internal/diag"
)

// --- Rule 4: no-implicit-truthiness ---

// Per the Kolang spec, conditions MUST be an explicit symbolic comparison plus
// the copula (باشد/نباشد). A bare expression («اگر x:») is invalid.
type noImplicitTruthinessRule struct{}

func (noImplicitTruthinessRule) ID() string { return "no-implicit-truthiness" }

func (noImplicitTruthinessRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	var out []diag.Diagnostic
	for i, t := range toks {
		if t.Type != linter.IF && t.Type != linter.WHILE {
			continue
		}
		// Scan forward from إذا/تاوقتی to the condition's terminating ':' at
		// bracket depth 0. If the region contains neither a comparison operator
		// nor the copula, the condition is an implicit truthiness test.
		hasCmp, hasCop := false, false
		depth := 0
		for j := i + 1; j < len(toks); j++ {
			tk := toks[j]
			switch tk.Type {
			case linter.LPAREN, linter.LBRACKET, linter.LBRACE:
				depth++
			case linter.RPAREN, linter.RBRACKET, linter.RBRACE:
				if depth > 0 {
					depth--
				}
			case linter.COLON:
				if depth == 0 {
					goto done
				}
			case linter.EQ, linter.LT, linter.GT, linter.LTE, linter.GTE, linter.IN:
				hasCmp = true
			case linter.COP_POS, linter.COP_NEG:
				hasCop = true
			}
		}
	done:
		if !hasCmp && !hasCop {
			out = append(out, diag.At(t.Line, t.Col, diag.SeverityError,
				"شرط باید شامل مقایسه و باشد/نباشد باشد — مقایسه ضمنی مجاز نیست",
				"no-implicit-truthiness"))
		}
	}
	return out
}

// --- Rule 5: negation-no-bang-eq ---

// Per the spec, «!=» is removed; نباشد is the sole negator.
type negationNoBangEqRule struct{}

func (negationNoBangEqRule) ID() string { return "negation-no-bang-eq" }

func (negationNoBangEqRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	rs := []rune(src)
	var out []diag.Diagnostic
	for i := 0; i+1 < len(rs); i++ {
		if rs[i] == '!' && rs[i+1] == '=' {
			line, col := posOf(rs, i)
			out = append(out, diag.Diagnostic{
				Line:     line,
				Col:      col,
				EndLine:  line,
				EndCol:   col + 2,
				Severity: diag.SeverityError,
				Message:  "عملگر != حذف شده — از == ... نباشد استفاده کنید",
				Rule:     "negation-no-bang-eq",
			})
		}
	}
	return out
}

// --- Rule 6: dot-access ---

// Dot access for members is invalid; ezafe (ِ) must be used. The '.' is only
// a decimal separator, so only a '.' flanked by letters on both sides is
// flagged.
type dotAccessRule struct{}

func (dotAccessRule) ID() string { return "dot-access" }

func (dotAccessRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	rs := []rune(src)
	var out []diag.Diagnostic
	for i, r := range rs {
		if r != '.' {
			continue
		}
		if i == 0 || i+1 >= len(rs) {
			continue
		}
		if !isLetterRune(rs[i-1]) || !isLetterRune(rs[i+1]) {
			continue // numeric decimal point or bare dot
		}
		// Span the whole «ident.ident» for a useful range.
		start := i
		for start > 0 && isIdentPartRune(rs[start-1]) {
			start--
		}
		end := i + 2
		for end < len(rs) && isIdentPartRune(rs[end]) {
			end++
		}
		line, col := posOf(rs, start)
		endLine, endCol := posOf(rs, end)
		out = append(out, diag.Diagnostic{
			Line:     line,
			Col:      col,
			EndLine:  endLine,
			EndCol:   endCol,
			Severity: diag.SeverityError,
			Message:  "دسترسی با نقطه مجاز نیست — از اضافه (ِ) استفاده کنید",
			Rule:     "dot-access",
		})
	}
	return out
}

// --- Rule 7: line-too-long ---

type lineTooLongRule struct{}

func (lineTooLongRule) ID() string { return "line-too-long" }

func (lineTooLongRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	lines := strings.Split(src, "\n")
	var out []diag.Diagnostic
	for i, ln := range lines {
		if len([]rune(ln)) > 100 {
			out = append(out, diag.At(i+1, 1, diag.SeverityWarning,
				"خط بیش از حد بلند است (>۱۰۰ نویسه)", "line-too-long"))
		}
	}
	return out
}

// --- Rule 8: mixed-indentation ---

type mixedIndentationRule struct{}

func (mixedIndentationRule) ID() string { return "mixed-indentation" }

func (mixedIndentationRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	rs := []rune(src)
	var out []diag.Diagnostic
	line, col := 1, 1
	atLineStart := true
	inIndent := false
	indentStartCol := 1
	hasSpace, hasTab := false, false

	flush := func() {
		if hasSpace && hasTab {
			out = append(out, diag.At(line, indentStartCol, diag.SeverityWarning,
				"ترکیب فاصله و تب در تورفتگی", "mixed-indentation"))
		}
		hasSpace, hasTab = false, false
	}

	for i := 0; i < len(rs); i++ {
		r := rs[i]
		if atLineStart {
			atLineStart = false
			inIndent = true
			hasSpace, hasTab = false, false
			indentStartCol = col
		}
		if inIndent {
			switch r {
			case ' ':
				hasSpace = true
				col++
				continue
			case '\t':
				hasTab = true
				col++
				continue
			}
			inIndent = false
			flush()
		}
		if r == '\n' {
			line++
			col = 1
			atLineStart = true
			continue
		}
		col++
	}
	flush()
	return out
}

// --- Rule 9: trailing-whitespace ---

type trailingWhitespaceRule struct{}

func (trailingWhitespaceRule) ID() string { return "trailing-whitespace" }

func (trailingWhitespaceRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	rs := []rune(src)
	var out []diag.Diagnostic
	line, col := 1, 1
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		if r == '\n' {
			line++
			col = 1
			continue
		}
		if r == ' ' || r == '\t' {
			j := i
			for j < len(rs) && (rs[j] == ' ' || rs[j] == '\t') {
				j++
			}
			if j >= len(rs) || rs[j] == '\n' {
				out = append(out, diag.At(line, col, diag.SeverityInfo,
					"فاصلهٔ پایانی خط", "trailing-whitespace"))
				col += (j - i)
				i = j - 1
				continue
			}
		}
		col++
	}
	return out
}
