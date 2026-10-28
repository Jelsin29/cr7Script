package resolver

import (
	"fmt"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/lexer"
)

// Diagnostic is a source-located static-analysis error.
type Diagnostic struct {
	Code     string
	Message  string
	Position lexer.Position
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s at %d:%d: %s", d.Code, d.Position.Line, d.Position.Column, d.Message)
}

type binding struct {
	mutable     bool
	initialized bool
}

type scope struct {
	parent   *scope
	bindings map[string]*binding
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, bindings: make(map[string]*binding)}
}

type resolver struct {
	current       *scope
	functionDepth int
	loopDepth     int
}

var preludeNames = []string{
	"siuuu",
	"entrevista",
	"hala_madrid",
	"el_bicho",
	"texto",
	"numero",
	"tipo",
	"palmares",
	"calma",
	"rango",
	"chilena",
	"hat_trick",
	"convoca",
	"banquillo",
	"consulta",
	"contiene",
	"exige",
}

// Resolve checks lexical bindings, initialization order, assignment rules, and control flow.
func Resolve(program *ast.Program) error {
	if program == nil {
		return &Diagnostic{Code: "SintaxisInvalida", Message: "program is nil"}
	}

	moduleScope := newScope(nil)
	for _, name := range preludeNames {
		moduleScope.bindings[name] = &binding{initialized: true}
	}

	r := &resolver{current: moduleScope}
	for _, declaration := range program.Imports {
		if err := r.declare(declaration.Alias, false, true, declaration.Start); err != nil {
			return err
		}
	}
	for _, declaration := range program.Declarations {
		switch declaration := declaration.(type) {
		case *ast.ConstDecl:
			if err := r.declare(declaration.Name, false, false, declaration.Start); err != nil {
				return err
			}
		case *ast.FunctionDecl:
			if err := r.declare(declaration.Name, true, true, declaration.Start); err != nil {
				return err
			}
		}
	}

	for _, declaration := range program.Declarations {
		if constant, ok := declaration.(*ast.ConstDecl); ok {
			if err := r.resolveExpression(constant.Initializer); err != nil {
				return err
			}
			moduleScope.bindings[constant.Name].initialized = true
			continue
		}
		if _, ok := declaration.(*ast.FunctionDecl); ok {
			continue
		}
		if err := r.resolveStatement(declaration); err != nil {
			return err
		}
	}

	for _, declaration := range program.Declarations {
		if function, ok := declaration.(*ast.FunctionDecl); ok {
			if err := r.resolveFunction(function, moduleScope); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *resolver) declare(name string, mutable, initialized bool, position lexer.Position) error {
	if _, exists := r.current.bindings[name]; exists {
		return diagnostic("DeclaracionDuplicada", "duplicate declaration", position)
	}
	r.current.bindings[name] = &binding{mutable: mutable, initialized: initialized}
	return nil
}

func (r *resolver) resolveFunction(function *ast.FunctionDecl, parent *scope) error {
	functionScope := newScope(parent)
	previousScope := r.current
	previousFunctionDepth := r.functionDepth
	previousLoopDepth := r.loopDepth
	r.current = functionScope
	r.functionDepth++
	r.loopDepth = 0

	for _, parameter := range function.Parameters {
		if err := r.declare(parameter, true, true, function.Start); err != nil {
			r.current = previousScope
			r.functionDepth = previousFunctionDepth
			r.loopDepth = previousLoopDepth
			return err
		}
	}
	err := r.resolveStatements(function.Body.Statements, functionScope)
	r.current = previousScope
	r.functionDepth = previousFunctionDepth
	r.loopDepth = previousLoopDepth
	return err
}

func (r *resolver) resolveStatements(statements []ast.Stmt, current *scope) error {
	previousScope := r.current
	r.current = current
	defer func() { r.current = previousScope }()

	for _, statement := range statements {
		switch declaration := statement.(type) {
		case *ast.MutableDecl:
			if err := r.declare(declaration.Name, true, false, declaration.Start); err != nil {
				return err
			}
		case *ast.ConstDecl:
			if err := r.declare(declaration.Name, false, false, declaration.Start); err != nil {
				return err
			}
		case *ast.FunctionDecl:
			if err := r.declare(declaration.Name, true, false, declaration.Start); err != nil {
				return err
			}
		}
	}

	deferredFunctions := make([]*ast.FunctionDecl, 0)
	for _, statement := range statements {
		switch declaration := statement.(type) {
		case *ast.MutableDecl:
			if err := r.resolveExpression(declaration.Initializer); err != nil {
				return err
			}
			current.bindings[declaration.Name].initialized = true
		case *ast.ConstDecl:
			if err := r.resolveExpression(declaration.Initializer); err != nil {
				return err
			}
			current.bindings[declaration.Name].initialized = true
		case *ast.FunctionDecl:
			current.bindings[declaration.Name].initialized = true
			deferredFunctions = append(deferredFunctions, declaration)
		default:
			if err := r.resolveStatement(statement); err != nil {
				return err
			}
		}
	}
	for _, function := range deferredFunctions {
		if err := r.resolveFunction(function, current); err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) resolveStatement(statement ast.Stmt) error {
	switch statement := statement.(type) {
	case *ast.BlockStmt:
		return r.resolveBlock(statement, r.current)
	case *ast.ExpressionStmt:
		return r.resolveExpression(statement.Expression)
	case *ast.AssignmentStmt:
		if err := r.resolveAssignmentTarget(statement.Target); err != nil {
			return err
		}
		return r.resolveExpression(statement.Value)
	case *ast.IfStmt:
		if err := r.resolveExpression(statement.Condition); err != nil {
			return err
		}
		if err := r.resolveBlock(statement.Then, r.current); err != nil {
			return err
		}
		if statement.Else == nil {
			return nil
		}
		return r.resolveStatement(statement.Else)
	case *ast.WhileStmt:
		if err := r.resolveExpression(statement.Condition); err != nil {
			return err
		}
		r.loopDepth++
		err := r.resolveBlock(statement.Body, r.current)
		r.loopDepth--
		return err
	case *ast.ForEachStmt:
		if err := r.resolveExpression(statement.Iterable); err != nil {
			return err
		}
		loopScope := newScope(r.current)
		previousScope := r.current
		r.current = loopScope
		if err := r.declare(statement.Name, true, true, statement.Start); err != nil {
			r.current = previousScope
			return err
		}
		r.loopDepth++
		err := r.resolveBlock(statement.Body, loopScope)
		r.loopDepth--
		r.current = previousScope
		return err
	case *ast.TryCatchStmt:
		if err := r.resolveBlock(statement.Protected, r.current); err != nil {
			return err
		}
		handlerScope := newScope(r.current)
		previousScope := r.current
		r.current = handlerScope
		if err := r.declare(statement.ErrorName, false, true, statement.Start); err != nil {
			r.current = previousScope
			return err
		}
		err := r.resolveStatements(statement.Handler.Statements, handlerScope)
		r.current = previousScope
		return err
	case *ast.ReturnStmt:
		if r.functionDepth == 0 {
			return diagnostic("DevolucionFueraDeFuncion", "return outside a function", statement.Start)
		}
		if statement.Value != nil {
			return r.resolveExpression(statement.Value)
		}
	case *ast.BreakStmt, *ast.ContinueStmt:
		if r.loopDepth == 0 {
			return diagnostic("ControlFueraDeBucle", "loop control outside a loop", nodePosition(statement))
		}
	case *ast.ThrowStmt:
		return r.resolveExpression(statement.Value)
	case *ast.MutableDecl, *ast.ConstDecl, *ast.FunctionDecl:
		return diagnostic("SintaxisInvalida", "declaration is not valid in this context", nodePosition(statement))
	default:
		return diagnostic("SintaxisInvalida", "unknown statement node", nodePosition(statement))
	}
	return nil
}

func (r *resolver) resolveBlock(block *ast.BlockStmt, parent *scope) error {
	if block == nil {
		return diagnostic("SintaxisInvalida", "block is nil", lexer.Position{})
	}
	return r.resolveStatements(block.Statements, newScope(parent))
}

func (r *resolver) resolveExpression(expression ast.Expr) error {
	switch expression := expression.(type) {
	case *ast.IdentifierExpr:
		binding, ok := r.lookup(expression.Name)
		if !ok {
			return diagnostic("NombreNoDefinido", "undefined name", expression.Start)
		}
		if !binding.initialized {
			return diagnostic("NombreSinInicializar", "binding read before initialization", expression.Start)
		}
	case *ast.NumberExpr, *ast.StringExpr, *ast.BooleanExpr, *ast.NullExpr:
		return nil
	case *ast.UnaryExpr:
		return r.resolveExpression(expression.Right)
	case *ast.BinaryExpr:
		if err := r.resolveExpression(expression.Left); err != nil {
			return err
		}
		return r.resolveExpression(expression.Right)
	case *ast.GroupingExpr:
		return r.resolveExpression(expression.Expression)
	case *ast.CallExpr:
		if err := r.resolveExpression(expression.Callee); err != nil {
			return err
		}
		for _, argument := range expression.Arguments {
			if err := r.resolveExpression(argument); err != nil {
				return err
			}
		}
	case *ast.IndexExpr:
		if err := r.resolveExpression(expression.Collection); err != nil {
			return err
		}
		return r.resolveExpression(expression.Index)
	case *ast.PropertyExpr:
		return r.resolveExpression(expression.Object)
	case *ast.ListExpr:
		for _, element := range expression.Elements {
			if err := r.resolveExpression(element); err != nil {
				return err
			}
		}
	case *ast.DictExpr:
		for _, entry := range expression.Entries {
			if err := r.resolveExpression(entry.Value); err != nil {
				return err
			}
		}
	default:
		return diagnostic("SintaxisInvalida", "unknown expression node", nodePosition(expression))
	}
	return nil
}

func (r *resolver) resolveAssignmentTarget(target ast.Expr) error {
	switch target := target.(type) {
	case *ast.IdentifierExpr:
		binding, ok := r.lookup(target.Name)
		if !ok {
			return diagnostic("NombreNoDefinido", "undefined name", target.Start)
		}
		if !binding.initialized {
			return diagnostic("NombreSinInicializar", "binding read before initialization", target.Start)
		}
		if !binding.mutable {
			return diagnostic("ReasignacionDeLeyenda", "cannot reassign an immutable binding", target.Start)
		}
		return nil
	case *ast.IndexExpr:
		if err := r.resolveExpression(target.Collection); err != nil {
			return err
		}
		return r.resolveExpression(target.Index)
	default:
		return diagnostic("SintaxisInvalida", "invalid assignment target", nodePosition(target))
	}
}

func (r *resolver) lookup(name string) (*binding, bool) {
	for current := r.current; current != nil; current = current.parent {
		if binding, ok := current.bindings[name]; ok {
			return binding, true
		}
	}
	return nil, false
}

func diagnostic(code, message string, position lexer.Position) *Diagnostic {
	return &Diagnostic{Code: code, Message: message, Position: position}
}

func nodePosition(node interface{ GetNode() ast.Node }) lexer.Position {
	return node.GetNode().Start
}
