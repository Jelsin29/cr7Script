package stdlib_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Jelsin29/cr7Script/internal/stdlib"
	"github.com/Jelsin29/cr7Script/internal/value"
)

func TestNumericAndEmptyBoundaries(t *testing.T) {
	l := stdlib.New(stdlib.Options{})
	for _, tt := range []struct{ input, want string }{{"+7", "7"}, {"-0", "0"}, {"1E-2", "0.01"}, {"001.5", "1.5"}, {"1e-999", "0"}} {
		t.Run(tt.input, func(t *testing.T) { check(t, l, "numero", tt.want, s(tt.input)) })
	}
	for _, input := range []string{"", " ", "NaN", "Inf", "1.", "1_000", "1 2", "++1", "1e", "\u20037", "7\u00a0"} {
		t.Run("invalid "+input, func(t *testing.T) { _, err := call(t, l, "numero", s(input)); code(t, err, "ConversionInvalida") })
	}
	tests := []struct {
		name, want string
		args       []value.Value
	}{
		{"chilena", "[]", []value.Value{value.NewList()}},
		{"hat_trick", "[nada, nada, nada]", []value.Value{value.Null()}},
		{"palmares", "0", []value.Value{value.NewList()}},
		{"palmares", "2", []value.Value{s("e\u0301")}},
		{"relato.separa", "[\"\"]", []value.Value{s(""), s(",")}},
		{"relato.reemplaza", "", []value.Value{s(""), s("a"), s("b")}},
		{"relato.reemplaza", "", []value.Value{s("aa"), s("a"), s("")}},
		{"contiene", "gol", []value.Value{s("abc"), s("")}},
		{"contiene", "fuera", []value.Value{value.NewList(), n(1)}},
		{"contiene", "fuera", []value.Value{value.NewDictionary(), s("missing")}},
		{"mr_champions.goles_totales", "0", []value.Value{value.NewList(n(1e16), n(1), n(-1e16))}},
		{"marcador.redondea", "3", []value.Value{n(2.5)}},
		{"marcador.absoluto", "0", []value.Value{n(math.Copysign(0, -1))}},
	}
	for _, tt := range tests {
		t.Run(tt.name+tt.want, func(t *testing.T) { check(t, l, tt.name, tt.want, tt.args...) })
	}
	for _, start := range []float64{9007199254740990, -9007199254740991} {
		v, err := call(t, l, "rango", n(start), n(start+1))
		list, ok := v.List()
		if err != nil || !ok || list.Len() != 1 {
			t.Fatalf("boundary range: %s %v", value.Display(v), err)
		}
		item, _ := list.Get(0)
		if !value.Equal(item, n(start)) {
			t.Fatal("range changed numeric boundary")
		}
	}
	for _, name := range []string{"mr_champions.soy_el_mejor", "mr_champions.minimo", "mr_champions.promedio"} {
		_, err := call(t, l, name, value.NewList(n(1), s("bad")))
		code(t, err, "TipoIncorrecto")
	}
	_, err := call(t, l, "mr_champions.promedio", value.NewList(n(math.MaxFloat64), n(math.MaxFloat64)))
	code(t, err, "NumeroNoFinito")
	_, err = call(t, l, "marcador.potencia", n(2), n(1024))
	code(t, err, "NumeroNoFinito")
	_, err = call(t, l, "calma", n(9007199254740992))
	code(t, err, "TipoIncorrecto")
	_, err = call(t, l, "exige", value.NewBoolean(false), s("exact failure"))
	code(t, err, "FaltaDelUsuario")
	var message interface{ Message() string }
	if !errors.As(err, &message) || message.Message() != "exact failure" {
		t.Fatalf("message: %v", err)
	}
}

func TestPichichiValidatesAllPlayers(t *testing.T) {
	l := stdlib.New(stdlib.Options{})
	player := func(goals value.Value) value.Value {
		v := value.NewDictionary()
		d, _ := v.Dictionary()
		d.Set("goles", goals)
		return v
	}
	a, b := player(n(-7)), player(n(-2))
	winner, err := call(t, l, "mr_champions.pichichi", value.NewList(a, b))
	if err != nil || !value.Equal(winner, b) {
		t.Fatalf("winner %v error %v", winner, err)
	}
	for _, tt := range []struct {
		item value.Value
		code string
	}{{player(s("7")), "TipoIncorrecto"}, {value.NewDictionary(), "ClaveAusente"}, {n(7), "TipoIncorrecto"}} {
		_, err := call(t, l, "mr_champions.pichichi", value.NewList(b, tt.item))
		code(t, err, tt.code)
	}
}

type flushFailure struct {
	bytes.Buffer
	reads int
}

func (f *flushFailure) Flush() error { return errors.New("flush failure") }

type countingReader struct{ reads int }

func (r *countingReader) Read([]byte) (int, error) { r.reads++; return 0, io.EOF }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

func TestInputEdgesAndFailures(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  []value.Value
	}{
		{"", []value.Value{value.Null()}},
		{"\n\r\n", []value.Value{s(""), s(""), value.Null()}},
		{" \tline\r", []value.Value{s(" \tline\r"), value.Null()}},
		{"a\r\r\nb\rc\n", []value.Value{s("a\r"), s("b\rc"), value.Null()}},
	} {
		t.Run(tt.input, func(t *testing.T) {
			var out bytes.Buffer
			l := stdlib.New(stdlib.Options{Input: strings.NewReader(tt.input), Output: &out})
			for _, want := range tt.want {
				got, err := call(t, l, "entrevista", s("?"))
				if err != nil || !value.Equal(got, want) {
					t.Fatalf("got %s %v, want %s", value.Display(got), err, value.Display(want))
				}
			}
			if out.String() != strings.Repeat("?", len(tt.want)) {
				t.Fatal(out.String())
			}
		})
	}
	reader := &countingReader{}
	output := &flushFailure{}
	l := stdlib.New(stdlib.Options{Input: reader, Output: output})
	_, err := call(t, l, "entrevista", s("prompt"))
	code(t, err, "EntradaSalida")
	if reader.reads != 0 || output.String() != "prompt" {
		t.Fatal("read after failed flush")
	}
	l = stdlib.New(stdlib.Options{Input: reader, Output: shortWriter{}})
	_, err = call(t, l, "entrevista", s("prompt"))
	code(t, err, "EntradaSalida")
	if reader.reads != 0 {
		t.Fatal("read after failed write")
	}
	var out bytes.Buffer
	var sleeps int
	l = stdlib.New(stdlib.Options{Input: reader, Output: &out, Sleep: func(time.Duration) { sleeps++ }})
	_, err = call(t, l, "entrevista", n(1))
	code(t, err, "TipoIncorrecto")
	_, err = call(t, l, "calma", n(-1))
	code(t, err, "TipoIncorrecto")
	if out.Len() != 0 || reader.reads != 0 || sleeps != 0 {
		t.Fatal("rejected input caused side effects")
	}
}

func TestLongSleepDoesNotOverflowDuration(t *testing.T) {
	const millis = 9007199254740991
	var total int64
	l := stdlib.New(stdlib.Options{Sleep: func(d time.Duration) {
		if d < 0 || d%time.Millisecond != 0 {
			t.Fatalf("invalid duration %v", d)
		}
		total += int64(d / time.Millisecond)
	}})
	check(t, l, "calma", "nada", n(millis))
	if total != millis {
		t.Fatalf("slept %d milliseconds, want %d", total, millis)
	}
}
