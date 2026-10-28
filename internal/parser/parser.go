package parser

import (
	"fmt"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/lexer"
)

// Diagnostic is a source-located parser error.
type Diagnostic struct {
	Code     string
	Message  string
	Position lexer.Position
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s at %d:%d: %s", d.Code, d.Position.Line, d.Position.Column, d.Message)
}

// Parse builds an AST for a CR7Script module.
func Parse(tokens []lexer.Token) (*ast.Program, error) {
	if len(tokens) == 0 {
		return nil, &Diagnostic{Code: "SintaxisInvalida", Message: "token stream is empty"}
	}
	if tokens[len(tokens)-1].Kind != lexer.EOF {
		return nil, &Diagnostic{Code: "SintaxisInvalida", Message: "token stream must end with EOF", Position: tokens[len(tokens)-1].End}
	}
	p := parserState{tokens: tokens}
	return p.parseProgram()
}

type parserState struct {
	tokens  []lexer.Token
	current int
}

func (p *parserState) parseProgram() (*ast.Program, error) {
	start := p.peek().Start
	program := &ast.Program{}
	for p.match(lexer.Fichar) {
		declaration, err := p.importDeclaration(p.previous())
		if err != nil {
			return nil, err
		}
		program.Imports = append(program.Imports, declaration)
	}
	for !p.atEnd() {
		declaration, err := p.moduleDeclaration()
		if err != nil {
			return nil, err
		}
		program.Declarations = append(program.Declarations, declaration)
	}
	program.Start = start
	program.End = p.peek().End
	return program, nil
}

func (p *parserState) importDeclaration(keyword lexer.Token) (ast.ImportDecl, error) {
	path, err := p.consume(lexer.StringLiteral, "expected a string module path after 'fichar'")
	if err != nil {
		return ast.ImportDecl{}, err
	}
	if _, err = p.consume(lexer.Como, "expected 'como' after the module path"); err != nil {
		return ast.ImportDecl{}, err
	}
	alias, err := p.consume(lexer.Identifier, "expected an import alias after 'como'")
	if err != nil {
		return ast.ImportDecl{}, err
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the import declaration")
	if err != nil {
		return ast.ImportDecl{}, err
	}
	value, ok := path.Literal.(string)
	if !ok {
		return ast.ImportDecl{}, p.errorAt(path, "module path token has no string value")
	}
	return ast.ImportDecl{
		Node:  ast.Node{Start: keyword.Start, End: semicolon.End},
		Path:  value,
		Alias: alias.Lexeme,
	}, nil
}

func (p *parserState) moduleDeclaration() (ast.Stmt, error) {
	if p.match(lexer.Leyenda) {
		return p.constantDeclaration(p.previous())
	}
	if p.match(lexer.Jugada) {
		return p.functionDeclaration(p.previous())
	}
	return nil, p.errorAt(p.peek(), "expected a module constant or function declaration")
}

func (p *parserState) declaration() (ast.Stmt, error) {
	switch {
	case p.match(lexer.Cantera):
		return p.mutableDeclaration(p.previous())
	case p.match(lexer.Leyenda):
		return p.constantDeclaration(p.previous())
	case p.match(lexer.Jugada):
		return p.functionDeclaration(p.previous())
	default:
		return p.statement()
	}
}

func (p *parserState) mutableDeclaration(keyword lexer.Token) (ast.Stmt, error) {
	name, err := p.consume(lexer.Identifier, "expected a name after 'cantera'")
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(lexer.Equal, "expected '=' after the mutable binding name"); err != nil {
		return nil, err
	}
	initializer, err := p.expression()
	if err != nil {
		return nil, err
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the binding initializer")
	if err != nil {
		return nil, err
	}
	return &ast.MutableDecl{
		Node:        ast.Node{Start: keyword.Start, End: semicolon.End},
		Name:        name.Lexeme,
		Initializer: initializer,
	}, nil
}

func (p *parserState) functionDeclaration(keyword lexer.Token) (ast.Stmt, error) {
	name, err := p.consume(lexer.Identifier, "expected a function name after 'jugada'")
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(lexer.LeftParen, "expected '(' after the function name"); err != nil {
		return nil, err
	}
	parameters := make([]string, 0)
	if !p.check(lexer.RightParen) {
		for {
			parameter, parseErr := p.consume(lexer.Identifier, "expected a parameter name")
			if parseErr != nil {
				return nil, parseErr
			}
			parameters = append(parameters, parameter.Lexeme)
			if !p.match(lexer.Comma) {
				break
			}
		}
	}
	if _, err = p.consume(lexer.RightParen, "expected ')' after the parameters"); err != nil {
		return nil, err
	}
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a function body")
	}
	body, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	return &ast.FunctionDecl{
		Node:       ast.Node{Start: keyword.Start, End: body.End},
		Name:       name.Lexeme,
		Parameters: parameters,
		Body:       body,
	}, nil
}

func (p *parserState) statement() (ast.Stmt, error) {
	if p.match(lexer.LeftBrace) {
		return p.blockStatement(p.previous())
	}
	if p.match(lexer.Si) {
		return p.ifStatement(p.previous())
	}
	if p.match(lexer.Entrena) {
		return p.whileStatement(p.previous())
	}
	if p.match(lexer.Temporada) {
		return p.forEachStatement(p.previous())
	}
	if p.match(lexer.Remontada) {
		return p.tryCatchStatement(p.previous())
	}
	if p.match(lexer.Devuelve) {
		return p.returnStatement(p.previous())
	}
	if p.match(lexer.Vestuario) {
		return p.controlStatement(p.previous(), true)
	}
	if p.match(lexer.Siguiente) {
		return p.controlStatement(p.previous(), false)
	}
	if p.match(lexer.Falta) {
		return p.throwStatement(p.previous())
	}
	return p.simpleStatement()
}

func (p *parserState) blockStatement(open lexer.Token) (*ast.BlockStmt, error) {
	statements := make([]ast.Stmt, 0)
	for !p.check(lexer.RightBrace) && !p.atEnd() {
		statement, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	close, err := p.consume(lexer.RightBrace, "expected '}' after the block")
	if err != nil {
		return nil, err
	}
	return &ast.BlockStmt{Node: ast.Node{Start: open.Start, End: close.End}, Statements: statements}, nil
}

func (p *parserState) ifStatement(keyword lexer.Token) (ast.Stmt, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a block after the condition")
	}
	thenBlock, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	var elseBranch ast.Stmt
	end := thenBlock.End
	if p.match(lexer.Sino) {
		if p.match(lexer.Si) {
			elseBranch, err = p.ifStatement(p.previous())
		} else if p.match(lexer.LeftBrace) {
			elseBranch, err = p.blockStatement(p.previous())
		} else {
			err = p.errorAt(p.peek(), "expected 'si' or a block after 'sino'")
		}
		if err != nil {
			return nil, err
		}
		end = nodeOf(elseBranch).End
	}
	return &ast.IfStmt{Node: ast.Node{Start: keyword.Start, End: end}, Condition: condition, Then: thenBlock, Else: elseBranch}, nil
}

func (p *parserState) whileStatement(keyword lexer.Token) (ast.Stmt, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a block after the loop condition")
	}
	body, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Node: ast.Node{Start: keyword.Start, End: body.End}, Condition: condition, Body: body}, nil
}

func (p *parserState) forEachStatement(keyword lexer.Token) (ast.Stmt, error) {
	name, err := p.consume(lexer.Identifier, "expected a loop variable after 'temporada'")
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(lexer.En, "expected 'en' after the loop variable"); err != nil {
		return nil, err
	}
	iterable, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a block after the iterable")
	}
	body, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	return &ast.ForEachStmt{Node: ast.Node{Start: keyword.Start, End: body.End}, Name: name.Lexeme, Iterable: iterable, Body: body}, nil
}

func (p *parserState) tryCatchStatement(keyword lexer.Token) (ast.Stmt, error) {
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a protected block after 'remontada'")
	}
	protected, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(lexer.Var, "expected 'var' after the protected block"); err != nil {
		return nil, err
	}
	name, err := p.consume(lexer.Identifier, "expected an error binding after 'var'")
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.LeftBrace) {
		return nil, p.errorAt(p.peek(), "expected a handler block")
	}
	handler, err := p.blockStatement(p.previous())
	if err != nil {
		return nil, err
	}
	return &ast.TryCatchStmt{Node: ast.Node{Start: keyword.Start, End: handler.End}, Protected: protected, ErrorName: name.Lexeme, Handler: handler}, nil
}

func (p *parserState) returnStatement(keyword lexer.Token) (ast.Stmt, error) {
	var value ast.Expr
	var err error
	if !p.check(lexer.Semicolon) {
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the return statement")
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStmt{Node: ast.Node{Start: keyword.Start, End: semicolon.End}, Value: value}, nil
}

func (p *parserState) controlStatement(keyword lexer.Token, isBreak bool) (ast.Stmt, error) {
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the control statement")
	if err != nil {
		return nil, err
	}
	span := ast.Node{Start: keyword.Start, End: semicolon.End}
	if isBreak {
		return &ast.BreakStmt{Node: span}, nil
	}
	return &ast.ContinueStmt{Node: span}, nil
}

func (p *parserState) throwStatement(keyword lexer.Token) (ast.Stmt, error) {
	value, err := p.expression()
	if err != nil {
		return nil, err
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the failure value")
	if err != nil {
		return nil, err
	}
	return &ast.ThrowStmt{Node: ast.Node{Start: keyword.Start, End: semicolon.End}, Value: value}, nil
}

func (p *parserState) simpleStatement() (ast.Stmt, error) {
	expression, err := p.expression()
	if err != nil {
		return nil, err
	}
	if p.match(lexer.Equal) {
		assignment := p.previous()
		if !isAssignmentTarget(expression) {
			return nil, p.errorAt(assignment, "invalid assignment target")
		}
		value, parseErr := p.expression()
		if parseErr != nil {
			return nil, parseErr
		}
		semicolon, parseErr := p.consume(lexer.Semicolon, "expected ';' after the assignment")
		if parseErr != nil {
			return nil, parseErr
		}
		return &ast.AssignmentStmt{Node: ast.Node{Start: nodeOf(expression).Start, End: semicolon.End}, Target: expression, Value: value}, nil
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the expression")
	if err != nil {
		return nil, err
	}
	return &ast.ExpressionStmt{Node: ast.Node{Start: nodeOf(expression).Start, End: semicolon.End}, Expression: expression}, nil
}

func isAssignmentTarget(expression ast.Expr) bool {
	switch expression.(type) {
	case *ast.IdentifierExpr, *ast.IndexExpr:
		return true
	default:
		return false
	}
}

func (p *parserState) constantDeclaration(keyword lexer.Token) (ast.Stmt, error) {
	name, err := p.consume(lexer.Identifier, "expected a name after 'leyenda'")
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(lexer.Equal, "expected '=' after the constant name"); err != nil {
		return nil, err
	}
	initializer, err := p.expression()
	if err != nil {
		return nil, err
	}
	semicolon, err := p.consume(lexer.Semicolon, "expected ';' after the constant initializer")
	if err != nil {
		return nil, err
	}
	return &ast.ConstDecl{
		Node:        ast.Node{Start: keyword.Start, End: semicolon.End},
		Name:        name.Lexeme,
		Initializer: initializer,
	}, nil
}

func (p *parserState) expression() (ast.Expr, error) {
	return p.logicalOr()
}

func (p *parserState) logicalOr() (ast.Expr, error) {
	expression, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.O) {
		operator := p.previous()
		right, parseErr := p.logicalAnd()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
	}
	return expression, nil
}

func (p *parserState) logicalAnd() (ast.Expr, error) {
	expression, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.Y) {
		operator := p.previous()
		right, parseErr := p.equality()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
	}
	return expression, nil
}

func (p *parserState) equality() (ast.Expr, error) {
	expression, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.EqualEqual, lexer.BangEqual) {
		operator := p.previous()
		right, parseErr := p.comparison()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
	}
	return expression, nil
}

func (p *parserState) comparison() (ast.Expr, error) {
	expression, err := p.term()
	if err != nil {
		return nil, err
	}
	if p.match(lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual) {
		operator := p.previous()
		right, parseErr := p.term()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
		if p.check(lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual) {
			return nil, p.errorAt(p.peek(), "chained comparisons are not allowed")
		}
	}
	return expression, nil
}

func (p *parserState) term() (ast.Expr, error) {
	expression, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.Plus, lexer.Minus) {
		operator := p.previous()
		right, parseErr := p.factor()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
	}
	return expression, nil
}

func (p *parserState) factor() (ast.Expr, error) {
	expression, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.Star, lexer.Slash, lexer.Percent) {
		operator := p.previous()
		right, parseErr := p.unary()
		if parseErr != nil {
			return nil, parseErr
		}
		expression = binary(expression, operator, right)
	}
	return expression, nil
}

func (p *parserState) unary() (ast.Expr, error) {
	if p.match(lexer.No, lexer.Minus) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Node: ast.Node{Start: operator.Start, End: nodeOf(right).End}, Operator: operator.Kind, Right: right}, nil
	}
	return p.postfix()
}

func (p *parserState) postfix() (ast.Expr, error) {
	expression, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.match(lexer.LeftParen):
			arguments, close, parseErr := p.arguments(lexer.RightParen)
			if parseErr != nil {
				return nil, parseErr
			}
			expression = &ast.CallExpr{Node: ast.Node{Start: nodeOf(expression).Start, End: close.End}, Callee: expression, Arguments: arguments}
		case p.match(lexer.LeftBracket):
			index, parseErr := p.expression()
			if parseErr != nil {
				return nil, parseErr
			}
			close, parseErr := p.consume(lexer.RightBracket, "expected ']' after the index")
			if parseErr != nil {
				return nil, parseErr
			}
			expression = &ast.IndexExpr{Node: ast.Node{Start: nodeOf(expression).Start, End: close.End}, Collection: expression, Index: index}
		case p.match(lexer.Dot):
			name, parseErr := p.consume(lexer.Identifier, "expected a property name after '.'")
			if parseErr != nil {
				return nil, parseErr
			}
			expression = &ast.PropertyExpr{Node: ast.Node{Start: nodeOf(expression).Start, End: name.End}, Object: expression, Name: name.Lexeme}
		default:
			return expression, nil
		}
	}
}

func (p *parserState) primary() (ast.Expr, error) {
	if p.match(lexer.Number) {
		token := p.previous()
		value, ok := token.Literal.(float64)
		if !ok {
			return nil, p.errorAt(token, "number token has no numeric value")
		}
		return &ast.NumberExpr{Node: ast.Node{Start: token.Start, End: token.End}, Value: value}, nil
	}
	if p.match(lexer.StringLiteral) {
		token := p.previous()
		value, ok := token.Literal.(string)
		if !ok {
			return nil, p.errorAt(token, "string token has no text value")
		}
		return &ast.StringExpr{Node: ast.Node{Start: token.Start, End: token.End}, Value: value}, nil
	}
	if p.match(lexer.Gol) {
		token := p.previous()
		return &ast.BooleanExpr{Node: ast.Node{Start: token.Start, End: token.End}, Value: true}, nil
	}
	if p.match(lexer.Fuera) {
		token := p.previous()
		return &ast.BooleanExpr{Node: ast.Node{Start: token.Start, End: token.End}, Value: false}, nil
	}
	if p.match(lexer.Nada) {
		token := p.previous()
		return &ast.NullExpr{Node: ast.Node{Start: token.Start, End: token.End}}, nil
	}
	if p.match(lexer.Identifier) {
		token := p.previous()
		return &ast.IdentifierExpr{Node: ast.Node{Start: token.Start, End: token.End}, Name: token.Lexeme}, nil
	}
	if p.match(lexer.LeftParen) {
		open := p.previous()
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		close, err := p.consume(lexer.RightParen, "expected ')' after the expression")
		if err != nil {
			return nil, err
		}
		return &ast.GroupingExpr{Node: ast.Node{Start: open.Start, End: close.End}, Expression: expression}, nil
	}
	if p.match(lexer.LeftBracket) {
		return p.listLiteral(p.previous())
	}
	if p.match(lexer.LeftBrace) {
		return p.dictionaryLiteral(p.previous())
	}
	return nil, p.errorAt(p.peek(), "expected an expression")
}

func (p *parserState) listLiteral(open lexer.Token) (ast.Expr, error) {
	elements, close, err := p.arguments(lexer.RightBracket)
	if err != nil {
		return nil, err
	}
	return &ast.ListExpr{Node: ast.Node{Start: open.Start, End: close.End}, Elements: elements}, nil
}

func (p *parserState) dictionaryLiteral(open lexer.Token) (ast.Expr, error) {
	entries := make([]ast.DictEntry, 0)
	keys := make(map[string]struct{})
	if !p.check(lexer.RightBrace) {
		for {
			key, err := p.consume(lexer.StringLiteral, "expected a string key in the dictionary")
			if err != nil {
				return nil, err
			}
			if _, err = p.consume(lexer.Colon, "expected ':' after the dictionary key"); err != nil {
				return nil, err
			}
			value, parseErr := p.expression()
			if parseErr != nil {
				return nil, parseErr
			}
			keyValue, ok := key.Literal.(string)
			if !ok {
				return nil, p.errorAt(key, "dictionary key token has no text value")
			}
			if _, exists := keys[keyValue]; exists {
				return nil, p.errorAt(key, "duplicate dictionary key")
			}
			keys[keyValue] = struct{}{}
			entries = append(entries, ast.DictEntry{Node: ast.Node{Start: key.Start, End: nodeOf(value).End}, Key: keyValue, Value: value})
			if !p.match(lexer.Comma) {
				break
			}
		}
	}
	close, err := p.consume(lexer.RightBrace, "expected '}' after the dictionary entries")
	if err != nil {
		return nil, err
	}
	return &ast.DictExpr{Node: ast.Node{Start: open.Start, End: close.End}, Entries: entries}, nil
}

func (p *parserState) arguments(closeKind lexer.Kind) ([]ast.Expr, lexer.Token, error) {
	arguments := make([]ast.Expr, 0)
	if !p.check(closeKind) {
		for {
			expression, err := p.expression()
			if err != nil {
				return nil, lexer.Token{}, err
			}
			arguments = append(arguments, expression)
			if !p.match(lexer.Comma) {
				break
			}
		}
	}
	close, err := p.consume(closeKind, "expected the closing delimiter")
	return arguments, close, err
}

func binary(left ast.Expr, operator lexer.Token, right ast.Expr) ast.Expr {
	return &ast.BinaryExpr{
		Node:     ast.Node{Start: nodeOf(left).Start, End: nodeOf(right).End},
		Left:     left,
		Operator: operator.Kind,
		Right:    right,
	}
}

func nodeOf(node interface{ GetNode() ast.Node }) ast.Node {
	return node.GetNode()
}

func (p *parserState) match(kinds ...lexer.Kind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parserState) consume(kind lexer.Kind, message string) (lexer.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	return lexer.Token{}, p.errorAt(p.peek(), message)
}

func (p *parserState) check(kinds ...lexer.Kind) bool {
	for _, kind := range kinds {
		if p.peek().Kind == kind {
			return true
		}
	}
	return false
}

func (p *parserState) advance() lexer.Token {
	if !p.atEnd() {
		p.current++
	}
	return p.previous()
}

func (p *parserState) atEnd() bool {
	return p.peek().Kind == lexer.EOF
}

func (p *parserState) peek() lexer.Token {
	if p.current >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current]
}

func (p *parserState) previous() lexer.Token {
	return p.tokens[p.current-1]
}

func (p *parserState) errorAt(token lexer.Token, message string) error {
	return &Diagnostic{Code: "SintaxisInvalida", Message: message, Position: token.Start}
}
