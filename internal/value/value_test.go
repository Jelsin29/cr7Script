package value_test

import (
	"math"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/value"
)

func TestNumberRejectsNonFiniteValues(t *testing.T) {
	for _, number := range []float64{math.Inf(1), math.Inf(-1), math.NaN()} {
		if _, err := value.NewNumber(number); err == nil {
			t.Fatalf("NewNumber(%v) succeeded, want a non-finite-number error", number)
		}
	}
	if _, err := value.NewNumber(3.5); err != nil {
		t.Fatalf("NewNumber(3.5) error = %v, want nil", err)
	}
}

func TestTypeLabelsAndEquality(t *testing.T) {
	one := mustNumber(t, 1)
	otherOne := mustNumber(t, 1)
	text := value.NewText("1")
	truth := value.NewBoolean(true)
	null := value.Null()
	list := value.NewList(one)
	listAlias := list
	separateList := value.NewList(otherOne)
	dictionary := value.NewDictionary()
	dictionaryAlias := dictionary
	function := value.NewFunction(testCallable{name: "score"})
	functionAlias := function
	separateFunction := value.NewFunction(testCallable{name: "score"})
	module := value.NewModule(testModule{})
	moduleAlias := module
	separateModule := value.NewModule(testModule{})
	runtimeError := value.NewError(testError{code: "FaltaDelUsuario"})
	errorAlias := runtimeError
	separateError := value.NewError(testError{code: "FaltaDelUsuario"})

	cases := []struct {
		name  string
		left  value.Value
		right value.Value
		want  bool
	}{
		{name: "equal numbers", left: one, right: otherOne, want: true},
		{name: "unequal scalar types", left: one, right: text, want: false},
		{name: "boolean does not equal number", left: truth, right: mustNumber(t, 1), want: false},
		{name: "same list identity", left: list, right: listAlias, want: true},
		{name: "different list identities", left: list, right: separateList, want: false},
		{name: "same dictionary identity", left: dictionary, right: dictionaryAlias, want: true},
		{name: "different dictionary identities", left: dictionary, right: value.NewDictionary(), want: false},
		{name: "same function identity", left: function, right: functionAlias, want: true},
		{name: "different function identities", left: function, right: separateFunction, want: false},
		{name: "same module identity", left: module, right: moduleAlias, want: true},
		{name: "different module identities", left: module, right: separateModule, want: false},
		{name: "same error identity", left: runtimeError, right: errorAlias, want: true},
		{name: "different error identities", left: runtimeError, right: separateError, want: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := value.Equal(test.left, test.right); got != test.want {
				t.Fatalf("Equal(%s, %s) = %t, want %t", value.TypeLabel(test.left), value.TypeLabel(test.right), got, test.want)
			}
		})
	}

	labels := []struct {
		value value.Value
		want  string
	}{
		{one, "numero"},
		{text, "texto"},
		{truth, "booleano"},
		{null, "nulo"},
		{list, "lista"},
		{dictionary, "diccionario"},
		{function, "funcion"},
		{module, "modulo"},
		{runtimeError, "error"},
	}
	for _, test := range labels {
		if got := value.TypeLabel(test.value); got != test.want {
			t.Errorf("TypeLabel(%v) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestListMutationAndSnapshots(t *testing.T) {
	listValue := value.NewList(mustNumber(t, 7))
	alias := listValue
	list, ok := listValue.List()
	if !ok {
		t.Fatal("List() reported a non-list value")
	}
	if !list.Set(0, mustNumber(t, 10)) {
		t.Fatal("Set(0) failed for an existing element")
	}
	before := list.Snapshot()
	list.Append(value.NewText("CR"))

	if got := value.Display(alias); got != `[10, "CR"]` {
		t.Fatalf("Display(alias) = %q, want [10, \\\"CR\\\"]", got)
	}
	if len(before) != 1 || value.Display(before[0]) != "10" {
		t.Fatalf("Snapshot before append = %#v, want a one-element copy containing 10", before)
	}
	if got := list.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
}

func TestListPopRemovesLastValueInPlaceAndReportsEmpty(t *testing.T) {
	listValue := value.NewList(mustNumber(t, 1), value.Null())
	alias := listValue
	list, ok := listValue.List()
	if !ok {
		t.Fatal("List() reported a non-list value")
	}

	removed, ok := list.Pop()
	if !ok || removed.Kind() != value.NullKind {
		t.Fatalf("Pop() = (%s, %t), want (nada, true)", value.Display(removed), ok)
	}
	aliasList, _ := alias.List()
	if got := aliasList.Len(); got != 1 {
		t.Fatalf("alias.Len() after Pop() = %d, want 1", got)
	}
	remaining, exists := aliasList.Get(0)
	if !exists || value.Display(remaining) != "1" {
		t.Fatalf("remaining value = %s, %t, want 1, true", value.Display(remaining), exists)
	}

	removed, ok = list.Pop()
	if !ok || value.Display(removed) != "1" {
		t.Fatalf("second Pop() = (%s, %t), want (1, true)", value.Display(removed), ok)
	}
	if got := aliasList.Len(); got != 0 {
		t.Fatalf("alias.Len() after removing all values = %d, want 0", got)
	}

	removed, ok = list.Pop()
	if ok || removed.Kind() != value.InvalidKind {
		t.Fatalf("Pop() on empty list = (%s, %t), want (invalid, false)", value.Display(removed), ok)
	}
	if got := list.Len(); got != 0 {
		t.Fatalf("Len() after empty Pop() = %d, want 0", got)
	}

	var nilList *value.List
	if removed, ok := nilList.Pop(); ok || removed.Kind() != value.InvalidKind {
		t.Fatalf("Pop() on nil list = (%s, %t), want (invalid, false)", value.Display(removed), ok)
	}
}

func TestDictionaryUpdatePreservesInsertionOrder(t *testing.T) {
	dictionaryValue := value.NewDictionary()
	alias := dictionaryValue
	dictionary, ok := dictionaryValue.Dictionary()
	if !ok {
		t.Fatal("Dictionary() reported a non-dictionary value")
	}
	dictionary.Set("first", mustNumber(t, 1))
	dictionary.Set("second", mustNumber(t, 2))
	dictionary.Set("first", mustNumber(t, 10))

	if got := value.Display(alias); got != `{"first": 10, "second": 2}` {
		t.Fatalf("Display(dictionary) = %q, want {\"first\": 10, \"second\": 2}", got)
	}
	keys := dictionary.KeysSnapshot()
	if len(keys) != 2 || keys[0] != "first" || keys[1] != "second" {
		t.Fatalf("KeysSnapshot() = %#v, want [first second]", keys)
	}
	if _, exists := dictionary.Get("missing"); exists {
		t.Fatal("Get(missing) reported a present key")
	}
}

func TestDisplayScalarAndCollectionValues(t *testing.T) {
	cases := []struct {
		name  string
		value value.Value
		want  string
	}{
		{name: "whole number", value: mustNumber(t, 7), want: "7"},
		{name: "fraction", value: mustNumber(t, 3.5), want: "3.5"},
		{name: "negative zero", value: mustNumber(t, math.Copysign(0, -1)), want: "0"},
		{name: "true", value: value.NewBoolean(true), want: "gol"},
		{name: "false", value: value.NewBoolean(false), want: "fuera"},
		{name: "null", value: value.Null(), want: "nada"},
		{name: "top-level text", value: value.NewText("Cristiano"), want: "Cristiano"},
		{name: "quoted nested text", value: value.NewList(value.NewText("line\none")), want: `["line\none"]`},
		{name: "opaque function", value: value.NewFunction(testCallable{name: "score"}), want: "<funcion>"},
		{name: "opaque module", value: value.NewModule(testModule{}), want: "<modulo>"},
		{name: "opaque error", value: value.NewError(testError{code: "FaltaDelUsuario"}), want: "<error FaltaDelUsuario>"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := value.Display(test.value); got != test.want {
				t.Fatalf("Display() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDisplayDetectsOnlyActiveCollectionCycles(t *testing.T) {
	selfList := value.NewList()
	list, _ := selfList.List()
	list.Append(selfList)
	if got := value.Display(selfList); got != `["<ciclo>"]` {
		t.Fatalf("Display(self-list) = %q, want [\"<ciclo>\"]", got)
	}

	selfDictionary := value.NewDictionary()
	dictionary, _ := selfDictionary.Dictionary()
	dictionary.Set("self", selfDictionary)
	if got := value.Display(selfDictionary); got != `{"self": "<ciclo>"}` {
		t.Fatalf("Display(self-dictionary) = %q, want {\"self\": \"<ciclo>\"}", got)
	}

	shared := value.NewList(mustNumber(t, 1))
	outer := value.NewList(shared, shared)
	if got := value.Display(outer); got != `[[1], [1]]` {
		t.Fatalf("Display(shared sublist) = %q, want [[1], [1]]", got)
	}
}

func TestTextLengthAndIndexCountUnicodeScalars(t *testing.T) {
	text := "a⚽b"
	if got := value.TextLength(text); got != 3 {
		t.Fatalf("TextLength(%q) = %d, want 3", text, got)
	}
	if got, ok := value.TextAt(text, 1); !ok || got != "⚽" {
		t.Fatalf("TextAt(%q, 1) = %q, %t, want ⚽, true", text, got, ok)
	}
	if got, ok := value.TextAt(text, 2); !ok || got != "b" {
		t.Fatalf("TextAt(%q, 2) = %q, %t, want b, true", text, got, ok)
	}
	for _, index := range []int{-1, 3} {
		if _, ok := value.TextAt(text, index); ok {
			t.Errorf("TextAt(%q, %d) succeeded, want out of range", text, index)
		}
	}
}

func mustNumber(t *testing.T, number float64) value.Value {
	t.Helper()
	result, err := value.NewNumber(number)
	if err != nil {
		t.Fatalf("NewNumber(%v) error = %v", number, err)
	}
	return result
}

type testCallable struct{ name string }

func (f testCallable) Name() string { return f.name }
func (f testCallable) Arity() int   { return 0 }
func (f testCallable) Call([]value.Value) (value.Value, error) {
	return value.Null(), nil
}

type testModule struct{}

func (testModule) Get(string) (value.Value, bool) { return value.Null(), false }

type testError struct{ code string }

func (e testError) Code() string  { return e.code }
func (e testError) Error() string { return e.code }
