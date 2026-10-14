---
description: "Use when creating, modifying, reviewing, or refactoring Go code in pacman-ai. Keep documentation comments complete and aligned with actual behavior."
applyTo: "**/*.go"
---

# Go Documentation

- Maintain brief English documentation comments for every function and method, including unexported helpers and tests. Start declaration documentation with the identifier's name.
- Document package-level constants and variables, and struct fields. A short shared comment may cover related declarations.
- Do not add documentation comments for local variables or loop variables. Inside functions, use comments only when they explain non-obvious logic, constraints, or implementation choices.
- When changing code, update its comments in the same change. After renames, moves, or delegation, describe the current owner and responsibility instead of the previous implementation.
- Keep comments accurate about units, coordinates, valid inputs, return values, errors, phase guards, and state mutations where relevant. Describe implemented behavior, not planned features.
- Preserve the concise existing style. Explain each declaration's role rather than merely restating its assignment; remove stale comments and commented-out code when they are part of the touched documentation.
- Before completing a change, check that added or modified functions, methods, package-level declarations, and struct fields remain documented and that nearby comments still match the code. Documentation-only changes must not alter runtime behavior.