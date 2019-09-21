# node-label-controller

Kubernetes controller to label nodes running "Container Linux" operating system.

If a node is running CoreOS Container Linux, it will be labeled with "kubermatic.io/uses-container-linux" = "true".

By modifying main.go you can set any other criteria based on the Node object and any other label.

## Install

```sh
kubectl apply -f deploy
```

Controller will be installed in *acim* namespace.

## Uninstall

```sh
kubectl delete -f deploy
```
