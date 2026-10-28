package lexer

import "unicode/utf8"

const syntaxErrorCode = "SintaxisInvalida"

var keywords = map[string]Kind{
	"cantera":   Cantera,
	"leyenda":   Leyenda,
	"jugada":    Jugada,
	"devuelve":  Devuelve,
	"si":        Si,
	"sino":      Sino,
	"entrena":   Entrena,
	"temporada": Temporada,
	"en":        En,
	"vestuario": Vestuario,
	"siguiente": Siguiente,
	"remontada": Remontada,
	"var":       Var,
	"falta":     Falta,
	"fichar":    Fichar,
	"como":      Como,
	"y":         Y,
	"o":         O,
	"no":        No,
	"gol":       Gol,
	"fuera":     Fuera,
	"nada":      Nada,
}

// Scan tokenizes source and appends one EOF token on success.
func Scan(source string) ([]Token, error) {
	s := scanner{source: source, line: 1, column: 1}
	return s.scanTokens()
}

type scanner struct {
	source  string
	start   int
	current int
	line    int
	column  int
	startAt Position
	tokens  []Token
}

func (s *scanner) scanTokens() ([]Token, error) {
	for !s.atEnd() {
		s.start = s.current
		s.startAt = s.position()
		if err := s.scanToken(); err != nil {
			return nil, err
		}
	}
	end := s.position()
	s.tokens = append(s.tokens, Token{Kind: EOF, Start: end, End: end})
	return s.tokens, nil
}

func (s *scanner) scanToken() error {
	character := s.advance()
	switch character {
	case ' ', '\r', '\t', '\n':
		return nil
	case '(':
		s.addToken(LeftParen, nil)
	case ')':
		s.addToken(RightParen, nil)
	case '{':
		s.addToken(LeftBrace, nil)
	case '}':
		s.addToken(RightBrace, nil)
	case '[':
		s.addToken(LeftBracket, nil)
	case ']':
		s.addToken(RightBracket, nil)
	case ',':
		s.addToken(Comma, nil)
	case '.':
		s.addToken(Dot, nil)
	case ':':
		s.addToken(Colon, nil)
	case ';':
		s.addToken(Semicolon, nil)
	case '+':
		s.addToken(Plus, nil)
	case '-':
		s.addToken(Minus, nil)
	case '*':
		s.addToken(Star, nil)
	case '/':
		s.addToken(Slash, nil)
	case '%':
		s.addToken(Percent, nil)
	case '=':
		if s.match('=') {
			s.addToken(EqualEqual, nil)
		} else {
			s.addToken(Equal, nil)
		}
	case '!':
		if !s.match('=') {
			return s.diagnostic("unexpected '!'", s.startAt)
		}
		s.addToken(BangEqual, nil)
	case '<':
		if s.match('=') {
			s.addToken(LessEqual, nil)
		} else {
			s.addToken(Less, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(GreaterEqual, nil)
		} else {
			s.addToken(Greater, nil)
		}
	default:
		switch {
		case isASCIILetter(character) || character == '_':
			s.identifier()
		case isASCIIDigit(character):
			s.number()
		default:
			return s.diagnostic("unexpected character", s.startAt)
		}
	}
	return nil
}

func (s *scanner) identifier() {
	for isASCIILetter(s.peek()) || isASCIIDigit(s.peek()) || s.peek() == '_' {
		s.advance()
	}
	lexeme := s.source[s.start:s.current]
	kind, ok := keywords[lexeme]
	if !ok {
		kind = Identifier
	}
	s.addToken(kind, nil)
}

func (s *scanner) number() {
	for isASCIIDigit(s.peek()) {
		s.advance()
	}
	if s.peek() == '.' && isASCIIDigit(s.peekNext()) {
		s.advance()
		for isASCIIDigit(s.peek()) {
			s.advance()
		}
	}
	if (s.peek() == 'e' || s.peek() == 'E') && s.validExponent() {
		s.advance()
		if s.peek() == '+' || s.peek() == '-' {
			s.advance()
		}
		for isASCIIDigit(s.peek()) {
			s.advance()
		}
	}
	s.addToken(Number, nil)
}

func (s *scanner) validExponent() bool {
	index := s.current + 1
	if index < len(s.source) && (s.source[index] == '+' || s.source[index] == '-') {
		index++
	}
	return index < len(s.source) && s.source[index] >= '0' && s.source[index] <= '9'
}

func (s *scanner) addToken(kind Kind, literal any) {
	s.tokens = append(s.tokens, Token{
		Kind:    kind,
		Lexeme:  s.source[s.start:s.current],
		Literal: literal,
		Start:   s.startAt,
		End:     s.position(),
	})
}

func (s *scanner) diagnostic(message string, position Position) error {
	return &Diagnostic{Code: syntaxErrorCode, Message: message, Position: position}
}

func (s *scanner) advance() rune {
	character, size := utf8.DecodeRuneInString(s.source[s.current:])
	s.current += size
	if character == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	return character
}

func (s *scanner) match(expected byte) bool {
	if s.atEnd() || s.source[s.current] != expected {
		return false
	}
	s.current++
	s.column++
	return true
}

func (s *scanner) peek() rune {
	if s.atEnd() {
		return 0
	}
	character, _ := utf8.DecodeRuneInString(s.source[s.current:])
	return character
}

func (s *scanner) peekNext() rune {
	if s.atEnd() {
		return 0
	}
	_, size := utf8.DecodeRuneInString(s.source[s.current:])
	index := s.current + size
	if index >= len(s.source) {
		return 0
	}
	character, _ := utf8.DecodeRuneInString(s.source[index:])
	return character
}

func (s *scanner) atEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) position() Position {
	return Position{Offset: s.current, Line: s.line, Column: s.column}
}

func isASCIILetter(character rune) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
}

func isASCIIDigit(character rune) bool {
	return character >= '0' && character <= '9'
}
