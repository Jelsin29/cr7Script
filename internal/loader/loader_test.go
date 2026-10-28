package loader_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/evaluator"
	"github.com/Jelsin29/cr7Script/internal/loader"
	"github.com/Jelsin29/cr7Script/internal/value"
)

type library struct {
	events  []string
	lookups []string
}

func (l *library) Prelude() map[string]value.Value {
	return map[string]value.Value{"siuuu": value.NewFunction(&record{library: l})}
}
func (l *library) Module(name string) (value.Module, bool) {
	l.lookups = append(l.lookups, name)
	return exports{"label": value.NewText("standard")}, name == "relato"
}

type exports map[string]value.Value

func (e exports) Get(name string) (value.Value, bool) { v, ok := e[name]; return v, ok }

type record struct{ library *library }

func (*record) Name() string { return "siuuu" }
func (*record) Arity() int   { return 1 }
func (r *record) Call(args []value.Value) (value.Value, error) {
	text, _ := args[0].Text()
	r.library.events = append(r.library.events, text)
	return args[0], nil
}
func fixture(t *testing.T, root, name, source string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func exported(t *testing.T, i *evaluator.Interpreter, name string) value.Value {
	t.Helper()
	m, ok := i.Exports()
	if !ok {
		t.Fatal("no exports")
	}
	v, ok := m.Get(name)
	if !ok {
		t.Fatalf("missing export %s", name)
	}
	return v
}
func textExport(t *testing.T, i *evaluator.Interpreter, name, want string) {
	t.Helper()
	v := exported(t, i, name)
	got, ok := v.Text()
	if !ok || got != want {
		t.Fatalf("%s = %v, want %q", name, v, want)
	}
}
func requireLoadError(t *testing.T, i *evaluator.Interpreter, err error, code string) {
	t.Helper()
	if i != nil || err == nil {
		t.Fatalf("Load = %v, %v; want nil and error", i, err)
	}
	var runtimeError *evaluator.RuntimeError
	if errors.As(err, &runtimeError) {
		t.Fatalf("loader exposed runtime error: %v", err)
	}
	if !strings.Contains(err.Error(), code) {
		t.Fatalf("error %q lacks %s", err, code)
	}
}

func TestRelativeDiamondCanonicalCacheAndInitialization(t *testing.T) {
	root := t.TempDir()
	lib := &library{}
	l := loader.New(lib)
	fixture(t, root, "shared.cr7", `leyenda items = [7]; leyenda initialized = siuuu("shared"); jugada comandante(){ siuuu("unwanted dependency entry"); }`)
	fixture(t, root, "nested/left.cr7", `fichar "../shared.cr7" como shared; leyenda items = shared.items; leyenda initialized = siuuu("left");`)
	fixture(t, root, "right.cr7", `fichar "./nested/../shared.cr7" como shared; leyenda items = shared.items; leyenda initialized = siuuu("right");`)
	entry := fixture(t, root, "main.cr7", `fichar "./nested/left.cr7" como left; fichar "./right.cr7" como right; leyenda same = left.items == right.items; leyenda initialized = siuuu("main"); jugada comandante(){ siuuu("unwanted entry"); }`)
	i, err := l.Load(entry)
	if err != nil {
		t.Fatal(err)
	}
	same, ok := exported(t, i, "same").Boolean()
	if !ok || !same {
		t.Fatal("diamond did not share collection identity")
	}
	if !reflect.DeepEqual(lib.events, []string{"shared", "left", "right", "main"}) {
		t.Fatalf("initialization = %v", lib.events)
	}
	again, err := l.Load(filepath.Join(root, "nested", "..", "main.cr7"))
	if err != nil || again != i {
		t.Fatalf("cached Load = %p, %v; want %p", again, err, i)
	}
	if len(lib.events) != 4 {
		t.Fatalf("reinitialized: %v", lib.events)
	}
	other, err := loader.New(lib).Load(entry)
	if err != nil || other == i || len(lib.events) != 8 {
		t.Fatalf("separate loader did not initialize independently: %p %v %v", other, err, lib.events)
	}
}

func TestSymlinkIdentityAndImporterBase(t *testing.T) {
	root := t.TempDir()
	lib := &library{}
	l := loader.New(lib)
	target := fixture(t, root, "real/module.cr7", `fichar "./child.cr7" como child; leyenda name = child.name; leyenda initialized = siuuu("module");`)
	fixture(t, root, "real/child.cr7", `leyenda name = "real child";`)
	alias := filepath.Join(root, "alias.cr7")
	if err := os.Symlink(target, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	i, err := l.Load(alias)
	if err != nil {
		t.Fatal(err)
	}
	textExport(t, i, "name", "real child")
	again, err := l.Load(target)
	if err != nil || again != i || !reflect.DeepEqual(lib.events, []string{"module"}) {
		t.Fatalf("symlink cache = %p %v %v", again, err, lib.events)
	}
}

func TestStandardModulesExportsAndReadOnlyAliases(t *testing.T) {
	root := t.TempDir()
	lib := &library{}
	l := loader.New(lib)
	fixture(t, root, "dep.cr7", `fichar "relato" como words; leyenda label = words.label; leyenda items = [7]; jugada change(){ items[0] = 10; } jugada invalid(){ words = nada; }`)
	// Invalid alias reassignment is a static failure, even in an uncalled function.
	i, err := l.Load(filepath.Join(root, "dep.cr7"))
	requireLoadError(t, i, err, "ReasignacionDeLeyenda")
	fixture(t, root, "good.cr7", `fichar "relato" como words; leyenda label = words.label; leyenda items = [7]; jugada change(){ items[0] = 10; }`)
	entry := fixture(t, root, "main.cr7", `fichar "./good.cr7" como dep; fichar "relato" como first; fichar "relato" como second; leyenda label = dep.label; leyenda same = first == second; leyenda changed = dep.change(); leyenda item = dep.items[0];`)
	i, err = l.Load(entry)
	if err != nil {
		t.Fatal(err)
	}
	textExport(t, i, "label", "standard")
	if !reflect.DeepEqual(lib.lookups, []string{"relato"}) {
		t.Fatalf("standard registry lookups = %v; want one cached lookup", lib.lookups)
	}
	same, _ := exported(t, i, "same").Boolean()
	if !same {
		t.Fatal("standard namespace identity not cached")
	}
	n, _ := exported(t, i, "item").Number()
	if n != 10 {
		t.Fatalf("collection mutation = %v", n)
	}
	m, _ := i.Exports()
	for _, name := range []string{"dep", "first", "siuuu"} {
		if _, ok := m.Get(name); ok {
			t.Fatalf("export leaked %s", name)
		}
	}
	good, err := l.Load(filepath.Join(root, "good.cr7"))
	if err != nil {
		t.Fatal(err)
	}
	gm, _ := good.Exports()
	if _, ok := gm.Get("words"); ok {
		t.Fatal("dependency leaked import")
	}
	if _, ok := gm.Get("change"); !ok {
		t.Fatal("missing function export")
	}
}

func TestMissingAndInvalidImports(t *testing.T) {
	for _, path := range []string{"unknown", "https://example.invalid/module.cr7", "module.cr7", "./missing.cr7", "./missing.txt", "../missing.cr7", "", "/absolute.cr7"} {
		t.Run(path, func(t *testing.T) {
			root := t.TempDir()
			lib := &library{}
			entry := fixture(t, root, "main.cr7", `fichar "`+path+`" como missing; leyenda initialized = siuuu("bad");`)
			i, err := loader.New(lib).Load(entry)
			requireLoadError(t, i, err, "ModuloNoEncontrado")
			if len(lib.events) != 0 {
				t.Fatalf("rejected input initialized: %v", lib.events)
			}
		})
	}
	root := t.TempDir()
	i, err := loader.New(&library{}).Load(filepath.Join(root, "missing.cr7"))
	requireLoadError(t, i, err, "ModuloNoEncontrado")
}

func TestImportChainAndCycles(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "cycle"}[cycle], func(t *testing.T) {
			root := t.TempDir()
			lib := &library{}
			a := fixture(t, root, "a.cr7", `fichar "./b.cr7" como b; leyenda initialized = siuuu("a");`)
			bsource := `fichar "./missing.cr7" como missing;`
			code := "ModuloNoEncontrado"
			if cycle {
				bsource = `fichar "./a.cr7" como a;`
				code = "FichajeCircular"
			}
			fixture(t, root, "b.cr7", bsource)
			i, err := loader.New(lib).Load(a)
			requireLoadError(t, i, err, code)
			for _, name := range []string{"a.cr7", "b.cr7"} {
				if !strings.Contains(err.Error(), name) {
					t.Fatalf("missing chain member %s: %v", name, err)
				}
			}
			if cycle && strings.Count(err.Error(), "a.cr7") < 2 {
				t.Fatalf("cycle chain not closed: %v", err)
			}
			if len(lib.events) != 0 {
				t.Fatalf("initialized cyclic graph: %v", lib.events)
			}
		})
	}
}

func TestRejectInvalidModuleSyntaxBeforeInitialization(t *testing.T) {
	for _, source := range []string{`siuuu("bad");`, `cantera x = 7;`, `leyenda x;`, `leyenda x = 7; fichar "relato" como words;`, `jugada f(){ fichar "relato" como words; }`} {
		t.Run(source, func(t *testing.T) {
			lib := &library{}
			entry := fixture(t, t.TempDir(), "main.cr7", source)
			i, err := loader.New(lib).Load(entry)
			requireLoadError(t, i, err, "SintaxisInvalida")
			if len(lib.events) != 0 {
				t.Fatalf("invalid module effects = %v", lib.events)
			}
		})
	}
}

func TestInitializationFailuresAreNotRuntimeErrorsAndDoNotRepeat(t *testing.T) {
	root := t.TempDir()
	lib := &library{}
	l := loader.New(lib)
	entry := fixture(t, root, "main.cr7", `leyenda effect = siuuu("once"); leyenda broken = 1 / 0;`)
	for range 2 {
		i, err := l.Load(entry)
		requireLoadError(t, i, err, "DivisionPorCero")
	}
	if !reflect.DeepEqual(lib.events, []string{"once"}) {
		t.Fatalf("failure repeated effects: %v", lib.events)
	}
}
