Draft for a resilient scheduler, that attempts to restart a task on errors.

Backoff delays restarts in growing intervals to prevent waste of resources caused by continuous restarts of broken tasks.