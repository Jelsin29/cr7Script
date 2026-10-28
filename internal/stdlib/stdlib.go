// Package stdlib provides CR7Script's native prelude and registered modules.
package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jelsin29/cr7Script/internal/evaluator"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/value"
)

// Options supplies explicit I/O and timing dependencies. Nil fields use the process defaults.
type Options struct {
	Input  io.Reader
	Output io.Writer
	Sleep  func(time.Duration)
}

// Library owns native functions and read-only standard module namespaces.
type Library struct {
	prelude map[string]value.Value
	modules map[string]value.Module
}
type namespace struct{ exports map[string]value.Value }

func (m *namespace) Get(name string) (value.Value, bool) { v, ok := m.exports[name]; return v, ok }

// Prelude returns a fresh map containing the 17 core functions.
func (l *Library) Prelude() map[string]value.Value {
	p := make(map[string]value.Value, len(l.prelude))
	for k, v := range l.prelude {
		p[k] = v
	}
	return p
}

// Module resolves only registered standard names and never performs I/O.
func (l *Library) Module(name string) (value.Module, bool) { m, ok := l.modules[name]; return m, ok }

type native struct {
	name  string
	arity int
	run   func([]value.Value) (value.Value, error)
}

func (f *native) Name() string { return f.name }

// Arity returns -1 for the variadic siuuu function.
func (f *native) Arity() int { return f.arity }
func (f *native) Call(args []value.Value) (value.Value, error) {
	if f.arity >= 0 && len(args) != f.arity {
		return fail("AridadIncorrecta", fmt.Sprintf("%s expects %d arguments, received %d", f.name, f.arity, len(args)))
	}
	return f.run(args)
}
func function(name string, arity int, run func([]value.Value) (value.Value, error)) value.Value {
	return value.NewFunction(&native{name, arity, run})
}
func fail(code, message string) (value.Value, error) {
	return value.Value{}, evaluator.NewRuntimeError(code, message, lexer.Position{})
}
func wrong() (value.Value, error) { return fail("TipoIncorrecto", "invalid argument type or value") }
func number(n float64) (value.Value, error) {
	v, err := value.NewNumber(n)
	if err != nil {
		return fail("NumeroNoFinito", "numeric result must be finite")
	}
	return v, nil
}
func safeInteger(n float64) bool { return math.Abs(n) <= 9007199254740991 && math.Trunc(n) == n }

var decimal = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// New creates an isolated library without reading, printing, or sleeping.
func New(options Options) *Library {
	if options.Input == nil {
		options.Input = os.Stdin
	}
	if options.Output == nil {
		options.Output = os.Stdout
	}
	if options.Sleep == nil {
		options.Sleep = time.Sleep
	}
	l := &Library{prelude: make(map[string]value.Value), modules: make(map[string]value.Module)}
	add := func(name string, arity int, run func([]value.Value) (value.Value, error)) {
		l.prelude[name] = function(name, arity, run)
	}
	write := func(text string) (value.Value, error) {
		n, err := io.WriteString(options.Output, text)
		if err != nil || n != len(text) {
			return fail("EntradaSalida", "cannot write output")
		}
		return value.Null(), nil
	}
	add("siuuu", -1, func(a []value.Value) (value.Value, error) {
		parts := make([]string, len(a))
		for i, v := range a {
			parts[i] = value.Display(v)
		}
		return write(strings.Join(parts, " ") + "\n")
	})
	reader := bufio.NewReader(options.Input)
	add("entrevista", 1, func(a []value.Value) (value.Value, error) {
		prompt, ok := a[0].Text()
		if !ok {
			return wrong()
		}
		if _, err := write(prompt); err != nil {
			return value.Value{}, err
		}
		if f, ok := options.Output.(interface{ Flush() error }); ok {
			if err := f.Flush(); err != nil {
				return fail("EntradaSalida", "cannot flush output")
			}
		}
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fail("EntradaSalida", "cannot read input")
		}
		if err == io.EOF && len(line) == 0 {
			return value.Null(), nil
		}
		if strings.HasSuffix(line, "\n") {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
		}
		return value.NewText(line), nil
	})
	add("hala_madrid", 0, func([]value.Value) (value.Value, error) { return write("Uno, dos, tres... ¡Hala Madrid!\n") })
	add("el_bicho", 1, textUnary(strings.ToUpper))
	add("texto", 1, func(a []value.Value) (value.Value, error) { return value.NewText(value.Display(a[0])), nil })
	add("tipo", 1, func(a []value.Value) (value.Value, error) { return value.NewText(value.TypeLabel(a[0])), nil })
	add("numero", 1, func(a []value.Value) (value.Value, error) {
		s, ok := a[0].Text()
		if !ok {
			return wrong()
		}
		s = strings.Trim(s, " \t\n\r\v\f")
		if !decimal.MatchString(s) {
			return fail("ConversionInvalida", "invalid decimal number")
		}
		n, err := strconv.ParseFloat(s, 64)
		if math.IsInf(n, 0) || math.IsNaN(n) {
			return fail("NumeroNoFinito", "converted number must be finite")
		}
		if err != nil {
			return fail("ConversionInvalida", "invalid decimal number")
		}
		return number(n)
	})
	add("palmares", 1, func(a []value.Value) (value.Value, error) {
		if s, ok := a[0].Text(); ok {
			return number(float64(value.TextLength(s)))
		}
		if list, ok := a[0].List(); ok {
			return number(float64(list.Len()))
		}
		if d, ok := a[0].Dictionary(); ok {
			return number(float64(d.Len()))
		}
		return wrong()
	})
	add("calma", 1, func(a []value.Value) (value.Value, error) {
		ms, ok := a[0].Number()
		if !ok || !safeInteger(ms) || ms < 0 {
			return wrong()
		}
		// Chunk large millisecond counts so converting to Duration cannot overflow.
		const chunk int64 = int64(math.MaxInt64) / int64(time.Millisecond)
		for ms > float64(chunk) {
			options.Sleep(time.Duration(chunk) * time.Millisecond)
			ms -= float64(chunk)
		}
		options.Sleep(time.Duration(ms) * time.Millisecond)
		return value.Null(), nil
	})
	add("rango", 2, func(a []value.Value) (value.Value, error) {
		start, ok := a[0].Number()
		end, ok2 := a[1].Number()
		if !ok || !ok2 {
			return wrong()
		}
		if !safeInteger(start) || !safeInteger(end) || start > end {
			return fail("RangoInvalido", "range requires safe integer bounds in ascending order")
		}
		result := value.NewList()
		list, _ := result.List()
		for x := start; x < end; x++ {
			v, _ := number(x)
			list.Append(v)
		}
		return result, nil
	})
	add("chilena", 1, func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		if !ok {
			return wrong()
		}
		items := list.Snapshot()
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
		return value.NewList(items...), nil
	})
	add("hat_trick", 1, func(a []value.Value) (value.Value, error) { return value.NewList(a[0], a[0], a[0]), nil })
	add("convoca", 2, func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		if !ok {
			return wrong()
		}
		list.Append(a[1])
		return value.Null(), nil
	})
	add("banquillo", 1, func(a []value.Value) (value.Value, error) {
		list, ok := a[0].List()
		if !ok {
			return wrong()
		}
		v, ok := list.Pop()
		if !ok {
			return fail("ListaVacia", "cannot remove from an empty list")
		}
		return v, nil
	})
	add("consulta", 3, func(a []value.Value) (value.Value, error) {
		d, ok := a[0].Dictionary()
		key, ok2 := a[1].Text()
		if !ok || !ok2 {
			return wrong()
		}
		if v, found := d.Get(key); found {
			return v, nil
		}
		return a[2], nil
	})
	add("contiene", 2, func(a []value.Value) (value.Value, error) {
		if list, ok := a[0].List(); ok {
			for _, v := range list.Snapshot() {
				if value.Equal(v, a[1]) {
					return value.NewBoolean(true), nil
				}
			}
			return value.NewBoolean(false), nil
		}
		text, ok := a[1].Text()
		if !ok {
			return wrong()
		}
		if d, ok := a[0].Dictionary(); ok {
			_, found := d.Get(text)
			return value.NewBoolean(found), nil
		}
		if s, ok := a[0].Text(); ok {
			return value.NewBoolean(strings.Contains(s, text)), nil
		}
		return wrong()
	})
	add("exige", 2, func(a []value.Value) (value.Value, error) {
		condition, ok := a[0].Boolean()
		message, ok2 := a[1].Text()
		if !ok || !ok2 {
			return wrong()
		}
		if !condition {
			return fail("FaltaDelUsuario", message)
		}
		return value.Null(), nil
	})
	l.installModules()
	return l
}
func textUnary(transform func(string) string) func([]value.Value) (value.Value, error) {
	return func(a []value.Value) (value.Value, error) {
		s, ok := a[0].Text()
		if !ok {
			return wrong()
		}
		return value.NewText(transform(s)), nil
	}
}
