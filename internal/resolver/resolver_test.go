package resolver_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/parser"
	"github.com/Jelsin29/cr7Script/internal/resolver"
)

func TestResolveRejectsUnknownNames(t *testing.T) {
	source := `jugada score() { devuelve missing; }`
	expectCodeAt(t, source, "NombreNoDefinido", strings.Index(source, "missing"))
}

func TestResolveRejectsReadsBeforeInitialization(t *testing.T) {
	t.Run("later module constant", func(t *testing.T) {
		source := `leyenda first = later; leyenda later = 2;`
		expectCodeAt(t, source, "NombreSinInicializar", strings.Index(source, "later"))
	})
	t.Run("inner declaration shadows outer binding before initialization", func(t *testing.T) {
		source := `jugada read() { cantera score = 1; { siuuu(score); cantera score = 2; } }`
		position := strings.Index(source, "siuuu(score)") + len("siuuu(")
		expectCodeAt(t, source, "NombreSinInicializar", position)
	})
	t.Run("self reference in initializer", func(t *testing.T) {
		source := `jugada read() { cantera score = score; }`
		position := strings.LastIndex(source, "score")
		expectCodeAt(t, source, "NombreSinInicializar", position)
	})
}

func TestResolveRejectsDuplicateDeclarations(t *testing.T) {
	t.Run("same-scope declarations", func(t *testing.T) {
		source := `jugada update() { cantera score = 1; leyenda score = 2; }`
		expectCodeAt(t, source, "DeclaracionDuplicada", strings.Index(source, "leyenda score"))
	})
	t.Run("duplicate parameters", func(t *testing.T) {
		source := `jugada update(score, score) { devuelve; }`
		expectCodeAt(t, source, "DeclaracionDuplicada", strings.Index(source, "jugada"))
	})
}

func TestResolveRejectsReassignmentToImmutableBindings(t *testing.T) {
	t.Run("constant", func(t *testing.T) {
		source := `jugada update() { leyenda score = 1; score = 2; }`
		expectCodeAt(t, source, "ReasignacionDeLeyenda", strings.Index(source, "score = 2"))
	})
	t.Run("import alias", func(t *testing.T) {
		source := `fichar "mr_champions" como champions; jugada update() { champions = nada; }`
		expectCodeAt(t, source, "ReasignacionDeLeyenda", strings.Index(source, "champions = nada"))
	})
}

func TestResolveAllowsFunctionReassignmentAndForwardCalls(t *testing.T) {
	source := `jugada caller(value) { value = 7; first = second; } jugada first() { second(); devuelve; } jugada second() { devuelve; }`
	if err := resolveSource(t, source); err != nil {
		t.Fatalf("Resolve(%q) error = %v, want nil", source, err)
	}
}

func TestResolveRejectsInvalidLoopControl(t *testing.T) {
	for _, keyword := range []string{"vestuario", "siguiente"} {
		t.Run(keyword, func(t *testing.T) {
			source := "jugada f() { " + keyword + "; }"
			expectCodeAt(t, source, "ControlFueraDeBucle", strings.Index(source, keyword))
		})
	}
}

func TestResolveRejectsReturnOutsideFunction(t *testing.T) {
	position := lexer.Position{Offset: 3, Line: 1, Column: 4}
	program := &ast.Program{Declarations: []ast.Stmt{
		&ast.ReturnStmt{Node: ast.Node{Start: position, End: lexer.Position{Offset: 12, Line: 1, Column: 13}}},
	}}
	err := resolver.Resolve(program)
	expectDiagnostic(t, err, "DevolucionFueraDeFuncion", position)
}

func TestResolveAcceptsLoopControlOnlyWithinTheSameFunction(t *testing.T) {
	valid := `jugada f() { entrena gol { vestuario; siguiente; } }`
	if err := resolveSource(t, valid); err != nil {
		t.Fatalf("Resolve(%q) error = %v, want nil", valid, err)
	}

	nested := `jugada f() { entrena gol { jugada nested() { vestuario; } } }`
	position := strings.Index(nested, "vestuario")
	expectCodeAt(t, nested, "ControlFueraDeBucle", position)
}

func TestResolveRejectsPreludeRedeclarationAndOutOfScopeLoopBinding(t *testing.T) {
	t.Run("prelude name", func(t *testing.T) {
		source := `jugada texto(value) { devuelve value; }`
		expectCodeAt(t, source, "DeclaracionDuplicada", strings.Index(source, "jugada"))
	})
	t.Run("loop binding outside its body", func(t *testing.T) {
		source := `jugada f() { temporada rival en [] { siuuu(rival); } siuuu(rival); }`
		position := strings.LastIndex(source, "rival")
		expectCodeAt(t, source, "NombreNoDefinido", position)
	})
}

func resolveSource(t *testing.T, source string) error {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan(%q) error = %v", source, err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse(%q) error = %v", source, err)
	}
	return resolver.Resolve(program)
}

func expectCodeAt(t *testing.T, source, code string, offset int) {
	t.Helper()
	if offset < 0 {
		t.Fatalf("expected source marker for %q", code)
	}
	line, column := 1, 1
	for _, r := range source[:offset] {
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	want := lexer.Position{Offset: offset, Line: line, Column: column}
	err := resolveSource(t, source)
	expectDiagnostic(t, err, code, want)
}

func expectDiagnostic(t *testing.T, err error, code string, position lexer.Position) {
	t.Helper()
	if err == nil {
		t.Fatalf("Resolve() succeeded, want %s", code)
	}
	var diagnostic *resolver.Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *resolver.Diagnostic", err)
	}
	if diagnostic.Code != code {
		t.Fatalf("diagnostic code = %q, want %q", diagnostic.Code, code)
	}
	if diagnostic.Position != position {
		t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, position)
	}
}
