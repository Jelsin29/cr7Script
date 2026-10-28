// Package value defines the runtime values shared by the CR7Script evaluator and
// standard library.
package value

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ErrNonFiniteNumber = errors.New("number must be finite")

// Kind is the runtime category of a value.
type Kind uint8

const (
	InvalidKind Kind = iota
	NumberKind
	TextKind
	BooleanKind
	NullKind
	ListKind
	DictionaryKind
	FunctionKind
	ModuleKind
	ErrorKind
)

// Callable is implemented by user-defined and native functions.
type Callable interface {
	Name() string
	Arity() int
	Call(arguments []Value) (Value, error)
}

// Module exposes values through a read-only namespace.
type Module interface {
	Get(name string) (Value, bool)
}

// ErrorValue is the immutable runtime error payload exposed to CR7Script code.
type ErrorValue interface {
	Code() string
	Error() string
}

type identity struct{ marker byte }

// Value is an immutable tagged value. Collection references point to mutable
// List or Dictionary objects; copying a Value preserves reference identity.
type Value struct {
	kind      Kind
	number    float64
	text      string
	boolean   bool
	reference any
	identity  *identity
}

// NewNumber creates a finite numeric value.
func NewNumber(number float64) (Value, error) {
	if math.IsInf(number, 0) || math.IsNaN(number) {
		return Value{}, ErrNonFiniteNumber
	}
	return Value{kind: NumberKind, number: number}, nil
}

// NewText creates an immutable text value.
func NewText(text string) Value { return Value{kind: TextKind, text: text} }

// NewBoolean creates a boolean value.
func NewBoolean(boolean bool) Value { return Value{kind: BooleanKind, boolean: boolean} }

// Null returns the absent value.
func Null() Value { return Value{kind: NullKind} }

// NewList creates a list containing a copy of values.
func NewList(values ...Value) Value {
	list := &List{values: append([]Value(nil), values...)}
	return newReference(ListKind, list)
}

// NewDictionary creates an empty insertion-ordered dictionary.
func NewDictionary() Value {
	return newReference(DictionaryKind, &Dictionary{positions: make(map[string]int)})
}

// NewFunction creates a function reference.
func NewFunction(function Callable) Value { return newReference(FunctionKind, function) }

// NewModule creates a read-only module namespace reference.
func NewModule(module Module) Value { return newReference(ModuleKind, module) }

// NewError creates an error reference.
func NewError(runtimeError ErrorValue) Value { return newReference(ErrorKind, runtimeError) }

func newReference(kind Kind, reference any) Value {
	return Value{kind: kind, reference: reference, identity: &identity{}}
}

// Kind reports the value's runtime category.
func (v Value) Kind() Kind { return v.kind }

// Number reports the finite number payload, if v is numeric.
func (v Value) Number() (float64, bool) { return v.number, v.kind == NumberKind }

// Text reports the text payload, if v is text.
func (v Value) Text() (string, bool) { return v.text, v.kind == TextKind }

// Boolean reports the boolean payload, if v is boolean.
func (v Value) Boolean() (bool, bool) { return v.boolean, v.kind == BooleanKind }

// List reports the mutable list reference, if v is a list.
func (v Value) List() (*List, bool) {
	list, ok := v.reference.(*List)
	return list, v.kind == ListKind && ok
}

// Dictionary reports the mutable dictionary reference, if v is a dictionary.
func (v Value) Dictionary() (*Dictionary, bool) {
	dictionary, ok := v.reference.(*Dictionary)
	return dictionary, v.kind == DictionaryKind && ok
}

// Function reports the callable payload, if v is a function.
func (v Value) Function() (Callable, bool) {
	function, ok := v.reference.(Callable)
	return function, v.kind == FunctionKind && ok
}

// Module reports the read-only namespace payload, if v is a module.
func (v Value) Module() (Module, bool) {
	module, ok := v.reference.(Module)
	return module, v.kind == ModuleKind && ok
}

// ErrorValue reports the runtime error payload, if v is an error.
func (v Value) ErrorValue() (ErrorValue, bool) {
	runtimeError, ok := v.reference.(ErrorValue)
	return runtimeError, v.kind == ErrorKind && ok
}

// Equal compares scalar values and compares reference values by identity.
func Equal(left, right Value) bool {
	if left.kind != right.kind {
		return false
	}
	switch left.kind {
	case NumberKind:
		return left.number == right.number
	case TextKind:
		return left.text == right.text
	case BooleanKind:
		return left.boolean == right.boolean
	case NullKind:
		return true
	case ListKind, DictionaryKind, FunctionKind, ModuleKind, ErrorKind:
		return left.identity != nil && left.identity == right.identity
	default:
		return false
	}
}

// TypeLabel returns the language-facing Spanish type name.
func TypeLabel(value Value) string {
	switch value.kind {
	case NumberKind:
		return "numero"
	case TextKind:
		return "texto"
	case BooleanKind:
		return "booleano"
	case NullKind:
		return "nulo"
	case ListKind:
		return "lista"
	case DictionaryKind:
		return "diccionario"
	case FunctionKind:
		return "funcion"
	case ModuleKind:
		return "modulo"
	case ErrorKind:
		return "error"
	default:
		return "desconocido"
	}
}

// List is a mutable ordered collection. Values returned from Snapshot are a
// shallow copy, preserving any nested collection references.
type List struct {
	values []Value
}

// Len returns the number of elements.
func (list *List) Len() int {
	if list == nil {
		return 0
	}
	return len(list.values)
}

// Get returns the value at index, or false when index is out of range.
func (list *List) Get(index int) (Value, bool) {
	if list == nil || index < 0 || index >= len(list.values) {
		return Value{}, false
	}
	return list.values[index], true
}

// Set updates an existing element and reports whether index was in range.
func (list *List) Set(index int, value Value) bool {
	if list == nil || index < 0 || index >= len(list.values) {
		return false
	}
	list.values[index] = value
	return true
}

// Append adds a value to the end of the list.
func (list *List) Append(value Value) {
	if list != nil {
		list.values = append(list.values, value)
	}
}

// Pop removes and returns the last value. It reports false for an empty or nil list.
func (list *List) Pop() (Value, bool) {
	if list == nil || len(list.values) == 0 {
		return Value{}, false
	}
	last := len(list.values) - 1
	removed := list.values[last]
	list.values[last] = Value{}
	list.values = list.values[:last]
	return removed, true
}

// Snapshot returns a shallow copy in list order.
func (list *List) Snapshot() []Value {
	if list == nil {
		return nil
	}
	return append([]Value(nil), list.values...)
}

// Dictionary is a mutable string-keyed collection that preserves insertion
// order. Updating an existing key does not change its position.
type Dictionary struct {
	entries   []dictionaryEntry
	positions map[string]int
}

type dictionaryEntry struct {
	key   string
	value Value
}

// Len returns the number of keys.
func (dictionary *Dictionary) Len() int {
	if dictionary == nil {
		return 0
	}
	return len(dictionary.entries)
}

// Get returns the value for key, or false when key is absent.
func (dictionary *Dictionary) Get(key string) (Value, bool) {
	if dictionary == nil {
		return Value{}, false
	}
	position, ok := dictionary.positions[key]
	if !ok {
		return Value{}, false
	}
	return dictionary.entries[position].value, true
}

// Set inserts or updates a key. Updating preserves the original insertion order.
func (dictionary *Dictionary) Set(key string, value Value) {
	if dictionary == nil {
		return
	}
	if dictionary.positions == nil {
		dictionary.positions = make(map[string]int)
	}
	if position, exists := dictionary.positions[key]; exists {
		dictionary.entries[position].value = value
		return
	}
	dictionary.positions[key] = len(dictionary.entries)
	dictionary.entries = append(dictionary.entries, dictionaryEntry{key: key, value: value})
}

// KeysSnapshot returns the keys in insertion order.
func (dictionary *Dictionary) KeysSnapshot() []string {
	if dictionary == nil {
		return nil
	}
	keys := make([]string, len(dictionary.entries))
	for index, entry := range dictionary.entries {
		keys[index] = entry.key
	}
	return keys
}

// EntriesSnapshot returns the key/value pairs in insertion order.
func (dictionary *Dictionary) EntriesSnapshot() []Entry {
	if dictionary == nil {
		return nil
	}
	entries := make([]Entry, len(dictionary.entries))
	for index, entry := range dictionary.entries {
		entries[index] = Entry{Key: entry.key, Value: entry.value}
	}
	return entries
}

// Entry is one insertion-ordered dictionary key/value pair.
type Entry struct {
	Key   string
	Value Value
}

// TextLength counts Unicode scalar values, not UTF-8 bytes or grapheme clusters.
func TextLength(text string) int { return utf8.RuneCountInString(text) }

// TextAt returns the Unicode scalar at a zero-based non-negative index.
func TextAt(text string, index int) (string, bool) {
	if index < 0 {
		return "", false
	}
	current := 0
	for _, runeValue := range text {
		if current == index {
			return string(runeValue), true
		}
		current++
	}
	return "", false
}

// Display renders a value for human-readable output. Collection cycles are
// represented as the quoted text "<ciclo>" at the point where the cycle recurs.
func Display(value Value) string {
	return display(value, false, make(map[*identity]bool))
}

func display(value Value, nested bool, active map[*identity]bool) string {
	switch value.kind {
	case NumberKind:
		if value.number == 0 {
			return "0"
		}
		return strconv.FormatFloat(value.number, 'g', -1, 64)
	case TextKind:
		if nested {
			return strconv.Quote(value.text)
		}
		return value.text
	case BooleanKind:
		if value.boolean {
			return "gol"
		}
		return "fuera"
	case NullKind:
		return "nada"
	case ListKind:
		if active[value.identity] {
			return cycleMarker(nested)
		}
		list, ok := value.List()
		if !ok {
			return "[]"
		}
		active[value.identity] = true
		defer delete(active, value.identity)
		items := list.Snapshot()
		parts := make([]string, len(items))
		for index, item := range items {
			parts[index] = display(item, true, active)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case DictionaryKind:
		if active[value.identity] {
			return cycleMarker(nested)
		}
		dictionary, ok := value.Dictionary()
		if !ok {
			return "{}"
		}
		active[value.identity] = true
		defer delete(active, value.identity)
		entries := dictionary.EntriesSnapshot()
		parts := make([]string, len(entries))
		for index, entry := range entries {
			parts[index] = strconv.Quote(entry.Key) + ": " + display(entry.Value, true, active)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case FunctionKind:
		return "<funcion>"
	case ModuleKind:
		return "<modulo>"
	case ErrorKind:
		if runtimeError, ok := value.ErrorValue(); ok {
			return "<error " + runtimeError.Code() + ">"
		}
		return "<error>"
	default:
		return "<invalido>"
	}
}

func cycleMarker(nested bool) string {
	if nested {
		return strconv.Quote("<ciclo>")
	}
	return "<ciclo>"
}
