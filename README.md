# CR7Script

CR7Script is a small, Spanish-first scripting language that gives programming concepts a football vocabulary. Its approved 0.1 design favors predictable evaluation, explicit rules, and a compact standard library. Cristiano Ronaldo and his Real Madrid era are historical inspiration for the names, not an endorsement or a source of programming semantics.

> **Status:** The CR7Script 0.1 language contract is approved and its lexer, parser, resolver, interpreter core, module loader, and standard library are implemented. A supported command-line runner, syntax checker, and REPL are **not available yet**. The examples below describe the language contract; they are not runnable CLI instructions.

## At a glance

| | |
|---|---|
| Source files | UTF-8 text with the `.cr7` extension |
| Implementation host | Go 1.27 or newer; Go is not CR7Script syntax and CR7Script is not transpiled to Go |
| Execution model | Tree-walking interpreter; single-threaded |
| Type model | Dynamic values with strict runtime type checks and no implicit truthiness |
| Scope | Lexical block scope, functions, closures, lists, dictionaries, modules, and runtime error recovery |
| Current limitation | No supported executable, `check` command, or REPL |

## What the language can express

- **Bindings:** `cantera` declares a mutable name; `leyenda` declares a name that cannot be rebound.
- **Functions:** `jugada` declares first-class functions and lexical closures. `devuelve` returns a value.
- **Control flow:** `si` / `sino` branch; `entrena` repeats while a boolean condition is true; `temporada name en values` iterates a snapshot; `vestuario` and `siguiente` break and continue the nearest loop.
- **Values and collections:** finite binary64 numbers, Unicode text, booleans (`gol` / `fuera`), `nada`, mutable lists, insertion-ordered string-key dictionaries, functions, modules, and errors.
- **Modules:** top-level `fichar` imports a registered standard module or an importer-relative local `.cr7` file. Imports do not download packages.
- **Errors:** stable diagnostic codes distinguish syntax, resolver, loading, and runtime failures. `remontada` catches runtime errors only; it does not undo earlier effects or retry automatically.

The exact grammar, operators, edge cases, diagnostics, module behavior, and standard-library contracts are in the [language reference](docs/language-reference.md).

## Example

This snippet illustrates the approved syntax; it has not been run through a supported command-line runner:

```cr7
jugada comandante() {
    leyenda dorsal = 7;
    cantera goles = 0;

    entrena goles < 3 {
        goles = goles + 1;
        siuuu("Gol", goles, "del", dorsal);
    }
}
```

The proposed runner will invoke one zero-argument top-level `comandante()` after module initialization. Until that runner exists, do not expect this file to execute from a shell.

## How CR7Script works

The Go implementation processes source through cooperating stages:

1. **Lexer** converts UTF-8 source into tokens with source locations.
2. **Parser** builds an abstract syntax tree (AST) using the approved grammar and operator precedence.
3. **Resolver** validates lexical names, initialization order, declarations, and control flow before evaluation.
4. **Module loader** resolves local and registered modules, detects cycles, caches canonical paths, and initializes dependencies.
5. **Value layer and evaluator** execute the resolved tree with lexical environments, closures, strict type checks, and runtime errors.
6. **Standard library** supplies core functions and four registered modules; importing a module does not itself print output.

At module scope, functions initialize before constants; constants then evaluate in source order. This supports forward calls between top-level functions while rejecting reads of later constants. See [architecture](docs/architecture.md) for package boundaries and [roadmap](docs/roadmap.md) for implementation status and evidence.

## Standard library

The core prelude includes output and input (`siuuu`, `entrevista`), conversion and inspection (`texto`, `numero`, `tipo`, `palmares`), timing and ranges (`calma`, `rango`), list/dictionary helpers, and assertions. Registered modules are:

- `mr_champions` — numeric statistics and player selection.
- `relato` — text operations.
- `marcador` — numeric operations.
- `rueda_de_prensa` — immutable text exports for the sourced Q1–Q7 quote catalog.

The quote module contains only the documented excerpts and preserves their provenance caveats; it is not a live statistics or quotation service. See the [language reference](docs/language-reference.md) for the complete proposed standard-library contract.

## Documentation

- [Documentation site source](docs/index.html) — overview and documentation hub; additional guides are being added in this branch.
- [Getting started and current status](docs/getting-started.html) — explains what is and is not available today.
- [Language tour](docs/language-tour.html) — guided syntax, values, control flow, functions, and collections.
- [Language reference](docs/language-reference.md) — authoritative 0.1 contract.
- [Architecture](docs/architecture.md) — interpreter pipeline and implementation boundaries.
- [Roadmap](docs/roadmap.md) — milestones and observed verification evidence.

The static site under `docs/` is compatible with GitHub Pages. This branch has not been published and no repository Pages settings were changed. If the owner later configures a project site for this repository, its expected URL is `https://jelsin29.github.io/cr7Script/`; this README makes no claim about existing Pages settings.

## Development checks

The Go implementation requires Go 1.27 or newer. From the repository root, contributors can run:

```sh
go test ./...
go vet ./...
```

These commands test and analyze the Go implementation; they do not execute CR7Script programs.

## Scope and safety

CR7Script 0.1 does not include classes, inheritance, macros, generics, asynchronous or parallel syntax, network access, package installation, JIT compilation, native code generation, or a security sandbox. Canonical path handling and module-cycle detection are not isolation boundaries. Do not run untrusted CR7Script as though the interpreter were sandboxed.

CR7Script is an unofficial fan-inspired project. No endorsement by Cristiano Ronaldo, Real Madrid, UEFA, or any other organization is asserted. Names, likenesses, logos, and commercialization may require separate permission.
