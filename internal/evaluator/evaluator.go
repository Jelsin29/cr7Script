// Package evaluator executes validated CR7Script syntax trees.
package evaluator

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/resolver"
	"github.com/Jelsin29/cr7Script/internal/value"
)

const maxCallDepth = 1024

// RuntimeError is a source-located language failure. Runtime errors are also
// exposed as immutable CR7Script error values.
type RuntimeError struct {
	code     string
	message  string
	position lexer.Position
}

// Code returns the stable language error code.
func (e *RuntimeError) Code() string { return e.code }

// Message returns the human-readable error message.
func (e *RuntimeError) Message() string { return e.message }

// Position returns the one-based source location.
func (e *RuntimeError) Position() lexer.Position { return e.position }

// Error formats the code, source location, and message.
func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%s at %d:%d: %s", e.code, e.position.Line, e.position.Column, e.message)
}

func runtimeError(code, message string, position lexer.Position) *RuntimeError {
	return &RuntimeError{code: code, message: message, position: position}
}

type binding struct {
	value       value.Value
	mutable     bool
	initialized bool
}

type environment struct {
	parent   *environment
	bindings map[string]*binding
}

func newEnvironment(parent *environment) *environment {
	return &environment{parent: parent, bindings: make(map[string]*binding)}
}

func (environment *environment) declare(name string, mutable bool) {
	if _, exists := environment.bindings[name]; !exists {
		environment.bindings[name] = &binding{mutable: mutable}
	}
}

func (environment *environment) define(name string, runtimeValue value.Value, mutable bool) {
	environment.bindings[name] = &binding{value: runtimeValue, mutable: mutable, initialized: true}
}

func (environment *environment) initialize(name string, runtimeValue value.Value, position lexer.Position) error {
	item, exists := environment.bindings[name]
	if !exists {
		return runtimeError("NombreNoDefinido", "undefined name", position)
	}
	if item.initialized {
		return runtimeError("DeclaracionDuplicada", "binding is already initialized", position)
	}
	item.value = runtimeValue
	item.initialized = true
	return nil
}

func (environment *environment) get(name string, position lexer.Position) (value.Value, error) {
	for current := environment; current != nil; current = current.parent {
		if item, exists := current.bindings[name]; exists {
			if !item.initialized {
				return value.Value{}, runtimeError("NombreSinInicializar", "binding read before initialization", position)
			}
			return item.value, nil
		}
	}
	return value.Value{}, runtimeError("NombreNoDefinido", "undefined name", position)
}

func (environment *environment) assign(name string, runtimeValue value.Value, position lexer.Position) error {
	for current := environment; current != nil; current = current.parent {
		if item, exists := current.bindings[name]; exists {
			if !item.initialized {
				return runtimeError("NombreSinInicializar", "binding assigned before initialization", position)
			}
			if !item.mutable {
				return runtimeError("ReasignacionDeLeyenda", "cannot reassign an immutable binding", position)
			}
			item.value = runtimeValue
			return nil
		}
	}
	return runtimeError("NombreNoDefinido", "undefined name", position)
}

// Interpreter holds one loaded module's global environment and call stack state.
type Interpreter struct {
	globals *environment
	loaded  bool
	depth   int
}

// New creates an interpreter without loading a module.
func New() *Interpreter { return &Interpreter{globals: newEnvironment(nil)} }

// Load statically validates and initializes a module. Top-level functions are
// installed before constants; constants then evaluate in source order.
func (interpreter *Interpreter) Load(program *ast.Program) error {
	if interpreter.loaded {
		return errors.New("interpreter already has a loaded module")
	}
	if err := resolver.Resolve(program); err != nil {
		return err
	}
	if len(program.Imports) != 0 {
		return errors.New("module imports require the module loader")
	}

	interpreter.globals = newEnvironment(nil)
	for _, declaration := range program.Declarations {
		switch declaration := declaration.(type) {
		case *ast.ConstDecl:
			interpreter.globals.declare(declaration.Name, false)
		case *ast.FunctionDecl:
			interpreter.globals.declare(declaration.Name, true)
		}
	}
	for _, declaration := range program.Declarations {
		function, ok := declaration.(*ast.FunctionDecl)
		if !ok {
			continue
		}
		callable := &userFunction{declaration: function, closure: interpreter.globals, interpreter: interpreter}
		if err := interpreter.globals.initialize(function.Name, value.NewFunction(callable), function.Start); err != nil {
			return err
		}
	}
	for _, declaration := range program.Declarations {
		constant, ok := declaration.(*ast.ConstDecl)
		if !ok {
			continue
		}
		initializer, err := interpreter.evaluateExpression(constant.Initializer, interpreter.globals)
		if err != nil {
			return err
		}
		if err := interpreter.globals.initialize(constant.Name, initializer, constant.Start); err != nil {
			return err
		}
	}
	interpreter.loaded = true
	return nil
}

// Call invokes a named top-level function with the supplied arguments.
func (interpreter *Interpreter) Call(name string, arguments ...value.Value) (value.Value, error) {
	callableValue, err := interpreter.globals.get(name, lexer.Position{})
	if err != nil {
		return value.Value{}, err
	}
	function, ok := callableValue.Function()
	if !ok {
		return value.Value{}, runtimeError("TipoIncorrecto", "top-level value is not callable", lexer.Position{})
	}
	return interpreter.invoke(function, arguments, lexer.Position{})
}

// EvaluateExpression evaluates one expression in the loaded module environment.
func (interpreter *Interpreter) EvaluateExpression(expression ast.Expr) (value.Value, error) {
	return interpreter.evaluateExpression(expression, interpreter.globals)
}

type userFunction struct {
	declaration *ast.FunctionDecl
	closure     *environment
	interpreter *Interpreter
}

func (function *userFunction) Name() string { return function.declaration.Name }
func (function *userFunction) Arity() int   { return len(function.declaration.Parameters) }
func (function *userFunction) Call(arguments []value.Value) (value.Value, error) {
	return function.interpreter.callUserFunction(function, arguments)
}

func (interpreter *Interpreter) invoke(function value.Callable, arguments []value.Value, position lexer.Position) (value.Value, error) {
	if len(arguments) != function.Arity() {
		return value.Value{}, runtimeError(
			"AridadIncorrecta",
			fmt.Sprintf("%s expects %d arguments, received %d", function.Name(), function.Arity(), len(arguments)),
			position,
		)
	}
	return function.Call(arguments)
}

func (interpreter *Interpreter) callUserFunction(function *userFunction, arguments []value.Value) (value.Value, error) {
	if interpreter.depth >= maxCallDepth {
		return value.Value{}, runtimeError("LimiteDeEjecucion", "maximum function call depth reached", function.declaration.Start)
	}
	interpreter.depth++
	defer func() { interpreter.depth-- }()

	functionEnvironment := newEnvironment(function.closure)
	for index, parameter := range function.declaration.Parameters {
		functionEnvironment.define(parameter, arguments[index], true)
	}
	interpreter.prepareDeclarations(functionEnvironment, function.declaration.Body.Statements)
	result, err := interpreter.executeStatements(function.declaration.Body.Statements, functionEnvironment)
	if err != nil {
		return value.Value{}, err
	}
	if result.kind == returnFlow {
		return result.value, nil
	}
	return value.Null(), nil
}

type flowKind uint8

const (
	normalFlow flowKind = iota
	returnFlow
	breakFlow
	continueFlow
)

type flow struct {
	kind  flowKind
	value value.Value
}

func (interpreter *Interpreter) prepareDeclarations(environment *environment, statements []ast.Stmt) {
	for _, statement := range statements {
		switch declaration := statement.(type) {
		case *ast.MutableDecl:
			environment.declare(declaration.Name, true)
		case *ast.ConstDecl:
			environment.declare(declaration.Name, false)
		case *ast.FunctionDecl:
			environment.declare(declaration.Name, true)
		}
	}
}

func (interpreter *Interpreter) executeBlock(block *ast.BlockStmt, parent *environment) (flow, error) {
	blockEnvironment := newEnvironment(parent)
	interpreter.prepareDeclarations(blockEnvironment, block.Statements)
	return interpreter.executeStatements(block.Statements, blockEnvironment)
}

func (interpreter *Interpreter) executeStatements(statements []ast.Stmt, environment *environment) (flow, error) {
	for _, statement := range statements {
		result, err := interpreter.executeStatement(statement, environment)
		if err != nil {
			return flow{}, err
		}
		if result.kind != normalFlow {
			return result, nil
		}
	}
	return flow{}, nil
}

func (interpreter *Interpreter) executeStatement(statement ast.Stmt, environment *environment) (flow, error) {
	switch statement := statement.(type) {
	case *ast.MutableDecl:
		initializer, err := interpreter.evaluateExpression(statement.Initializer, environment)
		if err != nil {
			return flow{}, err
		}
		return flow{}, environment.initialize(statement.Name, initializer, statement.Start)
	case *ast.ConstDecl:
		initializer, err := interpreter.evaluateExpression(statement.Initializer, environment)
		if err != nil {
			return flow{}, err
		}
		return flow{}, environment.initialize(statement.Name, initializer, statement.Start)
	case *ast.FunctionDecl:
		callable := &userFunction{declaration: statement, closure: environment, interpreter: interpreter}
		return flow{}, environment.initialize(statement.Name, value.NewFunction(callable), statement.Start)
	case *ast.BlockStmt:
		return interpreter.executeBlock(statement, environment)
	case *ast.ExpressionStmt:
		_, err := interpreter.evaluateExpression(statement.Expression, environment)
		return flow{}, err
	case *ast.AssignmentStmt:
		return flow{}, interpreter.executeAssignment(statement, environment)
	case *ast.IfStmt:
		condition, err := interpreter.evaluateExpression(statement.Condition, environment)
		if err != nil {
			return flow{}, err
		}
		truth, err := requireBoolean(condition, statement.Condition.GetNode().Start)
		if err != nil {
			return flow{}, err
		}
		if truth {
			return interpreter.executeBlock(statement.Then, environment)
		}
		if statement.Else == nil {
			return flow{}, nil
		}
		return interpreter.executeStatement(statement.Else, environment)
	case *ast.WhileStmt:
		for {
			condition, err := interpreter.evaluateExpression(statement.Condition, environment)
			if err != nil {
				return flow{}, err
			}
			truth, err := requireBoolean(condition, statement.Condition.GetNode().Start)
			if err != nil {
				return flow{}, err
			}
			if !truth {
				return flow{}, nil
			}
			result, err := interpreter.executeBlock(statement.Body, environment)
			if err != nil {
				return flow{}, err
			}
			switch result.kind {
			case returnFlow:
				return result, nil
			case breakFlow:
				return flow{}, nil
			case continueFlow:
				continue
			}
		}
	case *ast.ForEachStmt:
		iterable, err := interpreter.evaluateExpression(statement.Iterable, environment)
		if err != nil {
			return flow{}, err
		}
		items, err := snapshotIterable(iterable, statement.Iterable.GetNode().Start)
		if err != nil {
			return flow{}, err
		}
		for _, item := range items {
			iterationEnvironment := newEnvironment(environment)
			iterationEnvironment.define(statement.Name, item, true)
			result, err := interpreter.executeBlock(statement.Body, iterationEnvironment)
			if err != nil {
				return flow{}, err
			}
			switch result.kind {
			case returnFlow:
				return result, nil
			case breakFlow:
				return flow{}, nil
			case continueFlow:
				continue
			}
		}
		return flow{}, nil
	case *ast.ReturnStmt:
		if statement.Value == nil {
			return flow{kind: returnFlow, value: value.Null()}, nil
		}
		result, err := interpreter.evaluateExpression(statement.Value, environment)
		return flow{kind: returnFlow, value: result}, err
	case *ast.BreakStmt:
		return flow{kind: breakFlow}, nil
	case *ast.ContinueStmt:
		return flow{kind: continueFlow}, nil
	case *ast.ThrowStmt:
		thrown, err := interpreter.evaluateExpression(statement.Value, environment)
		if err != nil {
			return flow{}, err
		}
		if text, ok := thrown.Text(); ok {
			return flow{}, runtimeError("FaltaDelUsuario", text, statement.Start)
		}
		if runtimeError, ok := thrown.ErrorValue(); ok {
			return flow{}, runtimeError
		}
		return flow{}, runtimeError("TipoIncorrecto", "falta expects text or an error value", statement.Start)
	case *ast.TryCatchStmt:
		return flow{}, fmt.Errorf("evaluator: recovery execution is not implemented")
	default:
		return flow{}, fmt.Errorf("evaluator: unsupported statement %T", statement)
	}
}

func (interpreter *Interpreter) executeAssignment(statement *ast.AssignmentStmt, environment *environment) error {
	switch target := statement.Target.(type) {
	case *ast.IdentifierExpr:
		assigned, err := interpreter.evaluateExpression(statement.Value, environment)
		if err != nil {
			return err
		}
		return environment.assign(target.Name, assigned, target.Start)
	case *ast.IndexExpr:
		collection, err := interpreter.evaluateExpression(target.Collection, environment)
		if err != nil {
			return err
		}
		index, err := interpreter.evaluateExpression(target.Index, environment)
		if err != nil {
			return err
		}
		assigned, err := interpreter.evaluateExpression(statement.Value, environment)
		if err != nil {
			return err
		}
		return assignIndex(collection, index, assigned, target.Start)
	default:
		return runtimeError("TipoIncorrecto", "invalid assignment target", statement.Start)
	}
}

func (interpreter *Interpreter) evaluateExpression(expression ast.Expr, environment *environment) (value.Value, error) {
	switch expression := expression.(type) {
	case *ast.IdentifierExpr:
		return environment.get(expression.Name, expression.Start)
	case *ast.NumberExpr:
		result, err := value.NewNumber(expression.Value)
		if err != nil {
			return value.Value{}, runtimeError("NumeroNoFinito", "number is not finite", expression.Start)
		}
		return result, nil
	case *ast.StringExpr:
		return value.NewText(expression.Value), nil
	case *ast.BooleanExpr:
		return value.NewBoolean(expression.Value), nil
	case *ast.NullExpr:
		return value.Null(), nil
	case *ast.UnaryExpr:
		right, err := interpreter.evaluateExpression(expression.Right, environment)
		if err != nil {
			return value.Value{}, err
		}
		switch expression.Operator {
		case lexer.Minus:
			number, ok := right.Number()
			if !ok {
				return value.Value{}, typeError("unary '-' requires a number", expression.Start, right)
			}
			return finiteNumber(-number, expression.Start)
		case lexer.No:
			boolean, err := requireBoolean(right, expression.Start)
			if err != nil {
				return value.Value{}, err
			}
			return value.NewBoolean(!boolean), nil
		default:
			return value.Value{}, runtimeError("TipoIncorrecto", "unsupported unary operator", expression.Start)
		}
	case *ast.BinaryExpr:
		return interpreter.evaluateBinary(expression, environment)
	case *ast.GroupingExpr:
		return interpreter.evaluateExpression(expression.Expression, environment)
	case *ast.CallExpr:
		callee, err := interpreter.evaluateExpression(expression.Callee, environment)
		if err != nil {
			return value.Value{}, err
		}
		function, ok := callee.Function()
		if !ok {
			return value.Value{}, typeError("value is not callable", expression.Start, callee)
		}
		arguments := make([]value.Value, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			result, err := interpreter.evaluateExpression(argument, environment)
			if err != nil {
				return value.Value{}, err
			}
			arguments = append(arguments, result)
		}
		return interpreter.invoke(function, arguments, expression.Start)
	case *ast.IndexExpr:
		collection, err := interpreter.evaluateExpression(expression.Collection, environment)
		if err != nil {
			return value.Value{}, err
		}
		index, err := interpreter.evaluateExpression(expression.Index, environment)
		if err != nil {
			return value.Value{}, err
		}
		return indexValue(collection, index, expression.Start)
	case *ast.PropertyExpr:
		object, err := interpreter.evaluateExpression(expression.Object, environment)
		if err != nil {
			return value.Value{}, err
		}
		return propertyValue(object, expression.Name, expression.Start)
	case *ast.ListExpr:
		items := make([]value.Value, 0, len(expression.Elements))
		for _, element := range expression.Elements {
			result, err := interpreter.evaluateExpression(element, environment)
			if err != nil {
				return value.Value{}, err
			}
			items = append(items, result)
		}
		return value.NewList(items...), nil
	case *ast.DictExpr:
		dictionaryValue := value.NewDictionary()
		dictionary, _ := dictionaryValue.Dictionary()
		for _, entry := range expression.Entries {
			result, err := interpreter.evaluateExpression(entry.Value, environment)
			if err != nil {
				return value.Value{}, err
			}
			dictionary.Set(entry.Key, result)
		}
		return dictionaryValue, nil
	default:
		return value.Value{}, runtimeError("SintaxisInvalida", "unsupported expression", expression.GetNode().Start)
	}
}

func (interpreter *Interpreter) evaluateBinary(expression *ast.BinaryExpr, environment *environment) (value.Value, error) {
	left, err := interpreter.evaluateExpression(expression.Left, environment)
	if err != nil {
		return value.Value{}, err
	}
	if expression.Operator == lexer.Y || expression.Operator == lexer.O {
		leftBoolean, err := requireBoolean(left, expression.Left.GetNode().Start)
		if err != nil {
			return value.Value{}, err
		}
		if expression.Operator == lexer.Y && !leftBoolean {
			return value.NewBoolean(false), nil
		}
		if expression.Operator == lexer.O && leftBoolean {
			return value.NewBoolean(true), nil
		}
		right, err := interpreter.evaluateExpression(expression.Right, environment)
		if err != nil {
			return value.Value{}, err
		}
		rightBoolean, err := requireBoolean(right, expression.Right.GetNode().Start)
		if err != nil {
			return value.Value{}, err
		}
		return value.NewBoolean(rightBoolean), nil
	}
	right, err := interpreter.evaluateExpression(expression.Right, environment)
	if err != nil {
		return value.Value{}, err
	}

	switch expression.Operator {
	case lexer.EqualEqual:
		return value.NewBoolean(value.Equal(left, right)), nil
	case lexer.BangEqual:
		return value.NewBoolean(!value.Equal(left, right)), nil
	case lexer.Plus:
		leftNumber, leftIsNumber := left.Number()
		rightNumber, rightIsNumber := right.Number()
		if leftIsNumber && rightIsNumber {
			return finiteNumber(leftNumber+rightNumber, expression.Start)
		}
		leftText, leftIsText := left.Text()
		rightText, rightIsText := right.Text()
		if leftIsText && rightIsText {
			return value.NewText(leftText + rightText), nil
		}
		return value.Value{}, typeError("'+' requires two numbers or two texts", expression.Start, left, right)
	case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
		leftNumber, leftIsNumber := left.Number()
		rightNumber, rightIsNumber := right.Number()
		if !leftIsNumber || !rightIsNumber {
			return value.Value{}, typeError("numeric operator requires two numbers", expression.Start, left, right)
		}
		if (expression.Operator == lexer.Slash || expression.Operator == lexer.Percent) && rightNumber == 0 {
			return value.Value{}, runtimeError("DivisionPorCero", "cannot divide by zero", expression.Start)
		}
		switch expression.Operator {
		case lexer.Minus:
			return finiteNumber(leftNumber-rightNumber, expression.Start)
		case lexer.Star:
			return finiteNumber(leftNumber*rightNumber, expression.Start)
		case lexer.Slash:
			return finiteNumber(leftNumber/rightNumber, expression.Start)
		default:
			return finiteNumber(math.Mod(leftNumber, rightNumber), expression.Start)
		}
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		leftNumber, leftIsNumber := left.Number()
		rightNumber, rightIsNumber := right.Number()
		if !leftIsNumber || !rightIsNumber {
			return value.Value{}, typeError("relational operators require two numbers", expression.Start, left, right)
		}
		switch expression.Operator {
		case lexer.Less:
			return value.NewBoolean(leftNumber < rightNumber), nil
		case lexer.LessEqual:
			return value.NewBoolean(leftNumber <= rightNumber), nil
		case lexer.Greater:
			return value.NewBoolean(leftNumber > rightNumber), nil
		default:
			return value.NewBoolean(leftNumber >= rightNumber), nil
		}
	default:
		return value.Value{}, runtimeError("TipoIncorrecto", "unsupported binary operator", expression.Start)
	}
}

func finiteNumber(number float64, position lexer.Position) (value.Value, error) {
	result, err := value.NewNumber(number)
	if err != nil {
		return value.Value{}, runtimeError("NumeroNoFinito", "numeric result is not finite", position)
	}
	return result, nil
}

func typeError(message string, position lexer.Position, values ...value.Value) *RuntimeError {
	labels := make([]string, len(values))
	for index, runtimeValue := range values {
		labels[index] = value.TypeLabel(runtimeValue)
	}
	if len(labels) != 0 {
		message += " (received " + strings.Join(labels, ", ") + ")"
	}
	return runtimeError("TipoIncorrecto", message, position)
}

func requireBoolean(runtimeValue value.Value, position lexer.Position) (bool, error) {
	boolean, ok := runtimeValue.Boolean()
	if !ok {
		return false, typeError("condition requires a boolean", position, runtimeValue)
	}
	return boolean, nil
}

func snapshotIterable(iterable value.Value, position lexer.Position) ([]value.Value, error) {
	if list, ok := iterable.List(); ok {
		return list.Snapshot(), nil
	}
	if text, ok := iterable.Text(); ok {
		items := make([]value.Value, 0, value.TextLength(text))
		for _, runeValue := range text {
			items = append(items, value.NewText(string(runeValue)))
		}
		return items, nil
	}
	if dictionary, ok := iterable.Dictionary(); ok {
		keys := dictionary.KeysSnapshot()
		items := make([]value.Value, len(keys))
		for index, key := range keys {
			items[index] = value.NewText(key)
		}
		return items, nil
	}
	return nil, typeError("value is not iterable", position, iterable)
}

func indexValue(collection, index value.Value, position lexer.Position) (value.Value, error) {
	if list, ok := collection.List(); ok {
		numericIndex, err := checkedIndex(index, position)
		if err != nil {
			return value.Value{}, err
		}
		result, exists := list.Get(numericIndex)
		if !exists {
			return value.Value{}, runtimeError("IndiceFueraDeCampo", "list index is out of range", position)
		}
		return result, nil
	}
	if dictionary, ok := collection.Dictionary(); ok {
		key, ok := index.Text()
		if !ok {
			return value.Value{}, typeError("dictionary index must be text", position, index)
		}
		result, exists := dictionary.Get(key)
		if !exists {
			return value.Value{}, runtimeError("ClaveAusente", "dictionary key is absent", position)
		}
		return result, nil
	}
	if text, ok := collection.Text(); ok {
		numericIndex, err := checkedIndex(index, position)
		if err != nil {
			return value.Value{}, err
		}
		result, exists := value.TextAt(text, numericIndex)
		if !exists {
			return value.Value{}, runtimeError("IndiceFueraDeCampo", "text index is out of range", position)
		}
		return value.NewText(result), nil
	}
	return value.Value{}, typeError("value cannot be indexed", position, collection)
}

func assignIndex(collection, index, assigned value.Value, position lexer.Position) error {
	if list, ok := collection.List(); ok {
		numericIndex, err := checkedIndex(index, position)
		if err != nil {
			return err
		}
		if !list.Set(numericIndex, assigned) {
			return runtimeError("IndiceFueraDeCampo", "list index is out of range", position)
		}
		return nil
	}
	if dictionary, ok := collection.Dictionary(); ok {
		key, ok := index.Text()
		if !ok {
			return typeError("dictionary index must be text", position, index)
		}
		dictionary.Set(key, assigned)
		return nil
	}
	if _, ok := collection.Text(); ok {
		return typeError("text values cannot be assigned by index", position, collection)
	}
	return typeError("value cannot be assigned by index", position, collection)
}

func checkedIndex(index value.Value, position lexer.Position) (int, error) {
	number, ok := index.Number()
	if !ok {
		return 0, typeError("index must be a number", position, index)
	}
	maximumSafeInteger := float64(9007199254740991)
	maximumInt := float64(int(^uint(0) >> 1))
	if number < 0 || math.Trunc(number) != number || number > maximumSafeInteger || number > maximumInt {
		return 0, runtimeError("IndiceFueraDeCampo", "index must be a non-negative safe integer", position)
	}
	return int(number), nil
}

func propertyValue(object value.Value, name string, position lexer.Position) (value.Value, error) {
	if module, ok := object.Module(); ok {
		result, exists := module.Get(name)
		if !exists {
			return value.Value{}, runtimeError("ClaveAusente", "module export is absent", position)
		}
		return result, nil
	}
	if errorValue, ok := object.ErrorValue(); ok {
		switch name {
		case "codigo":
			return value.NewText(errorValue.Code()), nil
		case "mensaje":
			if diagnostic, ok := errorValue.(*RuntimeError); ok {
				return value.NewText(diagnostic.Message()), nil
			}
			return value.NewText(errorValue.Error()), nil
		case "archivo":
			return value.NewText(""), nil
		case "linea":
			if diagnostic, ok := errorValue.(*RuntimeError); ok {
				return finiteNumber(float64(diagnostic.Position().Line), position)
			}
			return finiteNumber(0, position)
		case "columna":
			if diagnostic, ok := errorValue.(*RuntimeError); ok {
				return finiteNumber(float64(diagnostic.Position().Column), position)
			}
			return finiteNumber(0, position)
		default:
			return value.Value{}, runtimeError("ClaveAusente", "error field is absent", position)
		}
	}
	return value.Value{}, typeError("property access requires a module or error", position, object)
}
