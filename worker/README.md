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
  -learn-parallelism int
    	Number of learning task that this worker can execute in parallel. (default 1)
  -learn-timeout duration
    	After this delay, learning tasks are timed out (default: 20m) (default 20m0s)
  -nsqlookupd-urls value
    	URL(s) of NSQLookupd instances to connect to
  -orchestrator-host string
    	Hostname of the orchestrator to send notifications 