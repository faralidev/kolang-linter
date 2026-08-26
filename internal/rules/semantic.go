package rules

import (
	"strconv"

	"github.com/faralidev/kolang/pkg/linter"

	"github.com/faralidev/kolang-linter/internal/diag"
)

// kolangBuiltins are names that may legitimately appear in expression
// position without a definition in the current file: keyword-derived names
// (بنویس/خود/والد/بسته‌است/ببند), common type annotations and stdlib names.
var kolangBuiltins = map[string]bool{
	"بنویس":    true,
	"خود":      true,
	"والد":     true,
	"بسته‌است": true,
	"بستهاست":  true,
	"ببند":     true,
	"بازکردن":  true,
	"محدوده":   true,
	"صحیح":     true,
	"اعشاری":   true,
	"متن":      true,
	"منطقی":    true,
	"بولی":     true,
	"فهرست":    true,
	"مجموعه":   true,
	"فرهنگ":    true,
	"گروه":     true,
	"تهی":      true,
	"کانال":    true,
	"خطا":      true,
	"استثنا":   true,
}

// bindingKind classifies a name-binding site so the unused-variable analysis
// can ignore function/class/module declarations (which are not "variables").
type bindingKind int

const (
	bindVariable bindingKind = iota
	bindFunction
	bindClass
	bindModule
	bindCompVar
)

// --- AST traversal with binding/usage callbacks ---

// walkStmts visits a statement list. bind is called at name-binding sites
// (with the kind of binding), use at expression-position identifier
// references. Either callback may be nil.
func walkStmts(stmts []linter.Stmt, bind func(name string, line int, kind bindingKind), use func(name string, line int)) {
	for _, s := range stmts {
		walkStmt(s, bind, use)
	}
}

func walkBlock(b *linter.Block, bind func(string, int, bindingKind), use func(string, int)) {
	if b == nil {
		return
	}
	walkStmts(b.Stmts, bind, use)
}

func walkBindTarget(e linter.Expr, bind func(string, int, bindingKind), use func(string, int)) {
	if id, ok := e.(*linter.Ident); ok {
		if bind != nil {
			bind(id.Name, id.L, bindVariable)
		}
		return
	}
	// Non-trivial targets (e.g. a[0] = 5) are walked as expressions: the
	// base identifier is a usage, not a definition.
	walkExpr(e, bind, use)
}

func walkStmt(s linter.Stmt, bind func(string, int, bindingKind), use func(string, int)) {
	switch n := s.(type) {
	case nil:
		return
	case *linter.ExprStmt:
		walkExpr(n.Expr, bind, use)
	case *linter.Assign:
		walkBindTarget(n.Target, bind, use)
		walkExpr(n.Value, bind, use)
	case *linter.MultiAssign:
		for _, t := range n.Targets {
			walkBindTarget(t, bind, use)
		}
		for _, v := range n.Values {
			walkExpr(v, bind, use)
		}
	case *linter.CompoundAssign:
		// Reads and writes the target; treat as a usage.
		walkExpr(n.Target, bind, use)
		walkExpr(n.Value, bind, use)
	case *linter.PrintStmt:
		for _, a := range n.Args {
			walkExpr(a, bind, use)
		}
	case *linter.InputStmt:
		walkBindTarget(n.Target, bind, use)
	case *linter.ReturnStmt:
		for _, v := range n.Vals {
			walkExpr(v, bind, use)
		}
	case *linter.IfStmt:
		walkExpr(n.Cond, bind, use)
		walkStmts(n.Body, bind, use)
		for _, e := range n.Elifs {
			walkExpr(e.Cond, bind, use)
			walkBlock(e.Body, bind, use)
		}
		walkBlock(n.Else, bind, use)
	case *linter.WhileStmt:
		walkExpr(n.Cond, bind, use)
		walkBlock(n.Body, bind, use)
	case *linter.ForRange:
		walkBindTarget(n.Var, bind, use)
		walkExpr(n.Start, bind, use)
		walkExpr(n.End, bind, use)
		walkExpr(n.Step, bind, use)
		walkBlock(n.Body, bind, use)
	case *linter.ForIn:
		for _, v := range n.Vars {
			walkBindTarget(v, bind, use)
		}
		walkExpr(n.Iter, bind, use)
		walkBlock(n.Body, bind, use)
	case *linter.WithStmt:
		walkExpr(n.Context, bind, use)
		if n.Name != "" && bind != nil {
			bind(n.Name, n.L, bindVariable)
		}
		walkStmts(n.Body, bind, use)
	case *linter.DefStmt:
		if bind != nil {
			bind(n.Name, n.L, bindFunction)
			for _, p := range n.Params {
				bind(p.Name, p.L, bindVariable)
			}
		}
		for _, p := range n.Params {
			if p.Default != nil {
				walkExpr(p.Default, bind, use)
			}
		}
		walkBlock(n.Body, bind, use)
	case *linter.ClassDef:
		if bind != nil {
			bind(n.Name, n.L, bindClass)
		}
		walkStmts(n.Body, bind, use)
	case *linter.InterfaceDef:
		if bind != nil {
			bind(n.Name, n.L, bindClass)
			for _, m := range n.Methods {
				bind(m.Name, m.L, bindFunction)
				for _, p := range m.Params {
					bind(p.Name, p.L, bindVariable)
				}
			}
		}
	case *linter.ImportStmt:
		if bind != nil {
			bind(n.Module, n.L, bindModule)
		}
	case *linter.FromImportStmt:
		if bind != nil {
			bind(n.Module, n.L, bindModule)
			bind(n.Name, n.L, bindModule)
			if n.Alias != "" {
				bind(n.Alias, n.L, bindModule)
			}
		}
	case *linter.GlobalStmt:
		if bind != nil {
			for _, nm := range n.Names {
				bind(nm, n.L, bindVariable)
			}
		}
	case *linter.NonlocalStmt:
		if bind != nil {
			for _, nm := range n.Names {
				bind(nm, n.L, bindVariable)
			}
		}
	case *linter.GoStmt:
		walkExpr(n.Expr, bind, use)
	case *linter.SendStmt:
		walkExpr(n.Channel, bind, use)
		walkExpr(n.Value, bind, use)
	case *linter.CloseStmt:
		walkExpr(n.Channel, bind, use)
	case *linter.RaiseStmt:
		walkExpr(n.Value, bind, use)
	case *linter.AppendStmt:
		walkExpr(n.List, bind, use)
		walkExpr(n.Value, bind, use)
	case *linter.RemoveStmt:
		walkExpr(n.List, bind, use)
		walkExpr(n.Value, bind, use)
	case *linter.TryStmt:
		walkBlock(n.Body, bind, use)
		for _, h := range n.Handlers {
			// h.Exception is a type name position — deliberately not walked.
			if h.Alias != "" && bind != nil {
				bind(h.Alias, h.L, bindVariable)
			}
			walkBlock(h.Body, bind, use)
		}
		walkBlock(n.Finally, bind, use)
	case *linter.DeferStmt:
		walkExpr(n.Call, bind, use)
	case *linter.YieldStmt:
		walkExpr(n.Value, bind, use)
	case *linter.YieldFromStmt:
		walkExpr(n.Value, bind, use)
	case *linter.DecoratorStmt:
		if use != nil {
			use(n.Name, n.L)
		}
		for _, a := range n.Args {
			walkExpr(a, bind, use)
		}
	}
}

func walkExpr(e linter.Expr, bind func(string, int, bindingKind), use func(string, int)) {
	switch n := e.(type) {
	case nil:
		return
	case *linter.Ident:
		if use != nil {
			use(n.Name, n.L)
		}
	case *linter.NumberLit, *linter.StrLit, *linter.BoolLit, *linter.NoneLit:
		return
	case *linter.Unary:
		walkExpr(n.Expr, bind, use)
	case *linter.BinaryOp:
		walkExpr(n.Left, bind, use)
		walkExpr(n.Right, bind, use)
	case *linter.Call:
		walkExpr(n.Fn, bind, use)
		for _, a := range n.Args {
			walkExpr(a, bind, use)
		}
		for _, k := range n.KwArgs {
			walkExpr(k.Value, bind, use)
		}
	case *linter.Index:
		walkExpr(n.Target, bind, use)
		walkExpr(n.Index, bind, use)
	case *linter.Slice:
		walkExpr(n.Target, bind, use)
		walkExpr(n.Low, bind, use)
		walkExpr(n.High, bind, use)
		walkExpr(n.Step, bind, use)
	case *linter.MemberAccess:
		// The receiver is a variable reference; the attribute is a member
		// name, not a variable.
		walkExpr(n.Receiver, bind, use)
	case *linter.MethodCall:
		walkExpr(n.Receiver, bind, use)
		for _, a := range n.Args {
			walkExpr(a, bind, use)
		}
	case *linter.ListLit:
		for _, el := range n.Elems {
			walkExpr(el, bind, use)
		}
	case *linter.TupleLit:
		for _, el := range n.Elems {
			walkExpr(el, bind, use)
		}
	case *linter.DictLit:
		for _, k := range n.Keys {
			walkExpr(k, bind, use)
		}
		for _, v := range n.Values {
			walkExpr(v, bind, use)
		}
	case *linter.SetLit:
		for _, el := range n.Elems {
			walkExpr(el, bind, use)
		}
	case *linter.PipeExpr:
		walkExpr(n.Left, bind, use)
		walkExpr(n.Right, bind, use)
	case *linter.TernaryExpr:
		walkExpr(n.Cond, bind, use)
		walkExpr(n.TrueBranch, bind, use)
		walkExpr(n.FalseBranch, bind, use)
	case *linter.ListComp:
		if bind != nil {
			for _, c := range n.Clauses {
				bind(c.Name, c.L, bindCompVar)
			}
		}
		walkExpr(n.Element, bind, use)
		for _, c := range n.Clauses {
			walkExpr(c.Iterable, bind, use)
			walkExpr(c.Filter, bind, use)
		}
	case *linter.DictComp:
		if bind != nil {
			for _, c := range n.Clauses {
				bind(c.Name, c.L, bindCompVar)
			}
		}
		walkExpr(n.Key, bind, use)
		walkExpr(n.Value, bind, use)
		for _, c := range n.Clauses {
			walkExpr(c.Iterable, bind, use)
			walkExpr(c.Filter, bind, use)
		}
	case *linter.SetComp:
		if bind != nil {
			for _, c := range n.Clauses {
				bind(c.Name, c.L, bindCompVar)
			}
		}
		walkExpr(n.Element, bind, use)
		for _, c := range n.Clauses {
			walkExpr(c.Iterable, bind, use)
			walkExpr(c.Filter, bind, use)
		}
	case *linter.GenExp:
		if bind != nil {
			for _, c := range n.Clauses {
				bind(c.Name, c.L, bindCompVar)
			}
		}
		walkExpr(n.Element, bind, use)
		for _, c := range n.Clauses {
			walkExpr(c.Iterable, bind, use)
			walkExpr(c.Filter, bind, use)
		}
	case *linter.ChannelLit:
		// n.Type is a type annotation position — deliberately not walked.
		walkExpr(n.Size, bind, use)
	case *linter.RecvExpr:
		walkExpr(n.Channel, bind, use)
	case *linter.YieldExpr:
		walkExpr(n.Value, bind, use)
	case *linter.YieldFromExpr:
		walkExpr(n.Value, bind, use)
	}
}

// --- Rule 10: undefined-variable ---

type undefinedVariableRule struct{}

func (undefinedVariableRule) ID() string { return "undefined-variable" }

func (undefinedVariableRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	if ast_ == nil {
		return nil // parse failed; the syntax rules own the reporting
	}
	defs := map[string]bool{}
	walkStmts(ast_, func(name string, line int, kind bindingKind) { defs[name] = true }, nil)

	var out []diag.Diagnostic
	walkStmts(ast_, nil, func(name string, line int) {
		if defs[name] || kolangBuiltins[name] || isExceptionName(name) {
			return
		}
		out = append(out, diag.At(line, 1, diag.SeverityWarning,
			"متغیر «"+name+"» تعریف نشده است", "undefined-variable"))
	})
	return out
}

// isExceptionName reports whether a name looks like an exception type
// (خطای…). Exception types are often raised without a prior definition.
func isExceptionName(name string) bool {
	return len(name) >= 4 && (name[:4] == "خطا" || name[:4] == "خطای")
}

// --- Rule 11: unused-variable ---

type unusedVariableRule struct{}

func (unusedVariableRule) ID() string { return "unused-variable" }

type occurrence struct {
	name  string
	line  int
	isUse bool
}

func (unusedVariableRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	if ast_ == nil {
		return nil
	}

	// One walk records every variable binding and every usage. Function,
	// class and module declarations are deliberately excluded: they are not
	// "variables".
	var occs []occurrence
	walkStmts(ast_,
		func(name string, line int, kind bindingKind) {
			if kind == bindVariable {
				occs = append(occs, occurrence{name: name, line: line})
			}
		},
		func(name string, line int) {
			occs = append(occs, occurrence{name: name, line: line, isUse: true})
		})

	firstBind := map[string]int{}
	var order []string
	for _, o := range occs {
		if o.isUse {
			continue
		}
		if _, ok := firstBind[o.name]; !ok {
			firstBind[o.name] = o.line
			order = append(order, o.name)
		}
	}

	var out []diag.Diagnostic
	for _, name := range order {
		if kolangBuiltins[name] {
			continue
		}
		used := false
		for _, o := range occs {
			if o.name != name {
				continue
			}
			// A usage anywhere, or a rebinding on a different line, counts as
			// "used" — keeps the rule conservative.
			if o.isUse || o.line != firstBind[name] {
				used = true
				break
			}
		}
		if !used {
			out = append(out, diag.At(firstBind[name], 1, diag.SeverityWarning,
				"متغیر «"+name+"» استفاده نشده است", "unused-variable"))
		}
	}
	return out
}

// --- Rule 12: naming-convention ---

type namingConventionRule struct{}

func (namingConventionRule) ID() string { return "naming-convention" }

func (namingConventionRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	var out []diag.Diagnostic
	for _, t := range toks {
		if t.Type != linter.IDENT {
			continue
		}
		if hasMixedScript(t.Literal) {
			out = append(out, diag.At(t.Line, t.Col, diag.SeverityInfo,
				"ترکیب حروف لاتین و فارسی در نام", "naming-convention"))
		}
	}
	return out
}

func hasMixedScript(s string) bool {
	var latin, persian bool
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			latin = true
		case isPersianScriptRune(r):
			persian = true
		}
	}
	return latin && persian
}

func isPersianScriptRune(r rune) bool {
	if r >= 0x0600 && r <= 0x06FF {
		// Exclude combining diacritics (kasra/ezafe and friends).
		if r >= 0x064B && r <= 0x065F {
			return false
		}
		return true
	}
	return (r >= 0xFB50 && r <= 0xFDFF) || (r >= 0xFE70 && r <= 0xFEFF)
}

// --- Rule 13: duplicate-import ---

type duplicateImportRule struct{}

func (duplicateImportRule) ID() string { return "duplicate-import" }

func (duplicateImportRule) Check(src string, toks []linter.Token, ast_ []linter.Stmt, parseErr error) []diag.Diagnostic {
	if ast_ == nil {
		return nil
	}
	seen := map[string]int{}
	var out []diag.Diagnostic
	walkAllStmts(ast_, func(s linter.Stmt) {
		var module string
		switch n := s.(type) {
		case *linter.ImportStmt:
			module = n.Module
		case *linter.FromImportStmt:
			module = n.Module
		default:
			return
		}
		if prev, ok := seen[module]; ok {
			out = append(out, diag.At(s.Line(), 1, diag.SeverityWarning,
				"ماژول «"+module+"» دوبار وارد شده است (بار اول در خط "+strconv.Itoa(prev)+")",
				"duplicate-import"))
			return
		}
		seen[module] = s.Line()
	})
	return out
}

// walkAllStmts visits every statement in the tree (including nested blocks)
// and calls fn on each.
func walkAllStmts(stmts []linter.Stmt, fn func(linter.Stmt)) {
	for _, s := range stmts {
		fn(s)
		switch n := s.(type) {
		case *linter.IfStmt:
			walkAllStmts(n.Body, fn)
			for _, e := range n.Elifs {
				if e.Body != nil {
					walkAllStmts(e.Body.Stmts, fn)
				}
			}
			if n.Else != nil {
				walkAllStmts(n.Else.Stmts, fn)
			}
		case *linter.WhileStmt:
			walkAllStmts(n.Body.Stmts, fn)
		case *linter.ForRange:
			walkAllStmts(n.Body.Stmts, fn)
		case *linter.ForIn:
			walkAllStmts(n.Body.Stmts, fn)
		case *linter.WithStmt:
			walkAllStmts(n.Body, fn)
		case *linter.DefStmt:
			if n.Body != nil {
				walkAllStmts(n.Body.Stmts, fn)
			}
		case *linter.ClassDef:
			walkAllStmts(n.Body, fn)
		case *linter.TryStmt:
			if n.Body != nil {
				walkAllStmts(n.Body.Stmts, fn)
			}
			for _, h := range n.Handlers {
				walkAllStmts(h.Body.Stmts, fn)
			}
			if n.Finally != nil {
				walkAllStmts(n.Finally.Stmts, fn)
			}
		}
	}
}
