// Package diag defines the diagnostic type emitted by the Kolang linter and
// its JSON serialization.
//
// The linter output is consumed by a Persian-language IDE (CodeMirror 6), so
// positions are 1-based and endCol is exclusive (one past the last character).
package diag

import "encoding/json"

// Severity is the diagnostic severity level.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Diagnostic is a single lint finding. Line/Col are the 1-based start
// position; EndLine/EndCol are the 1-based, exclusive end position.
type Diagnostic struct {
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	EndLine  int      `json:"endLine"`
	EndCol   int      `json:"endCol"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Rule     string   `json:"rule"`
}

// At builds a diagnostic with a single-character range (endCol = col+1).
func At(line, col int, sev Severity, msg, rule string) Diagnostic {
	return Diagnostic{
		Line:     line,
		Col:      col,
		EndLine:  line,
		EndCol:   col + 1,
		Severity: sev,
		Message:  msg,
		Rule:     rule,
	}
}

// Result is the top-level JSON document emitted by the CLI.
type Result struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Marshal serializes diagnostics to the stable JSON shape. An empty (or nil)
// slice still produces «{"diagnostics":[]}» — never a bare top-level array.
func Marshal(diags []Diagnostic) ([]byte, error) {
	if diags == nil {
		diags = []Diagnostic{}
	}
	return json.Marshal(Result{Diagnostics: diags})
}
