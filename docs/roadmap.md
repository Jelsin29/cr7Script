# CR7Script proposed delivery roadmap

> **Status: approved design; implementation in progress.** Milestone scopes and acceptance evidence below remain plans unless an outcome is explicitly verified. The owner approved Go 1.27+ and resolved the four previously open decisions. Implementation began with the lexer; loader and standard-library integration now have focused and repository-wide Go verification, while the runner/checker remains pending.
>
> Documentation: English · Language vocabulary: Spanish-first
>
> Identity: a small scripting language inspired by Cristiano Ronaldo, especially his Real Madrid era.

This plan sequences the proposed 0.1 tree-walking interpreter without expanding the [language reference](language-reference.md). Use the [architecture](architecture.md) for pipeline responsibilities, package boundaries, and proposed test layers. Milestone names and order follow the blueprint; the evidence below is a proposed review gate, not a delivery claim or schedule commitment.

## Quick path through implementation

1. Use the approved reference and architecture together; keep language rules separate from package choices.
2. Apply the [owner-approved decisions](#approved-design-baseline) for Go, module initialization, literal diagnostics, name resolution, and `check`.
3. Progress through M2–M7 using observed acceptance evidence; do not claim a milestone is complete from its plan alone.

## Proposed milestones and acceptance evidence

Milestone scopes and acceptance criteria are planned; mark implementation outcomes complete only after observing their evidence. Dependencies are readiness gates, not claims that earlier milestones are complete. The approved minimum host toolchain is Go 1.27.

| Milestone | Proposed scope | Dependencies | Proposed acceptance evidence |
|---|---|---|---|
| M1 — Kickoff | Record the approved 0.1 contract and owner decisions; confirm the Go 1.27+ toolchain and agree the minimal repository and test layout. Keep language design distinct from implementation planning. | Owner review of the reference, architecture, roadmap, and resolved decisions. | Owner-approved decisions recorded in the public reference and architecture; minimum host version and test plan documented. |
| M2 — Cantera | Establish the frontend foundations: source locations, tokens, comments, literals, AST, grammar, and precedence. | M1 approval, including the non-finite literal diagnostic decision. | Proposed lexer/parser tests demonstrate valid and invalid lexical forms, Unicode locations, semicolons, precedence, assignment targets, and syntax diagnostics. Compare tokens/AST and exact diagnostic fixtures with the reference. |
| M3 — Training | Introduce values, lexical bindings, expression evaluation, mutation, conditionals, and iteration with strict runtime checks. Establish the relevant resolver binding/control-flow rules. | M2 frontend; approved name-error phase and early-read rules from M1. | Proposed conformance cases demonstrate strict booleans, left-to-right evaluation, short-circuiting, arithmetic errors, collection identity/mutation, shadowing, iteration snapshots, fresh loop bindings, and invalid loop control. Report expected versus observed outcomes. |
| M4 — El Comandante | Add user functions, closures, recursion, return behavior, exact arity, and entry-point validation/invocation. | M3 environments, values, evaluator, and resolver; M2 function syntax. | Proposed cases demonstrate captured bindings, recursive calls, argument order, arity errors, invalid returns/control flow, and one automatic entry invocation only in runner context. Check `nada` and exit codes 0–125, including rejected entry signatures/results. |
| M5 — Remontada | Add runtime error values, raising/rethrowing, protected blocks, handlers, and diagnostic stack presentation. | M4 call frames and control flow; M3 runtime checks; M1 phase decisions. | Proposed cases demonstrate `falta`, immutable handler bindings, read-only error fields, exact codes/locations, and stack frames. Confirm recovery preserves earlier effects, does not retry, does not catch return/break/continue, and does not catch frontend or loading failures. |
| M6 — Mr. Champions | Integrate module loading/cache and native contracts: core prelude, `mr_champions`, `relato`, `marcador`, and `rueda_de_prensa`, without adding library scope. The latter exposes exactly the seven catalogued quote constants as immutable text; quotes are data, not new syntax. | M5 error model; M4 callable values; M1 approved constant-initialization rules and reviewed quote provenance. | Proposed local-module fixtures demonstrate canonical paths, symlink identity, rejected cycles with import chains, once-per-process initialization, exports, and no automatic imported entry call. Native-function tests cover arity/types, empty inputs, Unicode, finite arithmetic, display, and documented mutation/non-mutation contracts. Quote-module tests compare all seven exports to the Q1–Q7 source excerpts, verify Q7 retains its colon, ensure importing emits no output, keep `HALA_MADRID` distinct from the `hala_madrid()` greeting and `mr_champions.soy_el_mejor(numbers)`, and confirm the unverified “losh dosh” lead has no export. |
| M7 — Matchday | Integrate the proposed runner, syntax checker, REPL, help, and version interfaces; review examples and release-facing 0.1 documentation against the implemented subset. | M2–M6 language/runtime/library evidence; M1 approved `check` guarantees and a specified CLI interface. | Proposed process-level tests assert exact stdout, stderr, and exit status for success and failure, with checker guarantees and REPL entry behavior tested explicitly. The reference quick example must produce its exact expected output on an actual runner before claiming it works. Record conformance results and remaining limitations rather than claiming untested completion. |

**Observed implementation evidence (M6 / T7):** the local-module loader and standard library are implemented; focused tests cover importer-relative and canonical paths, symlink/cache identity, dependency initialization, cycle diagnostics, the 17-function prelude, four registered modules, and exact Q1–Q7 immutable quote exports without implicit output. Repository-wide `go test ./...` and `go vet ./...` pass. This does not establish the runner/checker or prove that reference examples execute; M7 remains pending.

These scopes organize work, not language additions. Cross-cutting value, source-location, and error plumbing may be needed before its milestone's full contract is accepted. Any change to an approved language rule requires a separately reviewed specification change, not an undocumented implementation shortcut.

## Proposed verification and completion gates

Observed validation so far: focused tests cover the implemented M2–M6 packages, and repository-wide `go test ./...` plus `go vet ./...` pass. No runner or CLI exists yet, so language-reference examples and process-level M7 behavior remain unverified. Continue using these evidence gates:

- **At each milestone:** run focused deterministic tests for its public behavior, retain regression checks for shared behavior, and record observed results with the implementation change.
- **Across boundaries:** check resolver/evaluator error phases, loader initialization, native values, and CLI presentation against the same reference rather than divergent package assumptions.
- **Before 0.1 acceptance:** compare documentation, examples, help text, diagnostics, and actual conformance results. Do not label a proposed example as tested merely because it is documented.

Future runner/checker/REPL command spellings are not specified here. `check` names the proposed checker interface; its validation guarantees remain an owner decision. No installation, build, test, or execution command is represented as available today.

## Approved design baseline

The owner resolved the former implementation gates:

| Decision | Approved behavior |
|---|---|
| Host toolchain | Go 1.27 or newer; standard library and local modules only. |
| Module constants | Initialize top-level functions first, then evaluate constants in source order; a later-constant read is `NombreSinInicializar`. |
| Non-finite literals | Reject during analysis as `NumeroNoFinito` at the literal location, before execution. |
| Names and `check` | Resolver reports `NombreNoDefinido` and `NombreSinInicializar`; `check` validates syntax and bindings without executing. |

Update the reference, architecture, and milestone acceptance criteria together if the owner later approves a language-contract change.

## Project change control

The approved baseline is recorded in the reference and architecture. For each milestone, compare implementation and observed acceptance evidence with that baseline; document deviations and obtain owner approval before changing language behavior. A written proposal is not evidence that the language executes.

## Deferred beyond proposed 0.1

The draft defers classes, inheritance, operator overloading, macros, generics, threads, network access, package installation, JIT, native code generation, and self-hosting. Async/parallel syntax, slices, and destructuring are outside the reference's 0.1 scope. Do not introduce them through milestone implementation.

Tree-walking remains the proposed initial execution model. Bytecode is a possible later optimization only, not an eighth milestone or a commitment. Any later evolution would require separate owner-approved scope and evidence that observable language behavior remains consistent.

## Next review

Review the [approved language decisions](language-reference.md#owner-approved-implementation-decisions) and [architecture baseline](architecture.md#owner-approved-decisions) together. The next implementation slice is the runner/checker; examples remain unverified until the runner demonstrates their behavior.
