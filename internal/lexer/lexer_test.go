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
