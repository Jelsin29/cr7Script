package lexer

import (
	"errors"
	"reflect"
	"testing"
)

func TestScanKeywordsAndIdentifier(t *testing.T) {
	source := "cantera leyenda jugada devuelve si sino entrena temporada en vestuario siguiente remontada var falta fichar como y o no gol fuera nada golazo"
	got := scanTokens(t, source)
	want := []Kind{
		Cantera, Leyenda, Jugada, Devuelve, Si, Sino, Entrena, Temporada, En,
		Vestuario, Siguiente, Remontada, Var, Falta, Fichar, Como, Y, O, No,
		Gol, Fuera, Nada, Identifier, EOF,
	}
	if kinds := tokenKinds(got); !reflect.DeepEqual(kinds, want) {
		t.Fatalf("token kinds = %v, want %v", kinds, want)
	}
	if got[22].Lexeme != "golazo" {
		t.Fatalf("identifier lexeme = %q, want golazo", got[22].Lexeme)
	}
}

func TestScanOperatorsAndDelimiters(t *testing.T) {
	source := "= == != < <= > >= + - * / % ( ) { } [ ] , . : ;"
	got := scanTokens(t, source)
	want := []Kind{
		Equal, EqualEqual, BangEqual, Less, LessEqual, Greater, GreaterEqual,
		Plus, Minus, Star, Slash, Percent,
		LeftParen, RightParen, LeftBrace, RightBrace, LeftBracket, RightBracket,
		Comma, Dot, Colon, Semicolon, EOF,
	}
	if kinds := tokenKinds(got); !reflect.DeepEqual(kinds, want) {
		t.Fatalf("token kinds = %v, want %v", kinds, want)
	}
	if got[1].Lexeme != "==" || got[2].Lexeme != "!=" || got[6].Lexeme != ">=" {
		t.Fatalf("multi-character operators lost their lexemes: %#v", got[1:7])
	}
}

func TestScanTracksLineAndColumn(t *testing.T) {
	got := scanTokens(t, "cantera\nx = 1;")
	want := Position{Offset: 8, Line: 2, Column: 1}
	if got[1].Start != want {
		t.Fatalf("identifier start = %+v, want %+v", got[1].Start, want)
	}
	if got[len(got)-1].Kind != EOF {
		t.Fatalf("last token kind = %v, want EOF", got[len(got)-1].Kind)
	}
}

func TestScanRejectsUnknownCharacterAtItsLocation(t *testing.T) {
	_, err := Scan("@")
	if err == nil {
		t.Fatal("Scan(@) succeeded, want a diagnostic")
	}
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *Diagnostic", err)
	}
	if diagnostic.Code != "SintaxisInvalida" {
		t.Fatalf("diagnostic code = %q, want SintaxisInvalida", diagnostic.Code)
	}
	want := Position{Offset: 0, Line: 1, Column: 1}
	if diagnostic.Position != want {
		t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, want)
	}
}

func TestScanParsesFiniteNumericLiterals(t *testing.T) {
	got := scanTokens(t, "0 12.5 1e3 1.25E-2 7e+1")
	want := []float64{0, 12.5, 1000, 0.0125, 70}
	for i, expected := range want {
		value, ok := got[i].Literal.(float64)
		if !ok || value != expected {
			t.Fatalf("token %d literal = %#v, want %v", i, got[i].Literal, expected)
		}
	}
}

func TestScanRejectsNonFiniteNumericLiteral(t *testing.T) {
	_, err := Scan("gol\n    1e999")
	if err == nil {
		t.Fatal("Scan(1e999) succeeded, want NumeroNoFinito")
	}
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *Diagnostic", err)
	}
	if diagnostic.Code != "NumeroNoFinito" {
		t.Fatalf("diagnostic code = %q, want NumeroNoFinito", diagnostic.Code)
	}
	want := Position{Offset: 8, Line: 2, Column: 5}
	if diagnostic.Position != want {
		t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, want)
	}
}

func TestScanDecodesStringEscapesAndPreservesLexeme(t *testing.T) {
	source := `"a\"b\\c\n\r\t\u00F1"`
	got := scanTokens(t, source)
	if len(got) != 2 || got[0].Kind != StringLiteral {
		t.Fatalf("tokens = %#v, want one string and EOF", got)
	}
	if got[0].Lexeme != source {
		t.Fatalf("lexeme = %q, want original source %q", got[0].Lexeme, source)
	}
	if value, ok := got[0].Literal.(string); !ok || value != "a\"b\\c\n\r\tñ" {
		t.Fatalf("string literal = %#v, want decoded text", got[0].Literal)
	}
}

func TestScanRejectsInvalidStringForms(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		position Position
	}{
		{name: "unknown escape", source: `"bad\q"`, position: Position{Offset: 4, Line: 1, Column: 5}},
		{name: "surrogate escape", source: `"\uD800"`, position: Position{Offset: 1, Line: 1, Column: 2}},
		{name: "raw line break", source: "\"first\nsecond\"", position: Position{Offset: 6, Line: 1, Column: 7}},
		{name: "unterminated", source: `"open`, position: Position{Offset: 0, Line: 1, Column: 1}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Scan(test.source)
			if err == nil {
				t.Fatalf("Scan(%q) succeeded, want a lexical diagnostic", test.source)
			}
			var diagnostic *Diagnostic
			if !errors.As(err, &diagnostic) || diagnostic.Code != "SintaxisInvalida" {
				t.Fatalf("error = %#v, want SintaxisInvalida diagnostic", err)
			}
			if diagnostic.Position != test.position {
				t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, test.position)
			}
		})
	}
}

func TestScanSkipsUnicodeCommentsAndKeepsNonNesting(t *testing.T) {
	got := scanTokens(t, "cantera // ¡Vamos!\n leyenda /* nota ñ */ gol")
	want := []Kind{Cantera, Leyenda, Gol, EOF}
	if kinds := tokenKinds(got); !reflect.DeepEqual(kinds, want) {
		t.Fatalf("comment token kinds = %v, want %v", kinds, want)
	}

	nested := scanTokens(t, "cantera /* outer /* inner */ leyenda */ gol")
	want = []Kind{Cantera, Leyenda, Star, Slash, Gol, EOF}
	if kinds := tokenKinds(nested); !reflect.DeepEqual(kinds, want) {
		t.Fatalf("non-nesting comment token kinds = %v, want %v", kinds, want)
	}
}

func TestScanReportsUnterminatedBlockComment(t *testing.T) {
	_, err := Scan("cantera /* open")
	if err == nil {
		t.Fatal("unterminated block comment succeeded, want a lexical diagnostic")
	}
	var diagnostic *Diagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "SintaxisInvalida" {
		t.Fatalf("error = %#v, want SintaxisInvalida diagnostic", err)
	}
	want := Position{Offset: 8, Line: 1, Column: 9}
	if diagnostic.Position != want {
		t.Fatalf("diagnostic position = %+v, want %+v", diagnostic.Position, want)
	}
}

func TestScanCountsUnicodeScalarsInColumns(t *testing.T) {
	got := scanTokens(t, `"ñ" x`)
	want := Position{Offset: 5, Line: 1, Column: 5}
	if got[1].Start != want {
		t.Fatalf("identifier start = %+v, want %+v", got[1].Start, want)
	}
}

func scanTokens(t *testing.T, source string) []Token {
	t.Helper()
	tokens, err := Scan(source)
	if err != nil {
		t.Fatalf("Scan(%q) error = %v", source, err)
	}
	return tokens
}

func tokenKinds(tokens []Token) []Kind {
	kinds := make([]Kind, len(tokens))
	for i, token := range tokens {
		kinds[i] = token.Kind
	}
	return kinds
}
