# Process supervision

Owns bounded subprocess streams, cancellation, process-group cleanup and
sanitized diagnostics. Each call spools its output to private temporary files
and removes them after reaping the owned process. Stream overflow fails the
call; retained diagnostic truncation never creates an accepted JSON result. On
Linux, SIGKILL cleanup waits up to one second for remaining group members to
stop executing before returning; zombies are left for their owning reaper.

`Executor.Exchange` supports protocols that need concurrent input and output.
Its handler owns framing, while the same process owner bounds and spools both
streams, closes stdin for shutdown, applies a finite exit grace, and reaps the
process group. Internal failure cleanup does not imply caller cancellation.

An independently owned Linux worker bridge can enable child adoption and reap
model tools that detached from the original process group. This capability is
never enabled on a shared application host. Reaping completes before the bridge
releases its existing runtime capacity reservation.
