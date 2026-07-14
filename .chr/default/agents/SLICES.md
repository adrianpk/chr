# Slices

Use small slices for implementation work.

## Slice Rules

- One slice should have one main behavioral goal.
- Each slice should have an explicit scope and validation path.
- Keep commits aligned to concrete tasks within the slice.
- Avoid mixing planning artifacts and runtime code unless the slice explicitly owns both.

## Suggested Flow

1. Create or update plan and tracker.
2. Create a branch from the project trunk.
3. Implement the slice.
4. Run validation.
5. Open a review request.
6. Merge only after approval.
