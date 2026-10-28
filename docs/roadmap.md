# CR7Script proposed delivery roadmap

> **Status: proposed design — not implemented.** All milestones, code paths, commands, tests, and acceptance evidence below are planned, not completed or executed. Implementation remains unstarted until owner approval of the design and resolution of the four open decisions.
>
> Documentation: English · Language vocabulary: Spanish-first
>
> Identity: a small scripting language inspired by Cristiano Ronaldo, especially his Real Madrid era.

This plan sequences the proposed 0.1 tree-walking interpreter without expanding the [language reference](language-reference.md). Use the [architecture](architecture.md) for pipeline responsibilities, package boundaries, and proposed test layers. Milestone names and order follow the blueprint; the evidence below is a proposed review gate, not a delivery claim or schedule commitment.

## Quick path to an approved implementation plan

1. Review the reference and architecture together; keep language rules separate from package choices.
2. Resolve the [four open decisions](#owner-decisions-required-before-implementation) and explicitly approve the draft.
3. Track approved implementation changes separately from design proposals, then progress through M1–M7 using observed acceptance evidence.

## Proposed milestones and acceptance evidence

Every row is **proposed / not implemented**. Dependencies are readiness gates, not claims that earlier milestones are complete. Future tests should use a confirmed host toolchain; no Go build or test command is prescribed before that decision.

| Milestone | Proposed scope | Dependencies | Proposed acceptance evidence |
|---|---|---|---|
| M1 — Kickoff | Approve the 0.1 contract, settle the four open decisions, confirm host/toolchain, and agree the minimal repository and test layout. Keep language design distinct from implementation planning. | Owner review of the reference, architecture, roadmap, and open decisions. | Recorded owner resolutions and approval; documented minimum host version; reviewed scope/layout and test plan. |
| M2 — Cantera | Establish the frontend foundations: source locations, tokens, comments, literals, AST, grammar, and precedence. | M1 approval, including the non-finite literal diagnostic decision. | Proposed lexer/parser tests demonstrate valid and invalid lexical forms, Unicode locations, semicolons, precedence, assignment targets, and syntax diagnostics. Compare tokens/AST and exact diagnostic fixtures with the reference. |
| M3 — Training | Introduce values, lexical bindings, expression evaluation, mutation, conditionals, and iteration with strict runtime checks. Establish the relevant resolver binding/control-flow rules. | M2 frontend; approved name-error phase and early-read rules from M1. | Proposed conformance cases demonstrate strict booleans, left-to-right evaluation, short-circuiting, arithmetic errors, collection identity/mutation, shadowing, iteration snapshots, fresh loop bindings, and invalid loop control. Report expected versus observed outcomes. |
| M4 — El Comandante | Add user functions, closures, recursion, return behavior, exact arity, and entry-point validation/invocation. | M3 environments, values, evaluator, and resolver; M2 function syntax. | Proposed cases demonstrate captured bindings, recursive calls, argument order, arity errors, invalid returns/control flow, and one automatic entry invocation only in runner context. Check `nada` and exit codes 0–125, including rejected entry signatures/results. |
| M5 — Remontada | Add runtime error values, raising/rethrowing, protected blocks, handlers, and diagnostic stack presentation. | M4 call frames and control flow; M3 runtime checks; M1 phase decisions. | Proposed cases demonstrate `falta`, immutable handler bindings, read-only error fields, exact codes/locations, and stack frames. Confirm recovery preserves earlier effects, does not retry, does not catch return/break/continue, and does not catch frontend or loading failures. |
| M6 — Mr. Champions | Integrate module loading/cache and native contracts: core prelude, `mr_champions`, `relato`, `marcador`, and `rueda_de_prensa`, without adding library scope. The latter exposes exactly the seven catalogued quote constants as immutable text; quotes are data, not new syntax. | M5 error model; M4 callable values; M1 approved constant-initialization rules and reviewed quote provenance. | Proposed local-module fixtures demonstrate canonical paths, symlink identity, rejected cycles with import chains, once-per-process initialization, exports, and no automatic imported entry call. Native-function tests cover arity/types, empty inputs, Unicode, finite arithmetic, display, and documented mutation/non-mutation contracts. Quote-module tests compare all seven exports to the Q1–Q7 source excerpts, verify Q7 retains its colon, ensure importing emits no output, keep `HALA_MADRID` distinct from the `hala_madrid()` greeting and `mr_champions.soy_el_mejor(numbers)`, and confirm the unverified “losh dosh” lead has no export. |
| M7 — Matchday | Integrate the proposed runner, syntax checker, REPL, help, and version interfaces; review examples and release-facing 0.1 documentation against the implemented subset. | M2–M6 language/runtime/library evidence; M1 approved `check` guarantees and a specified CLI interface. | Proposed process-level tests assert exact stdout, stderr, and exit status for success and failure, with checker guarantees and REPL entry behavior tested explicitly. The reference quick example must produce its exact expected output on an actual runner before claiming it works. Record conformance results and remaining limitations rather than claiming untested completion. |

These scopes organize work, not language additions. Cross-cutting value, source-location, and error plumbing may be needed before its milestone's full contract is accepted. Any change to an approved language rule requires a separately reviewed specification change, not an undocumented implementation shortcut.

## Proposed verification and completion gates

No executable validation has occurred under this roadmap. The future evidence plan is:

- **At each milestone:** run focused deterministic tests for its public behavior, retain regression checks for shared behavior, and record observed results with the implementation change.
- **Across boundaries:** check resolver/evaluator error phases, loader initialization, native values, and CLI presentation against the same reference rather than divergent package assumptions.
- **Before 0.1 acceptance:** compare documentation, examples, help text, diagnostics, and actual conformance results. Do not label a proposed example as tested merely because it is documented.

Future runner/checker/REPL command spellings are not specified here. `check` names the proposed checker interface; its validation guarantees remain an owner decision. No installation, build, test, or execution command is represented as available today.

## Owner decisions required before implementation

All four questions remain **open**; this plan makes no implicit selection:

| Open question | Why it gates implementation |
|---|---|
| What is module-level constant initialization order, and what happens when an initializer reads a later constant? | Determines loader initialization and early-read conformance expectations. |
| How are non-finite numeric literals such as `1e999` handled, and at which diagnostic phase? | Determines frontend acceptance and exact error fixtures. |
| When are `NombreNoDefinido` and `NombreSinInicializar` detected, and what does `check` validate? | Determines resolver/runtime boundaries, catchability, and checker promises. |
| Is Go confirmed as the implementation language, and what is its minimum supported version? | Determines host toolchain, package layout, and runnable test/build commands. |

**Planning assumption only:** the architecture uses Go-oriented paths because Go is the draft direction. Neither this roadmap nor those paths confirm the host choice. Owner approval must be recorded before implementation starts.

## Project change control

Before implementation begins, record the owner's approval of the language contract and resolutions to the four open decisions. For each milestone, compare implementation and observed acceptance evidence with the approved reference; document deviations and obtain approval before changing the language contract. A written proposal is not evidence that the language executes.

## Deferred beyond proposed 0.1

The draft defers classes, inheritance, operator overloading, macros, generics, threads, network access, package installation, JIT, native code generation, and self-hosting. Async/parallel syntax, slices, and destructuring are outside the reference's 0.1 scope. Do not introduce them through milestone implementation.

Tree-walking remains the proposed initial execution model. Bytecode is a possible later optimization only, not an eighth milestone or a commitment. Any later evolution would require separate owner-approved scope and evidence that observable language behavior remains consistent.

## Next review

Review the [language reference](language-reference.md#open-questions-before-implementation) and [architecture open decisions](architecture.md#open-decisions-before-implementation) together. Approve or resolve the draft explicitly before opening implementation work; this roadmap does not record that approval.
