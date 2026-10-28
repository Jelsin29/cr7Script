package ast

import "github.com/Jelsin29/cr7Script/internal/lexer"

// Node carries the source range shared by AST declarations, statements, and expressions.
type Node struct {
	Start lexer.Position
	End   lexer.Position
}

type Program struct {
	Node
	Imports      []ImportDecl
	Declarations []Stmt
}

type ImportDecl struct {
	Node
	Path  string
	Alias string
}

type Stmt interface {
	statementNode()
	GetNode() Node
}

type Expr interface {
	expressionNode()
	GetNode() Node
}

type MutableDecl struct {
	Node
	Name        string
	Initializer Expr
}

func (*MutableDecl) statementNode() {}

type ConstDecl struct {
	Node
	Name        string
	Initializer Expr
}

func (*ConstDecl) statementNode() {}

type FunctionDecl struct {
	Node
	Name       string
	Parameters []string
	Body       *BlockStmt
}

func (*FunctionDecl) statementNode() {}

type BlockStmt struct {
	Node
	Statements []Stmt
}

func (*BlockStmt) statementNode() {}

type ExpressionStmt struct {
	Node
	Expression Expr
}

func (*ExpressionStmt) statementNode() {}

type AssignmentStmt struct {
	Node
	Target Expr
	Value  Expr
}

func (*AssignmentStmt) statementNode() {}

type IfStmt struct {
	Node
	Condition Expr
	Then      *BlockStmt
	Else      Stmt
}

func (*IfStmt) statementNode() {}

type WhileStmt struct {
	Node
	Condition Expr
	Body      *BlockStmt
}

func (*WhileStmt) statementNode() {}

type ForEachStmt struct {
	Node
	Name     string
	Iterable Expr
	Body     *BlockStmt
}

func (*ForEachStmt) statementNode() {}

type TryCatchStmt struct {
	Node
	Protected *BlockStmt
	ErrorName string
	Handler   *BlockStmt
}

func (*TryCatchStmt) statementNode() {}

type ReturnStmt struct {
	Node
	Value Expr
}

func (*ReturnStmt) statementNode() {}

type BreakStmt struct {
	Node
}

func (*BreakStmt) statementNode() {}

type ContinueStmt struct {
	Node
}

func (*ContinueStmt) statementNode() {}

type ThrowStmt struct {
	Node
	Value Expr
}

func (*ThrowStmt) statementNode() {}

type IdentifierExpr struct {
	Node
	Name string
}

func (*IdentifierExpr) expressionNode() {}

type NumberExpr struct {
	Node
	Value float64
}

func (*NumberExpr) expressionNode() {}

type StringExpr struct {
	Node
	Value string
}

func (*StringExpr) expressionNode() {}

type BooleanExpr struct {
	Node
	Value bool
}

func (*BooleanExpr) expressionNode() {}

type NullExpr struct {
	Node
}

func (*NullExpr) expressionNode() {}

type UnaryExpr struct {
	Node
	Operator lexer.Kind
	Right    Expr
}

func (*UnaryExpr) expressionNode() {}

type BinaryExpr struct {
	Node
	Left     Expr
	Operator lexer.Kind
	Right    Expr
}

func (*BinaryExpr) expressionNode() {}

type GroupingExpr struct {
	Node
	Expression Expr
}

func (*GroupingExpr) expressionNode() {}

type CallExpr struct {
	Node
	Callee    Expr
	Arguments []Expr
}

func (*CallExpr) expressionNode() {}

type IndexExpr struct {
	Node
	Collection Expr
	Index      Expr
}

func (*IndexExpr) expressionNode() {}

type PropertyExpr struct {
	Node
	Object Expr
	Name   string
}

func (*PropertyExpr) expressionNode() {}

type ListExpr struct {
	Node
	Elements []Expr
}

func (*ListExpr) expressionNode() {}

type DictEntry struct {
	Node
	Key   string
	Value Expr
}

type DictExpr struct {
	Node
	Entries []DictEntry
}

func (*DictExpr) expressionNode() {}
