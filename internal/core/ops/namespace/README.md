# Operational namespaces

Valid applies the existing lowercase alphanumeric and hyphen namespace rule after
trimming whitespace. Layout, configuration and transport callers share this exact
rule. Namespace selection and directory creation remain with their callers.

Resolve selects an explicit name, OPS_USER or Default with the existing trimming
and spelling rules. UserEnv names the environment reference. Callers retain
project validation, artifact paths, selection authority and transaction ownership.
