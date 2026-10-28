package evaluator_test

import (
	"errors"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/evaluator"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/parser"
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

func callSource(t *testing.T, source, functionName string) value.Value {
	t.Helper()
	interpreter := loadSource(t, source)
	result, err := interpreter.Call(functionName)
	if err != nil {
		t.Fatalf("Call(%s) error = %v", functionName, err)
	}
	return result
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
