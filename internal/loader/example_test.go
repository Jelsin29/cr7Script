package loader_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jelsin29/cr7Script/internal/loader"
	"github.com/Jelsin29/cr7Script/internal/value"
)

func TestReferenceLocalFunctionExample(t *testing.T) {
	root := t.TempDir()
	fixture(t, root, "jugadas.cr7", `jugada rematar(potencia, precision) { devuelve potencia * precision; }`)
	entry := fixture(t, root, "partido.cr7", `fichar "./jugadas.cr7" como tactica; jugada comandante(){ devuelve tactica.rematar(7, 10); }`)
	i, err := loader.New(nil).Load(entry)
	if err != nil {
		t.Fatal(err)
	}
	function, ok := exported(t, i, "comandante").Function()
	if !ok {
		t.Fatal("entry function not exported")
	}
	result, err := function.Call(nil)
	if err != nil {
		t.Fatal(err)
	}
	number, ok := result.Number()
	if !ok || number != 70 {
		t.Fatalf("explicit entry call = %v; want 70", result)
	}
}

func TestParentRelativeImportsAreNotSandboxed(t *testing.T) {
	root := t.TempDir()
	fixture(t, root, "outside.cr7", `leyenda label = "outside entry directory";`)
	entry := fixture(t, root, "application/main.cr7", `fichar "../outside.cr7" como outside; leyenda label = outside.label;`)
	i, err := loader.New(nil).Load(entry)
	if err != nil {
		t.Fatal(err)
	}
	textExport(t, i, "label", "outside entry directory")
}

func TestDependencyFailureKeepsChainAndEffectsOnce(t *testing.T) {
	root := t.TempDir()
	lib := &library{}
	l := loader.New(lib)
	fixture(t, root, "dep.cr7", `leyenda effect = siuuu("dep"); leyenda broken = 1 / 0;`)
	entry := fixture(t, root, "main.cr7", `fichar "./dep.cr7" como dep; leyenda effect = siuuu("main");`)
	for range 2 {
		i, err := l.Load(entry)
		requireLoadError(t, i, err, "DivisionPorCero")
		for _, name := range []string{"main.cr7", "dep.cr7"} {
			if !strings.Contains(err.Error(), name) {
				t.Fatalf("missing %s in failure chain: %v", name, err)
			}
		}
	}
	if len(lib.events) != 1 || lib.events[0] != "dep" {
		t.Fatalf("initialization effects = %v", lib.events)
	}
	// Loading the failed dependency directly must not rerun its initializer.
	i, err := l.Load(filepath.Join(root, "dep.cr7"))
	requireLoadError(t, i, err, "DivisionPorCero")
	if len(lib.events) != 1 {
		t.Fatalf("failed dependency initialized again: %v", lib.events)
	}
}

var _ value.Module = exports{}
