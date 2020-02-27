# Cluster Operator
Project that provisions kuberneres (k8s) cluster using k8s 
[operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

The project is build around 
[Operator SDK Framework](https://github.com/operator-framework/operator-sdk).
Reference its
[Prerequisites](https://github.com/operator-framework/operator-sdk#prerequisites) 
and how to
[install Operator SDK CLI](https://github.com/operator-framework/operator-sdk#install-the-operator-sdk-cli)
if you are going to extend this project.

## Development

The base project was created using:
```bash
operator-sdk new cluster-operator
cd cluster-operator
operator-sdk add api --api-version=cluster-operator.seizadi.github.com/v1alpha1 --kind=Cluster
operator-sdk add controller  --api-version=cluster-operator.seizadi.github.com/v1alpha1 --kind=Cluster
```

You can use 
[kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
or 
[minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) for development
```bash
kind create cluster
```
