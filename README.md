This repo contains a go program to demonstrate that the controller-runtime's envtest
environment does not perform cascading deletion on an object's dependends as
as described here:
https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents

This test demonstrates this by:
- Creating two ConfigMaps (A & B)
- setting ConfigMap B as a dependent of ConfigMap A (through controllerutil.SetControllerReference())
- Deleting ConfigMap A
  This is done with the DeletePropagationForeground option, 
  so the object's deletion won't occur until its dependents have been deleted.
- Checking to see if ConfigMap B was also deleted

The -t flag can be used to toggle between using controller-runtime's EnvTest environment,
and whatever k8s cluster is present in the session (minikube for me).

On minikube, cascading deletion works as expected, in EnvTest it does not.

Example test output:
```
[rcampos@rh-laptop cr-envtest-demo]$ kubectl get node
NAME       STATUS   ROLES    AGE     VERSION
minikube   Ready    master   5d15h   v1.19.2
[rcampos@rh-laptop cr-envtest-demo]$ # Running on minikube
[rcampos@rh-laptop cr-envtest-demo]$ go run main.go
2021/03/26 14:07:07 Creating ConfigMap A
2021/03/26 14:07:07 Creating ConfigMap B
2021/03/26 14:07:07 Setting ConfigMap B to have ConfigMap A as its controller reference
2021/03/26 14:07:07 Deleting ConfigMap A
2021/03/26 14:07:07 Waiting for ConfigMap A to be deleted...
2021/03/26 14:07:08 ConfigMap A has been deleted
2021/03/26 14:07:08 Verifying that ConfigMap B was also deleted...
2021/03/26 14:07:08 ConfigMap B has been deleted
[rcampos@rh-laptop cr-envtest-demo]$ # Running on envtest environment
[rcampos@rh-laptop cr-envtest-demo]$ go run main.go -t
2021/03/26 14:07:28 Creating ConfigMap A
2021/03/26 14:07:28 Creating ConfigMap B
2021/03/26 14:07:28 Setting ConfigMap B to have ConfigMap A as its controller reference
2021/03/26 14:07:28 Deleting ConfigMap A
2021/03/26 14:07:28 Waiting for ConfigMap A to be deleted...
2021/03/26 14:07:40 Timed out waiting for ConfigMap A to be deleted
exit status 1
```
