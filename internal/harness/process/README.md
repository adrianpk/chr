# Process supervision

Owns bounded subprocess streams, cancellation, process-group cleanup and
sanitized diagnostics. Each call spools its output to private temporary files
and removes them after reaping the owned process. Stream overflow fails the
call; retained diagnostic truncation never creates an accepted JSON result.
