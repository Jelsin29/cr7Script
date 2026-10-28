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

func n(x float64) value.Value { v, _ := value.NewNumber(x); return v }
func s(x string) value.Value  { return value.NewText(x) }
func call(t *testing.T, l *stdlib.Library, name string, args ...value.Value) (value.Value, error) {
	t.Helper()
	var v value.Value
	if parts := strings.Split(name, "."); len(parts) == 2 {
		m, ok := l.Module(parts[0])
		if !ok {
			t.Fatal(name)
		}
		v, ok = m.Get(parts[1])
		if !ok {
			t.Fatal(name)
		}
	} else {
		var ok bool
		v, ok = l.Prelude()[name]
		if !ok {
			t.Fatal(name)
		}
	}
	f, ok := v.Function()
	if !ok {
		t.Fatal(name)
	}
	return f.Call(args)
}
func check(t *testing.T, l *stdlib.Library, name, want string, args ...value.Value) value.Value {
	t.Helper()
	v, err := call(t, l, name, args...)
	if err != nil || value.Display(v) != want {
		t.Fatalf("%s: got %s, %v; want %s", name, value.Display(v), err, want)
	}
	return v
}
func code(t *testing.T, err error, want string) {
	t.Helper()
	var e interface{ Code() string }
	if !errors.As(err, &e) || e.Code() != want {
		t.Fatalf("got %v, want %s", err, want)
	}
}

func TestContracts(t *testing.T) {
	l := stdlib.New(stdlib.Options{})
	tests := []struct {
		name, want string
		args       []value.Value
	}{
		{"el_bicho", "ÉΣß", []value.Value{s("éσß")}}, {"texto", "gol", []value.Value{value.NewBoolean(true)}}, {"tipo", "nulo", []value.Value{value.Null()}},
		{"numero", "-125", []value.Value{s(" \t-1.25e2\r\n")}}, {"palmares", "3", []value.Value{s("a⚽b")}}, {"palmares", "0", []value.Value{value.NewDictionary()}},
		{"rango", "[-2, -1, 0, 1]", []value.Value{n(-2), n(2)}}, {"rango", "[]", []value.Value{n(2), n(2)}},
		{"chilena", "[3, 2, 1]", []value.Value{value.NewList(n(1), n(2), n(3))}}, {"hat_trick", "[7, 7, 7]", []value.Value{n(7)}},
		{"contiene", "gol", []value.Value{s("a⚽b"), s("⚽")}}, {"contiene", "fuera", []value.Value{value.NewList(n(1)), value.NewBoolean(true)}},
		{"exige", "nada", []value.Value{value.NewBoolean(true), s("ok")}},
		{"mr_champions.soy_el_mejor", "-1", []value.Value{value.NewList(n(-3), n(-1))}}, {"mr_champions.minimo", "-3", []value.Value{value.NewList(n(-3), n(-1))}},
		{"mr_champions.goles_totales", "0", []value.Value{value.NewList()}}, {"mr_champions.promedio", "2", []value.Value{value.NewList(n(1), n(3))}},
		{"relato.minusculas", "éσ", []value.Value{s("ÉΣ")}}, {"relato.separa", "[\"\", \"a\", \"\"]", []value.Value{s(",a,"), s(",")}},
		{"relato.une", "a⚽b", []value.Value{value.NewList(s("a"), s("b")), s("⚽")}}, {"relato.une", "", []value.Value{value.NewList(), s(",")}},
		{"relato.reemplaza", "xxa", []value.Value{s("aaaaa"), s("aa"), s("x")}}, {"marcador.absoluto", "2", []value.Value{n(-2)}},
		{"marcador.redondea", "-3", []value.Value{n(-2.5)}}, {"marcador.potencia", "8", []value.Value{n(2), n(3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name+tt.want, func(t *testing.T) { check(t, l, tt.name, tt.want, tt.args...) })
	}
	if len(l.Prelude()) != 17 {
		t.Fatal("prelude must have exactly 17 functions")
	}
}

func TestErrorsAndRejectedMutation(t *testing.T) {
	l := stdlib.New(stdlib.Options{})
	tests := []struct {
		name, code string
		args       []value.Value
	}{
		{"numero", "ConversionInvalida", []value.Value{s(".5")}}, {"numero", "ConversionInvalida", []value.Value{s("0x10")}}, {"numero", "ConversionInvalida", []value.Value{s("\u00a01")}}, {"numero", "NumeroNoFinito", []value.Value{s("1e999")}},
		{"rango", "RangoInvalido", []value.Value{n(2), n(1)}}, {"rango", "RangoInvalido", []value.Value{n(0.5), n(1)}}, {"rango", "RangoInvalido", []value.Value{n(0), n(9007199254740992)}},
		{"calma", "TipoIncorrecto", []value.Value{n(-1)}}, {"calma", "TipoIncorrecto", []value.Value{n(0.5)}},
		{"banquillo", "ListaVacia", []value.Value{value.NewList()}}, {"convoca", "TipoIncorrecto", []value.Value{s("x"), n(1)}},
		{"consulta", "TipoIncorrecto", []value.Value{value.NewDictionary(), n(1), n(2)}}, {"contiene", "TipoIncorrecto", []value.Value{s("x"), n(1)}},
		{"contiene", "TipoIncorrecto", []value.Value{value.NewDictionary(), n(1)}}, {"exige", "FaltaDelUsuario", []value.Value{value.NewBoolean(false), s("failure")}},
		{"mr_champions.soy_el_mejor", "ListaVacia", []value.Value{value.NewList()}}, {"mr_champions.minimo", "ListaVacia", []value.Value{value.NewList()}}, {"mr_champions.promedio", "ListaVacia", []value.Value{value.NewList()}},
		{"mr_champions.goles_totales", "TipoIncorrecto", []value.Value{value.NewList(n(1), s("x"))}}, {"mr_champions.goles_totales", "NumeroNoFinito", []value.Value{value.NewList(n(math.MaxFloat64), n(math.MaxFloat64))}},
		{"mr_champions.pichichi", "ListaVacia", []value.Value{value.NewList()}}, {"mr_champions.pichichi", "ClaveAusente", []value.Value{value.NewList(value.NewDictionary())}}, {"mr_champions.pichichi", "TipoIncorrecto", []value.Value{value.NewList(n(1))}},
		{"relato.separa", "TipoIncorrecto", []value.Value{s("x"), s("")}}, {"relato.une", "TipoIncorrecto", []value.Value{value.NewList(n(1)), s(",")}}, {"relato.une", "TipoIncorrecto", []value.Value{value.NewList(), s("")}},
		{"relato.reemplaza", "TipoIncorrecto", []value.Value{s("x"), s(""), s("y")}}, {"marcador.potencia", "NumeroNoFinito", []value.Value{n(-1), n(0.5)}},
	}
	for _, tt := range tests {
		t.Run(tt.name+tt.code, func(t *testing.T) {
			before := make([]string, len(tt.args))
			for i, v := range tt.args {
				before[i] = value.Display(v)
			}
			_, err := call(t, l, tt.name, tt.args...)
			code(t, err, tt.code)
			for i, v := range tt.args {
				if value.Display(v) != before[i] {
					t.Fatal("rejected call mutated input")
				}
			}
		})
	}
	// Every fixed-arity function rejects both missing and excess arguments and wrong types.
	groups := map[string][]string{"": {"entrevista", "hala_madrid", "el_bicho", "texto", "numero", "tipo", "palmares", "calma", "rango", "chilena", "hat_trick", "convoca", "banquillo", "consulta", "contiene", "exige"}, "mr_champions.": {"soy_el_mejor", "minimo", "goles_totales", "promedio", "pichichi"}, "relato.": {"minusculas", "separa", "une", "reemplaza"}, "marcador.": {"absoluto", "redondea", "potencia"}}
	for prefix, names := range groups {
		for _, name := range names {
			name = prefix + name
			var f value.Callable
			if prefix == "" {
				f, _ = l.Prelude()[name].Function()
			} else {
				p := strings.Split(name, ".")
				m, _ := l.Module(p[0])
				v, _ := m.Get(p[1])
				f, _ = v.Function()
			}
			a := make([]value.Value, f.Arity()+1)
			_, err := f.Call(a)
			code(t, err, "AridadIncorrecta")
			if f.Arity() > 0 {
				_, err = f.Call(nil)
				code(t, err, "AridadIncorrecta")
			}
			if name != "texto" && name != "tipo" && name != "hat_trick" && f.Arity() > 0 {
				a = make([]value.Value, f.Arity())
				for i := range a {
					a[i] = value.Null()
				}
				_, err = f.Call(a)
				code(t, err, "TipoIncorrecto")
			}
		}
	}
}

func TestIdentityAndMutation(t *testing.T) {
	l := stdlib.New(stdlib.Options{})
	child := value.NewList(n(7))
	original := value.NewList(n(1), child)
	reversed := check(t, l, "chilena", "[[7], 1]", original)
	if value.Equal(original, reversed) || value.Display(original) != "[1, [7]]" {
		t.Fatal("reversal mutated or reused original")
	}
	r, _ := reversed.List()
	got, _ := r.Get(0)
	if !value.Equal(got, child) {
		t.Fatal("not shallow")
	}
	triple := check(t, l, "hat_trick", "[[7], [7], [7]]", child)
	tr, _ := triple.List()
	for _, v := range tr.Snapshot() {
		if !value.Equal(v, child) {
			t.Fatal("identity lost")
		}
	}
	check(t, l, "convoca", "nada", original, n(9))
	check(t, l, "banquillo", "9", original)
	if value.Display(original) != "[1, [7]]" {
		t.Fatal("pop did not mutate alias")
	}
	d := value.NewDictionary()
	dict, _ := d.Dictionary()
	dict.Set("present", value.Null())
	check(t, l, "consulta", "nada", d, s("present"), n(7))
	check(t, l, "consulta", "7", d, s("missing"), n(7))
	check(t, l, "contiene", "gol", d, s("present"))
	check(t, l, "contiene", "gol", value.NewList(child), child)
	check(t, l, "contiene", "fuera", value.NewList(child), value.NewList(n(7)))
	a, b := value.NewDictionary(), value.NewDictionary()
	ad, _ := a.Dictionary()
	bd, _ := b.Dictionary()
	ad.Set("goles", n(7))
	bd.Set("goles", n(7))
	winner := check(t, l, "mr_champions.pichichi", "{\"goles\": 7}", value.NewList(a, b))
	if !value.Equal(winner, a) {
		t.Fatal("first tie identity lost")
	}
}

type flushBuffer struct {
	bytes.Buffer
	flushed int
}

func (b *flushBuffer) Flush() error { b.flushed++; return nil }

type badIO struct{}

func (badIO) Read([]byte) (int, error)  { return 0, errors.New("read failure") }
func (badIO) Write([]byte) (int, error) { return 0, errors.New("write failure") }
func TestIOAndSleep(t *testing.T) {
	var out flushBuffer
	var slept []time.Duration
	l := stdlib.New(stdlib.Options{Input: strings.NewReader(" one \r\ntwo\nlast"), Output: &out, Sleep: func(d time.Duration) { slept = append(slept, d) }})
	check(t, l, "siuuu", "nada", s("Gol"), n(7), value.NewBoolean(true))
	check(t, l, "siuuu", "nada")
	check(t, l, "hala_madrid", "nada")
	for _, want := range []string{" one ", "two", "last", "nada", "nada"} {
		check(t, l, "entrevista", want, s(">"))
	}
	if out.String() != "Gol 7 gol\n\nUno, dos, tres... ¡Hala Madrid!\n>>>>>" || out.flushed != 5 {
		t.Fatalf("output %q flushes %d", out.String(), out.flushed)
	}
	check(t, l, "calma", "nada", n(7))
	check(t, l, "calma", "nada", n(0))
	if len(slept) != 2 || slept[0] != 7*time.Millisecond || slept[1] != 0 {
		t.Fatal(slept)
	}
	for _, name := range []string{"siuuu", "hala_madrid", "entrevista"} {
		l := stdlib.New(stdlib.Options{Input: badIO{}, Output: badIO{}})
		args := []value.Value{}
		if name == "entrevista" {
			args = append(args, s("prompt"))
		}
		_, err := call(t, l, name, args...)
		code(t, err, "EntradaSalida")
	}
	l = stdlib.New(stdlib.Options{Input: badIO{}, Output: io.Discard})
	_, err := call(t, l, "entrevista", s(""))
	code(t, err, "EntradaSalida")
}

func TestExactQuotesAndSilentModules(t *testing.T) {
	var out bytes.Buffer
	l := stdlib.New(stdlib.Options{Output: &out})
	quotes := map[string]string{"RICO_GUAPO_Y_GRAN_JUGADOR": "soy guapo, rico y un gran jugador", "ESTOY_TRISTE": "No lo hago cuando estoy triste", "TODOS_A_MI_NIVEL": "Si todos estuviesen en mi nivel seríamos los primeros", "CINCO_CHAMPIONS": "Cinco Champions tengo yo y el Atleti, cero", "THE_BEST_SOY_YO": "el premio The Best fue para el 'the best', que soy yo", "SOY_EL_MEJOR": "Soy el mejor jugador de la historia", "HALA_MADRID": "Uno, dos, tres: ¡Hala Madrid!"}
	for _, name := range []string{"mr_champions", "relato", "marcador", "rueda_de_prensa"} {
		if _, ok := l.Module(name); !ok {
			t.Fatal(name)
		}
	}
	m, _ := l.Module("rueda_de_prensa")
	for name, want := range quotes {
		t.Run(name, func(t *testing.T) {
			v, ok := m.Get(name)
			got, text := v.Text()
			if !ok || !text || got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	}
	for _, name := range []string{"losh_dosh", "los_dos", "missing"} {
		if _, ok := m.Get(name); ok {
			t.Fatal("unrequested export")
		}
		if _, ok := l.Module(name); ok {
			t.Fatal("unrequested module")
		}
	}
	if out.Len() != 0 {
		t.Fatal("imports printed output")
	}
	p := l.Prelude()
	delete(p, "siuuu")
	if len(l.Prelude()) != 17 {
		t.Fatal("prelude map leaked mutation")
	}
}
