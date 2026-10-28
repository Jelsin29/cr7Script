package lexer

import "strconv"

// Kind identifies one lexical token.
type Kind string

const (
	EOF Kind = "EOF"

	Identifier    Kind = "IDENTIFIER"
	Number        Kind = "NUMBER"
	StringLiteral Kind = "STRING"

	LeftParen    Kind = "LEFT_PAREN"
	RightParen   Kind = "RIGHT_PAREN"
	LeftBrace    Kind = "LEFT_BRACE"
	RightBrace   Kind = "RIGHT_BRACE"
	LeftBracket  Kind = "LEFT_BRACKET"
	RightBracket Kind = "RIGHT_BRACKET"
	Comma        Kind = "COMMA"
	Dot          Kind = "DOT"
	Colon        Kind = "COLON"
	Semicolon    Kind = "SEMICOLON"

	Equal        Kind = "EQUAL"
	EqualEqual   Kind = "EQUAL_EQUAL"
	BangEqual    Kind = "BANG_EQUAL"
	Less         Kind = "LESS"
	LessEqual    Kind = "LESS_EQUAL"
	Greater      Kind = "GREATER"
	GreaterEqual Kind = "GREATER_EQUAL"
	Plus         Kind = "PLUS"
	Minus        Kind = "MINUS"
	Star         Kind = "STAR"
	Slash        Kind = "SLASH"
	Percent      Kind = "PERCENT"

	Cantera   Kind = "CANTERA"
	Leyenda   Kind = "LEYENDA"
	Jugada    Kind = "JUGADA"
	Devuelve  Kind = "DEVUELVE"
	Si        Kind = "SI"
	Sino      Kind = "SINO"
	Entrena   Kind = "ENTRENA"
	Temporada Kind = "TEMPORADA"
	En        Kind = "EN"
	Vestuario Kind = "VESTUARIO"
	Siguiente Kind = "SIGUIENTE"
	Remontada Kind = "REMONTADA"
	Var       Kind = "VAR"
	Falta     Kind = "FALTA"
	Fichar    Kind = "FICHAR"
	Como      Kind = "COMO"
	Y         Kind = "Y"
	O         Kind = "O"
	No        Kind = "NO"
	Gol       Kind = "GOL"
	Fuera     Kind = "FUERA"
	Nada      Kind = "NADA"
)

// Position uses a byte offset and one-based Unicode-scalar line and column.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Token retains its original lexeme and source range for later diagnostics.
type Token struct {
	Kind    Kind
	Lexeme  string
	Literal any
	Start   Position
	End     Position
}

// Diagnostic is a source-located lexical error.
type Diagnostic struct {
	Code     string
	Message  string
	Position Position
}

func (d *Diagnostic) Error() string {
	return d.Code + " at " + strconv.Itoa(d.Position.Line) + ":" + strconv.Itoa(d.Position.Column) + ": " + d.Message
}
