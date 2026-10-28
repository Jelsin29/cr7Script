// Package loader resolves and initializes local and registered standard modules.
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jelsin29/cr7Script/internal/evaluator"
	"github.com/Jelsin29/cr7Script/internal/lexer"
	"github.com/Jelsin29/cr7Script/internal/parser"
	"github.com/Jelsin29/cr7Script/internal/resolver"
	"github.com/Jelsin29/cr7Script/internal/value"
)

// Library supplies the prelude and the complete registry of standard modules.
// Module names are never fetched from the network.
type Library interface {
	Prelude() map[string]value.Value
	Module(name string) (value.Module, bool)
}

type result struct {
	interpreter *evaluator.Interpreter
	namespace   value.Value
	err         error
}

// Loader owns one single-threaded runtime and its canonical module cache.
// Successful and failed initializations are retained to avoid repeating effects.
type Loader struct {
	library   Library
	prelude   map[string]value.Value
	execution *evaluator.ExecutionContext
	modules   map[string]result
	standard  map[string]value.Value
	active    []string
}

// New creates an independent loader/runtime using the supplied library.
func New(library Library) *Loader {
	prelude := make(map[string]value.Value)
	if library != nil {
		for name, item := range library.Prelude() {
			prelude[name] = item
		}
	}
	return &Loader{library: library, prelude: prelude, execution: evaluator.NewExecutionContext(), modules: make(map[string]result), standard: make(map[string]value.Value)}
}

// Load initializes the entry and its dependencies without invoking comandante.
// The entry may be an absolute or working-directory-relative filesystem path;
// only import declarations require ./ or ../ local paths.
func (l *Loader) Load(entryPath string) (*evaluator.Interpreter, error) {
	canonical, err := canonicalPath(entryPath)
	if err != nil {
		return nil, l.failure("ModuloNoEncontrado", entryPath, err.Error())
	}
	loaded := l.loadFile(canonical)
	return loaded.interpreter, loaded.err
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func (l *Loader) failure(code, path, message string) error {
	chain := append(append([]string(nil), l.active...), path)
	return fmt.Errorf("%s: %s (import chain: %s)", code, message, strings.Join(chain, " -> "))
}

func (l *Loader) loadFile(path string) (loaded result) {
	if cached, ok := l.modules[path]; ok {
		return cached
	}
	for _, active := range l.active {
		if active == path {
			return result{err: l.failure("FichajeCircular", path, "cyclic module dependency")}
		}
	}
	l.active = append(l.active, path)
	defer func() { l.active = l.active[:len(l.active)-1]; l.modules[path] = loaded }()
	source, err := os.ReadFile(path)
	if err != nil {
		return result{err: l.failure("ModuloNoEncontrado", path, err.Error())}
	}
	tokens, err := lexer.Scan(string(source))
	if err != nil {
		return result{err: l.sourceFailure(path, err)}
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return result{err: l.sourceFailure(path, err)}
	}
	// Validate bindings before executing any dependencies of an invalid module.
	if err = resolver.Resolve(program); err != nil {
		return result{err: l.sourceFailure(path, err)}
	}
	imports := make(map[string]value.Value, len(program.Imports))
	for _, declaration := range program.Imports {
		namespace, err := l.resolve(path, declaration.Path)
		if err != nil {
			return result{err: fmt.Errorf("%s:%d:%d: %v", path, declaration.Start.Line, declaration.Start.Column, err)}
		}
		imports[declaration.Alias] = namespace
	}
	interpreter := evaluator.NewWithRuntime(l.prelude, l.execution)
	if err = interpreter.LoadFromWithImports(program, path, imports); err != nil {
		return result{err: l.sourceFailure(path, err)}
	}
	exports, ok := interpreter.Exports()
	if !ok {
		return result{err: l.sourceFailure(path, fmt.Errorf("module initialized without exports"))}
	}
	return result{interpreter: interpreter, namespace: value.NewModule(exports)}
}

func (l *Loader) sourceFailure(path string, err error) error {
	// Deliberately do not wrap with %w: loading errors must not expose a runtime
	// error through errors.As, even when a constant initializer caused the failure.
	return fmt.Errorf("%s: %v (import chain: %s)", path, err, strings.Join(l.active, " -> "))
}

func (l *Loader) resolve(importer, path string) (value.Value, error) {
	if filepath.IsAbs(path) || filepath.VolumeName(path) != "" {
		return value.Value{}, l.failure("ModuloNoEncontrado", path, "absolute import paths are not allowed")
	}
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		if !strings.HasSuffix(path, ".cr7") {
			return value.Value{}, l.failure("ModuloNoEncontrado", path, "local imports require a .cr7 extension")
		}
		canonical, err := canonicalPath(filepath.Join(filepath.Dir(importer), path))
		if err != nil {
			return value.Value{}, l.failure("ModuloNoEncontrado", path, err.Error())
		}
		loaded := l.loadFile(canonical)
		return loaded.namespace, loaded.err
	}
	if cached, ok := l.standard[path]; ok {
		return cached, nil
	}
	if l.library != nil {
		if module, ok := l.library.Module(path); ok && module != nil {
			namespace := value.NewModule(module)
			l.standard[path] = namespace
			return namespace, nil
		}
	}
	return value.Value{}, l.failure("ModuloNoEncontrado", path, "unknown standard module")
}
