# go adminz pages

A simple set of adminz pages for use in go services.

## Healthz

Adds a handler for `/healthz` that returns whether a server is OK or not. "OK"
is defined as a set of conditions:

* There is not a `killfile` AND
* healthy() is unset (and we ignore it) OR
* healthy() returns true

A `killfile` is a file on disk that indicates a server should appear to be
unhealthy. Generally this is used during shutdown or startup or during
a maintance period. `Killfiles()` is provided to generate a default set.
Killfiles are checked every second for existence.

`Pause()` is called when the service first sees a killfile.

`Resume()` is called when the service sees the killfile go away.

## Servicez

Adds a handler for `/servicez` that returns some JSON. Generally this would be
information about what the server does or perhaps some configurations or
similar.
