package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/parser"
)

func TestParseImportsAndModuleConstantPrecedence(t *testing.T) {
	source := `fichar "./lib.cr7" como tactics; leyenda total = 1 + 2 * 3;`
	program := parseSource(t, source)
	if len(program.Imports) != 1 || program.Imports[0].Path != "./lib.cr7" || program.Imports[0].Alias != "tactics" {
		t.Fatalf("imports = %#v, want one ./lib.cr7 import as tactics", program.Imports)
	}
	if len(program.Declarations) != 1 {
		t.Fatalf("declarations = %#v, want one module declaration", program.Declarations)
	}
	constant, ok := program.Declarations[0].(*ast.ConstDecl)
	if !ok || constant.Name != "total" {
		t.Fatalf("module declaration = %#v, want constant total", program.Declarations[0])
	}
	wantStart := strings.Index(source, "leyenda")
	wantEnd := strings.LastIndex(source, ";") + 1
	if constant.Start.Offset != wantStart || constant.End.Offset != wantEnd {
		t.Fatalf("constant span = %d..%d, want %d..%d", constant.Start.Offset, constant.End.Offset, wantStart, wantEnd)
	}
	addition, ok := constant.Initializer.(*ast.BinaryExpr)
	if !ok || addition.Operator != lexer.Plus {
		t.Fatalf("initializer = %#v, want addition at the root", constant.Initializer)
	}
	multiplication, ok := addition.Right.(*ast.BinaryExpr)
	if !ok || multiplication.Operator != lexer.Star {
		t.Fatalf("right operand = %#v, want multiplication", addition.Right)
	}
}

func TestParsePostfixAndCollectionExpressions(t *testing.T) {
	program := parseSource(t, `leyenda result = engine.players[0].score(7); leyenda squad = ["CR", {"name": "Benzema"}];`)
	if len(program.Declarations) != 2 {
		t.Fatalf("declarations = %d, want 2", len(program.Declarations))
	}
	first := program.Declarations[0].(*ast.ConstDecl).Initializer
	call, ok := first.(*ast.CallExpr)
	if !ok {
		t.Fatalf("first initializer = %#v, want call", first)
	}
	property, ok := call.Callee.(*ast.PropertyExpr)
	if !ok || property.Name != "score" {
		t.Fatalf("call callee = %#v, want .score property", call.Callee)
	}
	index, ok := property.Object.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("property receiver = %#v, want index access", property.Object)
	}
	players, ok := index.Collection.(*ast.PropertyExpr)
	if !ok || players.Name != "players" {
		t.Fatalf("indexed collection = %#v, want .players property", index.Collection)
	}

	list, ok := program.Declarations[1].(*ast.ConstDecl).Initializer.(*ast.ListExpr)
	if !ok || len(list.Elements) != 2 {
		t.Fatalf("second initializer = %#v, want two-element list", program.Declarations[1])
	}
	dictionary, ok := list.Elements[1].(*ast.DictExpr)
	if !ok || len(dictionary.Entries) != 1 || dictionary.Entries[0].Key != "name" {
		t.Fatalf("second list element = %#v, want name dictionary entry", list.Elements[1])
	}
}

func TestParseLogicalOperatorPrecedence(t *testing.T) {
	program := parseSource(t, `leyenda result = gol o fuera y no (1 < 2) == fuera;`)
	root := program.Declarations[0].(*ast.ConstDecl).Initializer.(*ast.BinaryExpr)
	if root.Operator != lexer.O {
		t.Fatalf("root operator = %v, want o", root.Operator)
	}
	conjunction, ok := root.Right.(*ast.BinaryExpr)
	if !ok || conjunction.Operator != lexer.Y {
		t.Fatalf("right operand = %#v, want y expression", root.Right)
	}
	equality, ok := conjunction.Right.(*ast.BinaryExpr)
	if !ok || equality.Operator != lexer.EqualEqual {
		t.Fatalf("right side of y = %#v, want equality", conjunction.Right)
	}
	negation, ok := equality.Left.(*ast.UnaryExpr)
	if !ok || negation.Operator != lexer.No {
		t.Fatalf("left side of equality = %#v, want no expression", equality.Left)
	}
	grouping, ok := negation.Right.(*ast.GroupingExpr)
	if !ok {
		t.Fatalf("negation operand = %#v, want grouped comparison", negation.Right)
	}
	comparison, ok := grouping.Expression.(*ast.BinaryExpr)
	if !ok || comparison.Operator != lexer.Less {
		t.Fatalf("grouped expression = %#v, want comparison", grouping.Expression)
	}
}

func TestParseElseIfChain(t *testing.T) {
	program := parseSource(t, `jugada classify(score) { si score > 0 { devuelve score; } sino si score == 0 { devuelve; } sino { devuelve nada; } }`)
	function := program.Declarations[0].(*ast.FunctionDecl)
	conditional := function.Body.Statements[0].(*ast.IfStmt)
	elseIf, ok := conditional.Else.(*ast.IfStmt)
	if !ok {
		t.Fatalf("else branch = %#v, want nested if", conditional.Else)
	}
	if elseIf.Else == nil {
		t.Fatal("nested if has no final else block")
	}
}

func TestParseRejectsDuplicateDictionaryKeys(t *testing.T) {
	source := `leyenda labels = {"name": 1, "name": 2};`
	position := strings.LastIndex(source, `"name"`)
	expectSyntaxErrorAt(t, source, lexer.Position{Offset: position, Line: 1, Column: position + 1})
}

func TestParseRejectsUnsupportedGrammarForms(t *testing.T) {
	cases := []struct {
		name   string
		source string
		marker string
	}{
		{name: "chained comparison", source: `leyenda x = 1 < 2 < 3;`, marker: "< 3"},
		{name: "list trailing comma", source: `leyenda x = [1,];`, marker: "]"},
		{name: "call trailing comma", source: `leyenda x = f(1,);`, marker: ")"},
		{name: "dictionary trailing comma", source: `leyenda x = {"a": 1,};`, marker: "}"},
		{name: "function parameter trailing comma", source: `jugada f(a,) { devuelve; }`, marker: ")"},
		{name: "non-string dictionary key", source: `leyenda x = {1: "bad"};`, marker: "1"},
		{name: "late import", source: `leyenda x = 1; fichar "lib.cr7" como lib;`, marker: "fichar"},
		{name: "block import", source: `jugada f() { fichar "lib.cr7" como lib; }`, marker: "fichar"},
		{name: "missing statement semicolon", source: `jugada f() { siuuu(1) }`, marker: "}"},
		{name: "top-level mutable binding", source: `cantera x = 1;`, marker: "cantera"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			position := strings.Index(test.source, test.marker)
			if position < 0 {
				t.Fatalf("test marker %q not found", test.marker)
			}
			expectSyntaxErrorAt(t, test.source, lexer.Position{Offset: position, Line: 1, Column: position + 1})
		})
	}
}

func TestParseFunctionAndControlFlowStatements(t *testing.T) {
	source := `jugada comandante(score) {
	cantera attempts = 0;
	score = score + 1;
	si score > 0 { siuuu(score); } sino { falta "missed"; }
	entrena score < 3 { siguiente; vestuario; }
	temporada rival en ["Porto"] { siuuu(rival); }
	remontada { falta "failure"; } var error { siuuu(error.codigo); }
	devuelve score;
}`
	program := parseSource(t, source)
	if len(program.Declarations) != 1 {
		t.Fatalf("declarations = %d, want one function", len(program.Declarations))
	}
	function, ok := program.Declarations[0].(*ast.FunctionDecl)
	if !ok || function.Name != "comandante" || len(function.Parameters) != 1 || function.Parameters[0] != "score" {
		t.Fatalf("function declaration = %#v, want comandante(score)", program.Declarations[0])
	}
	statements := function.Body.Statements
	if len(statements) != 7 {
		t.Fatalf("function statements = %d, want 7", len(statements))
	}
	if declaration, ok := statements[0].(*ast.MutableDecl); !ok || declaration.Name != "attempts" {
		t.Fatalf("first statement = %#v, want mutable attempts declaration", statements[0])
	}
	if assignment, ok := statements[1].(*ast.AssignmentStmt); !ok || assignment.Target.(*ast.IdentifierExpr).Name != "score" {
		t.Fatalf("second statement = %#v, want score assignment", statements[1])
	}
	if conditional, ok := statements[2].(*ast.IfStmt); !ok || conditional.Else == nil {
		t.Fatalf("third statement = %#v, want if/else", statements[2])
	}
	if loop, ok := statements[3].(*ast.WhileStmt); !ok || len(loop.Body.Statements) != 2 {
		t.Fatalf("fourth statement = %#v, want while with two control statements", statements[3])
	}
	if loop, ok := statements[4].(*ast.ForEachStmt); !ok || loop.Name != "rival" {
		t.Fatalf("fifth statement = %#v, want temporada rival loop", statements[4])
	}
	if recovery, ok := statements[5].(*ast.TryCatchStmt); !ok || recovery.ErrorName != "error" {
		t.Fatalf("sixth statement = %#v, want remontada handler", statements[5])
	}
	if result, ok := statements[6].(*ast.ReturnStmt); !ok || result.Value == nil {
		t.Fatalf("seventh statement = %#v, want value return", statements[6])
	}
}

func TestParseNestedFunctionAndExpressionStatements(t *testing.T) {
	program := parseSource(t, `jugada outer() { jugada inner(value) { devuelve; } inner(3); ({"name": "CR"}); }`)
	outer := program.Declarations[0].(*ast.FunctionDecl)
	if len(outer.Body.Statements) != 3 {
		t.Fatalf("outer statements = %d, want nested declaration and two expressions", len(outer.Body.Statements))
	}
	inner, ok := outer.Body.Statements[0].(*ast.FunctionDecl)
	if !ok || inner.Name != "inner" || len(inner.Parameters) != 1 {
		t.Fatalf("first outer statement = %#v, want inner function", outer.Body.Statements[0])
	}
	if result, ok := inner.Body.Statements[0].(*ast.ReturnStmt); !ok || result.Value != nil {
		t.Fatalf("inner statement = %#v, want return without value", inner.Body.Statements[0])
	}
	for i := 1; i < 3; i++ {
		if _, ok := outer.Body.Statements[i].(*ast.ExpressionStmt); !ok {
			t.Fatalf("outer statement %d = %#v, want expression statement", i, outer.Body.Statements[i])
		}
	}
}

func TestParseIndexAssignmentAndRejectsInvalidTargets(t *testing.T) {
	program := parseSource(t, `jugada update(values) { values[0] = 7; values = [7]; }`)
	function := program.Declarations[0].(*ast.FunctionDecl)
	if len(function.Body.Statements) != 2 {
		t.Fatalf("assignment statements = %d, want 2", len(function.Body.Statements))
	}
	indexed, ok := function.Body.Statements[0].(*ast.AssignmentStmt)
	if !ok {
		t.Fatalf("first statement = %#v, want assignment", function.Body.Statements[0])
	}
	if _, ok := indexed.Target.(*ast.IndexExpr); !ok {
		t.Fatalf("assignment target = %#v, want index expression", indexed.Target)
	}

	cases := []struct {
		name   string
		source string
	}{
		{name: "binary target", source: `jugada update(value) { (value + 1) = 2; }`},
		{name: "property target", source: `jugada update(value) { value.field = 2; }`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			position := strings.Index(test.source, "= 2")
			expectSyntaxErrorAt(t, test.source, lexer.Position{Offset: position, Line: 1, Column: position + 1})
		})
	}
}

func parseSource(t *testing.T, source string) *ast.Program {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan(%q) error = %v", source, err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse(%q) error = %v", source, err)
	}
	return program
}

func expectSyntaxErrorAt(t *testing.T, source string, want lexer.Position) {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan(%q) error = %v", source, err)
	}
	_, err = parser.Parse(tokens)
	if err == nil {
		t.Fatalf("parser.Parse(%q) succeeded, want syntax error", source)
	}
	var diagnostic *parser.Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *parser.Diagnostic", err)
	}
	if diagnostic.Code != "SintaxisInvalida" {
		t.Fatalf("diagnostic code = %q, want SintaxisInvalida", diagnostic.Code)
	}
	if diagnostic.Position != want {
		t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, want)
	}
}
