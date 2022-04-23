Morpheo: Compute API
====================

The compute API is a simple HTTP API that accepts learning and prediction tasks
(as *learnuplets* and *preduplets*) from the orchestrator (and the orchestrator
only), validates them and puts them in a distributed task queue.
This folder contains the code for the `compute` api only. The code that runs the
`compute` workers lives in the `compute-worker` folder and is documented there
as well.

API Spec
--------

The API is dead simple. It consists in 4 routes, two of them being completely
trivial:
 * `GET /`: lists all the routes
 * `GET /health`: service liveness probe
 * `POST /pred`: post a preduplet to this route
 * `POST 