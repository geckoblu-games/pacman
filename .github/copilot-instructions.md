# Project Guidelines

## Documentation and Gameplay Requirements

- Keep [the implementation reference](../doc/PROJECT%20IMPLEMENTATION%20REFERENCE.md) aligned with changes to behavior, ownership, and project structure. Describe actual implementation and choices; do not present planned features as implemented.
- Treat [the gameplay design dossier](../doc/Pacman_Clone_Design_Dossier.md) as the requirements baseline. Before implementing a mechanic, check its FACT, ASSUMPTION, and AMBIGUITY labels and the relevant sections of [The Pac-Man Dossier](../doc/The%20Pac-Man%20Dossier.md).
- Identify discrepancies between proposed or implemented behavior and the gameplay requirements. Distinguish intentional choices from arcade-fidelity deviations, preserving documented arcade quirks when fidelity is the goal.
- Ask the project owner when an unresolved gameplay ambiguity changes observable behavior. After a decision, update the design dossier so it remains authoritative.
- [The Go/Ebitengine implementation proposal](../doc/go-ebiten-implementation-proposal.md) is non-binding. Evaluate its ideas against the gameplay dossier and existing code; do not adopt its package tree or abstractions automatically.

## Implementation Boundaries

- Keep gameplay simulation deterministic and renderer-independent. Use tile-coordinate logic, fixed 60 Hz simulation ticks, shared ghost pathfinding, level parameter tables, and tests for observable behavior as those systems are implemented. Ebitengine presents simulation state; it must not own game rules.
- Prefer a small, incremental Go layout. Add packages only when they provide a real ownership or testability boundary; a request for a new organizational file does not by itself require a new package.

## Go Comments

- Follow [the Go documentation instructions](instructions/go-documentation.instructions.md): maintain brief, accurate comments for functions, methods, package-level constants and variables, and struct fields, including tests. Do not document local or loop variables; reserve internal comments for non-obvious logic.