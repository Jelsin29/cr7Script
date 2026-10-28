# CR7Script proposed architecture

> **Status: proposed design — not implemented.** No interpreter, compiler, CLI, or tests are implemented. Implementation must remain unstarted until the owner approves the design and resolves the open decisions below.
>
> Documentation: English · Language vocabulary: Spanish-first
>
> Identity: a small scripting language inspired by Cristiano Ronaldo, especially his Real Madrid era.

This document proposes implementation boundaries for the [0.1 language reference](language-reference.md). It is an implementation plan, not an alternative language contract. Read the reference for syntax and semantics, then use the [roadmap](roadmap.md) for dependencies and proposed acceptance evidence.

## Quick review path

1. Review the pipeline and boundary table below against the language reference.
2. Resolve the four [open decisions](#open-decisions-before-implementation); do not infer answers from package names or phase diagrams.
3. Approve the design before starting the proposed milestone sequence.

## Proposed execution model

The draft proposes a tree-walking interpreter: source is tokenized, parsed into an AST, resolved, loaded with its dependencies, and evaluated. This describes responsibilities, not a requirement that all modules complete one phase globally before the next begins. The loader would coordinate frontend processing of dependencies before initialization.

All responsibilities below are **proposed, not implemented**; their labels preserve the blueprint:

- Lexer: Tokens, comments, literals, locations
- Parser: Grammar, precedence, AST
- Resolver: Scope mapping, invalid control flow, binding rules
- Loader: Canonical paths, dependency cycles, module cache
- Evaluator: Expressions, calls, mutation, runtime checks
- Value layer: Runtime types, equality, collection identity, display
- Standard library: Native function contracts
- CLI: Arguments, exit status, streams, presentation

### Proposed boundaries

| Boundary | Proposed responsibility and handoff | Must not silently decide |
|---|---|---|
| Frontend | Lexer retains source locations; parser builds AST nodes; resolver maps lexical bindings and rejects invalid control flow and binding forms. Diagnostics retain source positions for later presentation. | Whether unknown and early-read names fail during resolution or evaluation; non-finite literal diagnostic phase. |
| Runtime | Evaluator walks the resolved AST with lexical environments and closures. Value layer centralizes types, equality, reference identity, display, and runtime checks. | Module constant ordering or guarantees for an unresolved `check` contract. |
| Module loading | Loader resolves standard names and importer-relative local paths, canonicalizes paths including symlinks, detects cycles with an import chain, caches modules, and initializes dependencies once per process. | Ordering and early reads among constants within a module. |
| Standard library | Native functions use the same value and error contracts as user functions; `rueda_de_prensa` exposes the seven sourced Q1–Q7 entries as immutable text constants, with no implicit output. | New syntax aliases, implicit conversions, remote packages, unverified quote exports, or filesystem/network APIs. |
| CLI | Proposed runner, syntax checker, REPL, help, and version interfaces share the frontend and runtime rather than reimplementing semantics. Presentation and process status remain outside language evaluation. | Exact command syntax and the full guarantees of `check`. |

The proposed runner initializes the entry module, then calls exactly one zero-argument top-level `comandante()` once. Its result must be `nada` or an integer exit code from 0 through 125. Imported modules and the REPL do not automatically invoke `comandante`. Proposed program output follows native I/O contracts; diagnostic presentation belongs on stderr, separately from program stdout. CLI spellings and diagnostic exit statuses still need specification before CLI implementation; no executable command is available now.

## Proposed runtime invariants

These are draft constraints from the language reference, not implemented guarantees:

- Dynamic values use strict runtime type checks, with eager left-to-right evaluation and boolean short-circuiting.
- Lexical block scope supports closures; fresh loop bindings preserve per-iteration captures. `leyenda` prevents rebinding, not collection mutation.
- Execution is single-threaded, with host-managed allocation and collection. Go is the proposed host, pending confirmation and a minimum version.
- Source files use UTF-8 and `.cr7`; simple statements require semicolons. Diagnostic locations are one-based and columns count Unicode scalar values.
- Runtime errors, ordinary return/break/continue control flow, and frontend/module-loading failures remain distinct. `remontada` catches runtime errors only; it neither rolls back effects nor retries.
- Modules expose read-only namespaces, while exported collection references can remain mutable. Canonical paths and cycle detection are not a security sandbox.

## Proposed repository layout

Every code and test path in this table is a **proposal, not an existing implementation**. Package names organize responsibilities only; they do not approve Go or settle language semantics.

| Proposed path | Purpose |
|---|---|
| `cmd/cr7script/` | Thin CLI entry point, if Go is confirmed. |
| `internal/lexer/`, `internal/parser/`, `internal/ast/` | Source locations, tokens, parsing, and syntax tree. |
| `internal/resolver/` | Lexical scope mapping and static validation. |
| `internal/loader/` | Local/standard module resolution, cycles, cache, and initialization coordination. |
| `internal/evaluator/`, `internal/value/` | AST execution, environments, closures, runtime values, and errors. |
| `internal/stdlib/` | Core prelude and proposed standard modules, including `rueda_de_prensa` for immutable sourced quote constants. |
| `internal/cli/` | Argument handling, streams, status, and REPL presentation. |
| `tests/conformance/`, `tests/cli/`, `tests/fixtures/` | Language contract cases, process-level checks, and source/diagnostic fixtures. |
| `examples/` | Proposed `.cr7` examples; not executable evidence until a runner exists. |

The documentation set is [language reference](language-reference.md), this architecture proposal, and [roadmap](roadmap.md). The layout does not authorize creating code during documentation work.

## Proposed test layers

No tests or test commands in this plan have been implemented or executed. Future acceptance evidence should distinguish expected results from observed results.

| Proposed layer | Contract to exercise | Proposed evidence |
|---|---|---|
| Frontend unit tests | Tokens, comments, literals, locations, grammar, precedence, and invalid syntax. | Token/AST expectations and exact diagnostic fixtures. |
| Resolver tests | Shadowing, binding rules, closures, and invalid control flow. | Public frontend results, including phase expectations after owner decisions. |
| Evaluator/value conformance | Strict types, evaluation order, arithmetic boundaries, identity, mutation, iteration snapshots, display, and recovery. | `.cr7` fixtures with expected values, output, errors, and retained side effects. |
| Loader integration | Canonical paths, aliases, cache identity, dependency initialization, and cycle rejection. | Local multi-module fixtures and expected import-chain diagnostics. |
| Native library tests | Arity, type checks, Unicode behavior, empty inputs, mutation contracts, and quote-data fidelity. | Cases derived from the reference for each prelude/standard function; verify every Q1–Q7 export is exact immutable text, importing the module emits nothing, and the unverified “losh dosh” lead is not exported. |
| CLI end-to-end tests | Runner, checker, REPL, help/version, entry validation, and stream separation. | Exact stdout, stderr, and exit-status assertions once CLI syntax and `check` guarantees are specified. |

The reference's quick example is a future conformance fixture, not proof of execution. Boundary/error cases must accompany happy paths; a recoverable failure must also demonstrate that prior effects remain. Quote fixtures should compare against the owner-supplied Q1–Q7 catalog and retain its source caveats; this documentation update does not claim independent audio/video review.

## Open decisions before implementation

Owner approval is required; these remain **open**, not corrected by this proposal:

1. **Module constants:** Define module-level constant initialization order and behavior when an initializer reads a later constant.
2. **Non-finite literals:** Define handling and diagnostic phase for non-finite numeric literals such as `1e999`.
3. **Name errors and checking:** Define when `NombreNoDefinido` and `NombreSinInicializar` are detected and what `check` validates.
4. **Host implementation:** Confirm Go as the implementation language and choose its minimum supported version.

**Planning assumption only:** Go-oriented package paths and host-managed memory describe the draft direction, not an owner-approved implementation choice. No schedule or package diagram resolves these questions.

## Scope and possible evolution

Proposed 0.1 excludes classes, inheritance, operator overloading, macros, generics, threads, network access, package installation, JIT, native code generation, and self-hosting. Async/parallel syntax, slices, and destructuring are also outside the reference's 0.1 scope. If Go is confirmed, the draft limits implementation dependencies to its standard library and local modules.

A bytecode interpreter is a **possible later optimization**, not a milestone commitment. The initial proposal remains tree-walking; future bytecode work would need a separate approved change and conformance evidence preserving language behavior.

## Project change control

Keep the proposed language contract distinct from implemented behavior. When the owner approves a language change, update the reference, architecture, and affected milestone acceptance criteria together before implementation.
