Morpheo: Compute API
====================

The compute API is a simple HTTP API that accepts learning and prediction tasks
(as *learnuplets* and *preduplets*) from the orchestrator (and the orchestrator
only), validates them and puts them in a distributed task queue.
This folder contains the code for the `compute` api only. The code that runs the
`compute`