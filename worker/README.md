Morpheo: Compute Workers
========================

Compute workers prepare and execute containerized machine learning workflows.

It retrieves tasks from a (distributed) broker, pulls the *problem workflow*
container (that describes how training and prediction tasks are executed and
evaluated) and runs the training/prediction tasks on the network-isolated
*submission* container. Training tasks' performan