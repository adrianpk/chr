# Slug values

Normalize retains Unicode letters and digits and replaces other character runs
with one dash. Leading and trailing dashes are removed. Callers retain their
own case policy. Artifact names and model tags share the original rule through
this package; it performs no IO or model decisions.

ArtifactName applies the original normalization and 48-byte artifact basename
bound shared by ticket and authoring layout. Callers retain numbering, collision
and create-only policy.
