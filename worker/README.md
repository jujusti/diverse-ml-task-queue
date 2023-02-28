Morpheo: Compute Workers
========================

Compute workers prepare and execute containerized machine learning workflows.

It retrieves tasks from a (distributed) broker, pulls the *problem workflow*
container (that describes how training and prediction tasks are executed and
evaluated) and runs the training/prediction tasks on the network-isolated
*submission* container. Training tasks' performance increase is also evaluated
and sent to the orchestrator.

The specifications of the containers ran by compute is documented
[here](https://morpheoorg.github.io/morpheo/).

Examples *problem workflow* and *submission* containers can be found
[here](https://github.com/MorpheoOrg/hypnogram-wf).

CLI Arguments
-------------

```
Usage of compute-worker:

  -docker-timeout duration
    	Docker commands timeout (concerns builds, runs, pulls, etc...) (default: 15m) (default 15m0s)
  -learn-parall