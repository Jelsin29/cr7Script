# CR7Script proposed architecture

> **Status: approved design; implementation in progress.** The owner approved the Go 1.27+ baseline and the previously open semantic decisions. The lexer, AST, and parser are implemented with focused tests; later milestones remain unimplemented until their acceptance evidence is observed.
>
> Documentation: English · Language vocabulary: Spanish-first
>
> Identity: a small scripting language inspired by Cristiano Ronaldo, especially his Real Madrid era.

This document proposes implementation boundaries for the [0.1 language reference](language-reference.md). It is an implementation plan, not an alternative language contract. Read the reference for syntax and semantics, then use the [roadmap](roadmap.md) for dependencies and proposed acceptance evidence.

## Quick review path

1. Review the pipeline and boundary table below against the language reference.
2. Use the [owner-approved decisions](#owner-approved-decisions) for module initialization, literal diagnostics, name resolution, and `check`.
3. Progress through the milestone sequence and record observed acceptance evidence; do not infer completion from the plan.

## Proposed execution model

The approved 0.1 baseline specifies a tree-walking interpreter: source is tokenized, parsed into an AST, resolved, loaded with its dependencies, and evaluated. This describes responsibilities, not a requirement that all modules complete one phase globally before the next begins. The loader coordinates frontend processing of dependencies before initialization.

The lexer, AST, parser, and focused frontend tests are implemented; the remaining responsibilities below are proposals, not implemented:

- Lexer: Tokens, comments, literals, locations — implemented
- AST: Syntax tree nodes for declarations, statements, and expressions — implemented
- Parser: Grammar, precedence, AST construction — implemented
- Resolver: Scope mapping, invalid control flow, binding rules
- Loader: Canonical paths, dependency cycles, module cache
- Evaluator: Expressions, calls, mutation, runtime checks
- Value layer: Runtime types, equality, collection identity, display
- Standard library: Native function contracts
- CLI: Arguments, exit status, streams, presentation

### Proposed boundaries

| Boundary | Approved responsibility and handoff | Must not silently decide |
|---|---|---|
| Frontend | Lexer retains source locations; parser builds AST nodes; resolver maps lexical bindings and rejects invalid control flow and binding forms. Unknown names and reads before initialization are resolver diagnostics; non-finite literals produce `NumeroNoFinito` during analysis. | New language semantics beyond the reference and owner-approved decisions. |
| Runtime | Evaluator walks the resolved AST with lexical environments and closures. Value layer centralizes types, equality, reference identity, display, and runtime checks. | New language semantics beyond the reference and owner-approved decisions. |
| Module loading | Loader resolves standard names and importer-relative local paths, canonicalizes paths including symlinks, detects cycles with an import chain, caches modules, and initializes dependencies once per process. Top-level functions initialize before constants; constants evaluate in source order. | New module semantics beyond the approved initialization order. |
| Standard library | Native functions use the same value and error contracts as user functions; `rueda_de_prensa` exposes the seven sourced Q1–Q7 entries as immutable text constants, with no implicit output. | New syntax aliases, implicit conversions, remote packages, unverified quote exports, or filesystem/network APIs. |
| CLI | Runner, non-executing syntax/binding checker, REPL, help, and version interfaces share the frontend rather than reimplementing semantics. Presentation and process status remain outside language evaluation. | Command spellings or checker behavior beyond the approved syntax/binding-only guarantee. |

The proposed runner initializes the entry module, then calls exactly one zero-argument top-level `comandante()` once. Its result must be `nada` or an integer exit code from 0 through 125. Imported modules and the REPL do not automatically invoke `comandante`. Proposed program output follows native I/O contracts; diagnostic presentation belongs on stderr, separately from program stdout. CLI spellings and diagnostic exit statuses still need specification before CLI implementation; no executable command is available now.

## Proposed runtime invariants

These are draft constraints from the language reference, not implemented guarantees:

- Dynamic values use strict runtime type checks, with eager left-to-right evaluation and boolean short-circuiting.
- Lexical block scope supports closures; fresh loop bindings preserve per-iteration captures. `leyenda` prevents rebinding, not collection mutation.
- Execution is single-threaded, with host-managed allocation and collection. Go 1.27 or newer is the approved host toolchain baseline.
- Source files use UTF-8 and `.cr7`; simple statements require semicolons. Diagnostic locations are one-based and columns count Unicode scalar values.
- Runtime errors, ordinary return/break/continue control flow, and frontend/module-loading failures remain distinct. `remontada` catches runtime errors only; it neither rolls back effects nor retries.
- Modules expose read-only namespaces, while exported collection references can remain mutable. Canonical paths and cycle detection are not a security sandbox.

## Proposed repository layout

`internal/lexer/`, `internal/ast/`, and `internal/parser/` exist; the lexer implements tokenization, literals, comments, and locations, while the AST and parser implement the approved syntax tree and grammar with focused tests. Every other code and test path in this table remains a proposal. Package names organize responsibilities only; they do not settle language semantics.

| Proposed path | Purpose |
|---|---|
| `cmd/cr7script/` | Thin CLI entry point, if Go is confirmed. |
| `internal/ast/` | Syntax tree nodes for declarations, statements, and expressions (implemented). |
| `internal/lexer/` | Source locations, tokens, comments, and literals (implemented). |
| `internal/parser/` | Approved grammar, precedence, and AST construction (implemented). |
| `internal/resolver/` | Lexical scope mapping and static validation. |
| `internal/loader/` | Local/standard module resolution, cycles, cache, and initialization coordination. |
| `internal/evaluator/`, `internal/value/` | AST execution, environments, closures, runtime values, and errors. |
| `internal/stdlib/` | Core prelude and proposed standard modules, including `rueda_de_prensa` for immutable sourced quote constants. |
| `internal/cli/` | Argument handling, streams, status, and REPL presentation. |
| `tests/conformance/`, `tests/cli/`, `tests/fixtures/` | Language contract cases, process-level checks, and source/diagnostic fixtures. |
| `examples/` | Proposed `.cr7` examples; not executable evidence until a runner exists. |

The documentation set is [language reference](language-reference.md), this architecture proposal, and [roadmap](roadmap.md). The layout does not authorize creating code during documentation work.

## Proposed test layers

Lexer and parser unit tests have been implemented and run for representative frontend cases. Resolver, evaluator/value, loader, standard-library, and CLI layers remain proposals; the test plan below describes future acceptance evidence and does not claim those layers exist.

| Proposed layer | Contract to exercise | Proposed evidence |
|---|---|---|
| Frontend unit tests | Tokens, comments, literals, locations, grammar, precedence, and invalid syntax. | Token/AST expectations and exact diagnostic fixtures. |
| Resolver tests | Shadowing, binding rules, closures, and invalid control flow. | Public frontend results, including phase expectations after owner decisions. |
| Evaluator/value conformance | Strict types, evaluation order, arithmetic boundaries, identity, mutation, iteration snapshots, display, and recovery. | `.cr7` fixtures with expected values, output, errors, and retained side effects. |
| Loader integration | Canonical paths, aliases, cache identity, dependency initialization, and cycle rejection. | Local multi-module fixtures and expected import-chain diagnostics. |
| Native library tests | Arity, type checks, Unicode behavior, empty inputs, mutation contracts, and quote-data fidelity. | Cases derived from the reference for each prelude/standard function; verify every Q1–Q7 export is exact immutable text, importing the module emits nothing, and the unverified “losh dosh” lead is not exported. |
| CLI end-to-end tests | Runner, checker, REPL, help/version, entry validation, and stream separation. | Exact stdout, stderr, and exit-status assertions once CLI syntax and `check` guarantees are specified. |

The reference's quick example is a future conformance fixture, not proof of execution. Boundary/error cases must accompany happy paths; a recoverable failure must also demonstrate that prior effects remain. Quote fixtures should compare against the owner-supplied Q1–Q7 catalog and retain its source caveats; this documentation update does not claim independent audio/video review.

## Owner-approved decisions

The owner resolved the former open questions for the 0.1 implementation baseline:

1. **Host implementation:** Go 1.27 or newer.
2. **Module constants:** Top-level functions initialize before constants; constants evaluate in source order; a read of a later constant is a resolver error `NombreSinInicializar`.
3. **Non-finite literals:** Analysis rejects values such as `1e999` with `NumeroNoFinito` at the literal location before execution.
4. **Name errors and checking:** The resolver reports `NombreNoDefinido` and `NombreSinInicializar`; `check` validates syntax and bindings without executing.

## Scope and possible evolution

Approved 0.1 excludes classes, inheritance, operator overloading, macros, generics, threads, network access, package installation, JIT, native code generation, and self-hosting. Async/parallel syntax, slices, and destructuring are also outside the reference's 0.1 scope. The Go implementation uses the standard library and local modules only.

A bytecode interpreter is a **possible later optimization**, not a milestone commitment. The initial proposal remains tree-walking; future bytecode work would need a separate approved change and conformance evidence preserving language behavior.

## Project change control

Keep the proposed language contract distinct from implemented behavior. When the owner approves a language change, update the reference, architecture, and affected milestone acceptance criteria together before implementation.
