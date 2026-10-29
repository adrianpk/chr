# Document catalog frames

Producer: `documentCatalogFrames` using the production TUI renderer. Plan frames
retain the original ef54c01fec17 baseline. Spec frames include concrete document
titles, the complete catalog's empty state and manual Create/Edit hints.
The frozen frames cover both artifact kinds, empty/missing/generated/busy states,
selected records, the plain screen and all maximized-pane layouts.
The fixture is read-only during validation; the test compares complete strings.
