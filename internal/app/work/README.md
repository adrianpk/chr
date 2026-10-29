# Work observations

This application-owned package contains transport-neutral snapshots for queue,
readiness, current work, retained results and activity. `app.service.ObserveWork`
produces them without starting, claiming or recovering work. Adapters may render
these values; a queue row is never an executable scheduler input.

Queue pages retain at most 128 rows, current contexts and events at most 128,
and terminal results at most 16. Requests filter and order the shared queue
before pagination. Executable requests still pass the canonical admission owner.
