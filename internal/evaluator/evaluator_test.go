package evaluator_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/ast"
	"github.com/Jelsin29/cr7Script/internal/evaluator"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/parser"
	"github.com/Jelsin29/cr7Script/internal/resolver"
	"github.com/Jelsin29/cr7Script/internal/stdlib"
	"github.com/Jelsin29/cr7Script/internal/value"
)

func TestEvaluateArithmeticAndModuloTruncatesTowardZero(t *testing.T) {
	result := callSource(t, `jugada calculate() { devuelve 7 / 2 + -7 % 3; }`, "calculate")
	expectNumber(t, result, 2.5)
}

func TestEvaluateRejectsInvalidNumericOperations(t *testing.T) {
	cases := []struct {
		name   string
		source string
		code   string
	}{
		{name: "mixed addition", source: `jugada bad() { devuelve 7 + "2"; }`, code: "TipoIncorrecto"},
		{name: "division by zero", source: `jugada bad() { devuelve 7 / 0; }`, code: "DivisionPorCero"},
		{name: "non-finite result", source: `jugada bad() { devuelve 1e308 * 1e308; }`, code: "NumeroNoFinito"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			interpreter := loadSource(t, test.source)
			_, err := interpreter.Call("bad")
			expectRuntimeError(t, err, test.code)
		})
	}
}

func TestLogicalOperatorsShortCircuitAndRequireBooleans(t *testing.T) {
	interpreter := loadSource(t, `
		jugada probe() {
			cantera count = 0;
			jugada increment() { count = count + 1; devuelve gol; }
			fuera y increment();
			gol o increment();
			devuelve count;
		}
		jugada invalid() { devuelve 1 y gol; }
	`)
	result, err := interpreter.Call("probe")
	if err != nil {
		t.Fatalf("Call(probe) error = %v", err)
	}
	expectNumber(t, result, 0)

	_, err = interpreter.Call("invalid")
	expectRuntimeError(t, err, "TipoIncorrecto")
}

func TestRecursiveFunctionsAndModuleFunctionInitialization(t *testing.T) {
	interpreter := loadSource(t, `
		leyenda base = 2;
		leyenda result = doubleBase();
		jugada doubleBase() { devuelve factorial(5) + base; }
		jugada factorial(n) {
			si n <= 1 { devuelve 1; }
			devuelve n * factorial(n - 1);
		}
		jugada readResult() { devuelve result; }
	`)
	result, err := interpreter.Call("readResult")
	if err != nil {
		t.Fatalf("Call(readResult) error = %v", err)
	}
	expectNumber(t, result, 122)
}

func TestParameterReassignmentIsLocalAndCollectionArgumentsKeepIdentity(t *testing.T) {
	result := callSource(t, `
		jugada rebind(score) { score = 10; devuelve score; }
		jugada update(values) { values[0] = 99; }
		jugada verify() {
			cantera outer = 3;
			cantera values = [1];
			cantera result = rebind(outer);
			update(values);
			devuelve outer * 100 + result * 10 + values[0];
		}
	`, "verify")
	expectNumber(t, result, 499)
}

func TestConstantInitializerCannotReadLaterConstantThroughFunction(t *testing.T) {
	tokens, err := lexer.Scan(`
		jugada readLater() { devuelve later; }
		leyenda first = readLater();
		leyenda later = 7;
	`)
	if err != nil {
		t.Fatalf("lexer.Scan() error = %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse() error = %v", err)
	}
	interpreter := evaluator.New()
	err = interpreter.Load(program)
	expectRuntimeError(t, err, "NombreSinInicializar")
}

func TestLexicalMutationCollectionsAndForeachSnapshot(t *testing.T) {
	interpreter := loadSource(t, `
		jugada state() {
			cantera values = [1, 2, 3];
			cantera total = 0;
			temporada item en values {
				si item == 1 { values[1] = 99; }
				total = total + item;
			}
			cantera stats = {"total": total};
			stats["total"] = stats["total"] + 1;
			devuelve stats["total"] + values[1];
		}
		jugada dictionarySnapshot() {
			cantera data = {"one": 1, "two": 2};
			cantera visited = 0;
			temporada key en data {
				visited = visited + 1;
				si key == "one" { data["three"] = 3; }
			}
			cantera finalCount = 0;
			temporada key en data { finalCount = finalCount + 1; }
			devuelve visited * 10 + finalCount;
		}
	`)
	result, err := interpreter.Call("state")
	if err != nil {
		t.Fatalf("Call(state) error = %v", err)
	}
	expectNumber(t, result, 106)

	dictionaryResult, err := interpreter.Call("dictionarySnapshot")
	if err != nil {
		t.Fatalf("Call(dictionarySnapshot) error = %v", err)
	}
	expectNumber(t, dictionaryResult, 23)
}

func TestClosuresCaptureFreshLoopBindings(t *testing.T) {
	result := callSource(t, `
		jugada captureAll() {
			cantera captures = [nada, nada];
			temporada item en [1, 2] {
				jugada capture() { devuelve item; }
				si item == 1 { captures[0] = capture; }
				sino { captures[1] = capture; }
			}
			devuelve captures[0]() + captures[1]();
		}
	`, "captureAll")
	expectNumber(t, result, 3)
}

func TestWhileLoopContinueAndBreak(t *testing.T) {
	result := callSource(t, `
		jugada tally() {
			cantera sum = 0;
			cantera n = 0;
			entrena n < 6 {
				n = n + 1;
				si n == 2 { siguiente; }
				si n == 5 { vestuario; }
				sum = sum + n;
			}
			devuelve sum;
		}
	`, "tally")
	expectNumber(t, result, 8)
}

func TestTextIndexingCountsUnicodeScalars(t *testing.T) {
	interpreter := loadSource(t, `
		jugada selectBall() { devuelve "a⚽b"[1]; }
		jugada selectFollowingScalar() { devuelve "a⚽b"[2]; }
	`)
	for _, test := range []struct {
		name string
		want string
	}{
		{name: "selectBall", want: "⚽"},
		{name: "selectFollowingScalar", want: "b"},
	} {
		result, err := interpreter.Call(test.name)
		if err != nil {
			t.Fatalf("Call(%s) error = %v", test.name, err)
		}
		text, ok := result.Text()
		if !ok || text != test.want {
			t.Errorf("Call(%s) = %q (%s), want text %q", test.name, value.Display(result), value.TypeLabel(result), test.want)
		}
	}
}

func TestCallChecksExactArity(t *testing.T) {
	interpreter := loadSource(t, `jugada one(value) { devuelve value; }`)
	_, err := interpreter.Call("one")
	expectRuntimeError(t, err, "AridadIncorrecta")
}

func TestTryCatchExposesUserErrorAndPreservesEffects(t *testing.T) {
	interpreter := loadSourceFrom(t, `jugada recover() {
cantera count = 0;
remontada {
count = count + 1;
falta "boom";
} var failure {
devuelve count == 1 y failure.codigo == "FaltaDelUsuario" y failure.mensaje == "boom" y failure.archivo == "fixture.cr7" y failure.linea == 5 y failure.columna == 1;
}
}`, "fixture.cr7")
	result, err := interpreter.Call("recover")
	if err != nil {
		t.Fatalf("Call(recover) error = %v", err)
	}
	boolean, ok := result.Boolean()
	if !ok || !boolean {
		t.Fatalf("Call(recover) = %s (%s), want gol", value.Display(result), value.TypeLabel(result))
	}
}

func TestRethrownRuntimeErrorPreservesIdentity(t *testing.T) {
	result := callSource(t, `
		jugada rethrow() {
			remontada { falta "boom"; } var original {
				remontada { falta original; } var repeated {
					devuelve original == repeated;
				}
			}
		}
	`, "rethrow")
	boolean, ok := result.Boolean()
	if !ok || !boolean {
		t.Fatalf("Call(rethrow) = %s (%s), want gol", value.Display(result), value.TypeLabel(result))
	}
}

func TestRecoveryDoesNotCatchReturnBreakOrContinue(t *testing.T) {
	interpreter := loadSource(t, `
		jugada returnFromProtected() {
			remontada { devuelve 7; } var failure { devuelve 0; }
		}
		jugada loopSignals() {
			cantera count = 0;
			entrena count < 4 {
				count = count + 1;
				remontada {
					si count == 2 { siguiente; }
					si count == 4 { vestuario; }
				} var failure { devuelve 100; }
			}
			devuelve count;
		}
	`)
	for _, test := range []struct {
		name string
		want float64
	}{
		{name: "returnFromProtected", want: 7},
		{name: "loopSignals", want: 4},
	} {
		result, err := interpreter.Call(test.name)
		if err != nil {
			t.Fatalf("Call(%s) error = %v", test.name, err)
		}
		expectNumber(t, result, test.want)
	}
}

func TestRecoveryDoesNotCatchStaticErrorsAndCatchBindingIsImmutable(t *testing.T) {
	cases := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "undefined name is resolved before execution",
			source: `jugada invalid() {
remontada { devuelve missing; } var failure { devuelve nada; }
}`,
			code: "NombreNoDefinido",
		},
		{
			name: "catch binding cannot be reassigned",
			source: `jugada invalid() {
remontada { falta "boom"; } var failure { failure = nada; }
}`,
			code: "ReasignacionDeLeyenda",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			tokens, err := lexer.Scan(test.source)
			if err != nil {
				t.Fatalf("lexer.Scan() error = %v", err)
			}
			program, err := parser.Parse(tokens)
			if err != nil {
				t.Fatalf("parser.Parse() error = %v", err)
			}
			interpreter := evaluator.New()
			err = interpreter.Load(program)
			var diagnostic *resolver.Diagnostic
			if !errors.As(err, &diagnostic) {
				t.Fatalf("Load() error = %T (%v), want *resolver.Diagnostic", err, err)
			}
			if diagnostic.Code != test.code {
				t.Fatalf("diagnostic.Code = %q, want %q", diagnostic.Code, test.code)
			}
		})
	}
}

func TestRuntimeErrorCarriesFileLocationAndInnermostFirstStack(t *testing.T) {
	interpreter := loadSourceFrom(t, `jugada leaf() {
 devuelve 7 / 0;
}
jugada outer() {
 devuelve leaf();
}`, "fixture.cr7")
	_, err := interpreter.Call("outer")
	expectRuntimeError(t, err, "DivisionPorCero")
	var runtimeError *evaluator.RuntimeError
	if !errors.As(err, &runtimeError) {
		t.Fatalf("error type = %T, want *evaluator.RuntimeError", err)
	}
	if runtimeError.File() != "fixture.cr7" {
		t.Errorf("RuntimeError.File() = %q, want fixture.cr7", runtimeError.File())
	}
	position := runtimeError.Position()
	if position.Line != 2 || position.Column != 11 {
		t.Errorf("RuntimeError.Position() = %d:%d, want 2:11", position.Line, position.Column)
	}
	frames := runtimeError.Stack()
	if len(frames) != 2 {
		t.Fatalf("RuntimeError.Stack() has %d frames, want 2: %#v", len(frames), frames)
	}
	if frames[0].Function != "leaf" || frames[1].Function != "outer" {
		t.Errorf("stack functions = %q, %q, want leaf, outer", frames[0].Function, frames[1].Function)
	}
	if frames[0].File != "fixture.cr7" || frames[1].File != "fixture.cr7" {
		t.Errorf("stack files = %q, %q, want fixture.cr7 for both frames", frames[0].File, frames[1].File)
	}
	if frames[0].Position.Line != 5 || frames[1].Position.Line != 4 {
		t.Errorf("stack call lines = %d, %d, want 5, 4", frames[0].Position.Line, frames[1].Position.Line)
	}
}

func TestInterpreterInvokesVariadicStandardLibraryFunctions(t *testing.T) {
	var output bytes.Buffer
	library := stdlib.New(stdlib.Options{Output: &output})
	interpreter := evaluator.NewWithRuntime(library.Prelude(), evaluator.NewExecutionContext())
	program := parseProgramForTest(t, `jugada empty() {
siuuu();
}
jugada several() {
siuuu("one", 2, "three");
}`)
	if err := interpreter.Load(program); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for _, name := range []string{"empty", "several"} {
		if _, err := interpreter.Call(name); err != nil {
			t.Fatalf("Call(%s) error = %v", name, err)
		}
	}
	if got, want := output.String(), "\none 2 three\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestLoadFromStillRequiresTheModuleLoaderForImports(t *testing.T) {
	program := parseProgramForTest(t, `fichar "./dependency.cr7" como dependency;`)
	err := evaluator.New().LoadFrom(program, "main.cr7")
	if err == nil || err.Error() != "module imports require the module loader" {
		t.Fatalf("LoadFrom() error = %v, want the existing module-loader requirement", err)
	}
}

func TestInjectedPreludeAndImportedModuleBindingsAreReadOnlyExports(t *testing.T) {
	prelude := map[string]value.Value{
		"numero": value.NewFunction(testCallable{name: "numero", arity: 1, result: numberForTest(t, 7)}),
	}
	interpreter := evaluator.NewWithRuntime(prelude, evaluator.NewExecutionContext())
	program := parseProgramForTest(t, `fichar "dependency" como dependency;
leyenda configured = dependency.answer;
jugada run() { devuelve numero("7") + configured; }`)
	imports := map[string]value.Value{
		"dependency": value.NewModule(testModule{"answer": numberForTest(t, 35)}),
	}
	if err := interpreter.LoadFromWithImports(program, "main.cr7", imports); err != nil {
		t.Fatalf("LoadFromWithImports() error = %v", err)
	}
	result, err := interpreter.Call("run")
	if err != nil {
		t.Fatalf("Call(run) error = %v", err)
	}
	expectNumber(t, result, 42)

	exports, ok := interpreter.Exports()
	if !ok {
		t.Fatal("Exports() reports an unloaded module")
	}
	for _, name := range []string{"configured", "run"} {
		if _, exists := exports.Get(name); !exists {
			t.Errorf("module exports omit %q", name)
		}
	}
	for _, name := range []string{"dependency", "native"} {
		if _, exists := exports.Get(name); exists {
			t.Errorf("module exports unexpectedly include %q", name)
		}
	}

	invalid := evaluator.New()
	invalidProgram := parseProgramForTest(t, `fichar "dependency" como dependency;
jugada run() { dependency = nada; devuelve nada; }`)
	err = invalid.LoadFromWithImports(invalidProgram, "main.cr7", imports)
	var diagnostic *resolver.Diagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "ReasignacionDeLeyenda" {
		t.Fatalf("LoadFromWithImports() error = %T (%v), want immutable-import diagnostic", err, err)
	}
}

func TestInterpretersShareCrossModuleRuntimeStack(t *testing.T) {
	context := evaluator.NewExecutionContext()
	dependency := evaluator.NewWithRuntime(nil, context)
	dependencyProgram := parseProgramForTest(t, `jugada fail() {
 devuelve 7 / 0;
}`)
	if err := dependency.LoadFrom(dependencyProgram, "dependency.cr7"); err != nil {
		t.Fatalf("dependency.LoadFrom() error = %v", err)
	}

	main := evaluator.NewWithRuntime(nil, context)
	mainProgram := parseProgramForTest(t, `fichar "dependency" como dependency;
jugada run() {
 devuelve dependency.fail();
}`)
	dependencyExports, ok := dependency.Exports()
	if !ok {
		t.Fatal("dependency.Exports() reports an unloaded module")
	}
	if err := main.LoadFromWithImports(mainProgram, "main.cr7", map[string]value.Value{
		"dependency": value.NewModule(dependencyExports),
	}); err != nil {
		t.Fatalf("main.LoadFromWithImports() error = %v", err)
	}
	_, err := main.Call("run")
	expectRuntimeError(t, err, "DivisionPorCero")
	var runtimeError *evaluator.RuntimeError
	if !errors.As(err, &runtimeError) {
		t.Fatalf("error type = %T, want *evaluator.RuntimeError", err)
	}
	if runtimeError.File() != "dependency.cr7" {
		t.Errorf("RuntimeError.File() = %q, want dependency.cr7", runtimeError.File())
	}
	frames := runtimeError.Stack()
	if len(frames) != 2 || frames[0].Function != "fail" || frames[1].Function != "run" {
		t.Fatalf("RuntimeError.Stack() = %#v, want innermost-first fail, run", frames)
	}
	if frames[0].File != "main.cr7" || frames[1].File != "main.cr7" {
		t.Errorf("stack call-site files = %q, %q, want main.cr7", frames[0].File, frames[1].File)
	}
}

func TestNativeRuntimeErrorUsesCallSite(t *testing.T) {
	prelude := map[string]value.Value{
		"entrevista": value.NewFunction(testCallable{
			name:  "entrevista",
			arity: 1,
			err:   evaluator.NewRuntimeError("EntradaSalida", "native failure", lexer.Position{}),
		}),
	}
	interpreter := evaluator.NewWithRuntime(prelude, evaluator.NewExecutionContext())
	if err := interpreter.LoadFrom(parseProgramForTest(t, `jugada run() {
  devuelve entrevista("prompt");
}`), "main.cr7"); err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	_, err := interpreter.Call("run")
	expectRuntimeError(t, err, "EntradaSalida")
	var runtimeError *evaluator.RuntimeError
	if !errors.As(err, &runtimeError) {
		t.Fatalf("error type = %T, want *evaluator.RuntimeError", err)
	}
	position := runtimeError.Position()
	if position.Line != 2 || position.Column != 12 {
		t.Errorf("RuntimeError.Position() = %d:%d, want 2:12", position.Line, position.Column)
	}
	if runtimeError.File() != "main.cr7" {
		t.Errorf("RuntimeError.File() = %q, want main.cr7", runtimeError.File())
	}
}

type testModule map[string]value.Value

func (module testModule) Get(name string) (value.Value, bool) {
	result, exists := module[name]
	return result, exists
}

type testCallable struct {
	name   string
	arity  int
	result value.Value
	err    error
}

func (function testCallable) Name() string { return function.name }
func (function testCallable) Arity() int   { return function.arity }
func (function testCallable) Call([]value.Value) (value.Value, error) {
	return function.result, function.err
}

func numberForTest(t *testing.T, number float64) value.Value {
	t.Helper()
	result, err := value.NewNumber(number)
	if err != nil {
		t.Fatalf("value.NewNumber(%v) error = %v", number, err)
	}
	return result
}

func parseProgramForTest(t *testing.T, source string) *ast.Program {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan() error = %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse() error = %v", err)
	}
	return program
}

func callSource(t *testing.T, source, functionName string) value.Value {
	t.Helper()
	interpreter := loadSource(t, source)
	result, err := interpreter.Call(functionName)
	if err != nil {
		t.Fatalf("Call(%s) error = %v", functionName, err)
	}
	return result
}

func loadSourceFrom(t *testing.T, source, fileName string) *evaluator.Interpreter {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan() error = %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse() error = %v", err)
	}
	interpreter := evaluator.New()
	if err := interpreter.LoadFrom(program, fileName); err != nil {
		t.Fatalf("Interpreter.LoadFrom() error = %v", err)
	}
	return interpreter
}

func loadSource(t *testing.T, source string) *evaluator.Interpreter {
	t.Helper()
	tokens, err := lexer.Scan(source)
	if err != nil {
		t.Fatalf("lexer.Scan() error = %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parser.Parse() error = %v", err)
	}
	interpreter := evaluator.New()
	if err := interpreter.Load(program); err != nil {
		t.Fatalf("Interpreter.Load() error = %v", err)
	}
	return interpreter
}

func expectNumber(t *testing.T, got value.Value, want float64) {
	t.Helper()
	number, ok := got.Number()
	if !ok || number != want {
		t.Fatalf("result = %s (%s), want %v", value.Display(got), value.TypeLabel(got), want)
	}
}

func expectRuntimeError(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want %s", code)
	}
	var runtimeError *evaluator.RuntimeError
	if !errors.As(err, &runtimeError) {
		t.Fatalf("error type = %T, want *evaluator.RuntimeError (%v)", err, err)
	}
	if got := runtimeError.Code(); got != code {
		t.Fatalf("RuntimeError.Code() = %q, want %q", got, code)
	}
}
