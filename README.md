This repo contains a go program to demonstrate that the controller-runtime's envtest
environment does not perform cascading deletion on an object's dependends as
is described here:
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
minikube   Ready    master   5d16h   v1.19.2
[rcampos@rh-laptop cr-envtest-demo]$ # Running on minikube
[rcampos@rh-laptop cr-envtest-demo]$ go run main.go
2021/03/26 14:33:03 Creating ConfigMap A
2021/03/26 14:33:03 Displaying ConfigMap A json:
{
 "metadata": {
  "name": "cm-a",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-a",
  "uid": "c618c261-9e84-4dee-a5d6-7454f85cda90",
  "resourceVersion": "186050",
  "creationTimestamp": "2021-03-26T18:33:03Z"
 }
}
2021/03/26 14:33:03 Creating ConfigMap B
2021/03/26 14:33:03 Displaying ConfigMap B json:
{
 "metadata": {
  "name": "cm-b",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-b",
  "uid": "7acef85d-bf8f-49cd-bd9c-de78ae0a7434",
  "resourceVersion": "186051",
  "creationTimestamp": "2021-03-26T18:33:03Z"
 }
}
2021/03/26 14:33:03 Setting ConfigMap B to have ConfigMap A as its controller reference
2021/03/26 14:33:03 Displaying updated ConfigMap B json:
{
 "metadata": {
  "name": "cm-b",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-b",
  "uid": "7acef85d-bf8f-49cd-bd9c-de78ae0a7434",
  "resourceVersion": "186052",
  "creationTimestamp": "2021-03-26T18:33:03Z",
  "ownerReferences": [
   {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "name": "cm-a",
    "uid": "c618c261-9e84-4dee-a5d6-7454f85cda90",
    "controller": true,
    "blockOwnerDeletion": true
   }
  ]
 }
}
2021/03/26 14:33:03 Deleting ConfigMap A
2021/03/26 14:33:03 Waiting for ConfigMap A to be deleted...
2021/03/26 14:33:04 ConfigMap A has been deleted
2021/03/26 14:33:04 Verifying that ConfigMap B was also deleted...
2021/03/26 14:33:04 ConfigMap B has been deleted
```


Running on envtest environment, note how the program hangs while deleting ConfigMap A.
This is due to foreground cascading deletion being set, but ConfigMap A's dependents
are not being properly deleted by envtest.
```
[rcampos@rh-laptop cr-envtest-demo]$ # Running on envtest environment
[rcampos@rh-laptop cr-envtest-demo]$ go run main.go -t
2021/03/26 14:33:40 Creating ConfigMap A
2021/03/26 14:33:40 Displaying ConfigMap A json:
{
 "metadata": {
  "name": "cm-a",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-a",
  "uid": "dc3cb059-77ed-428b-b535-eed024a65935",
  "resourceVersion": "44",
  "creationTimestamp": "2021-03-26T18:33:40Z"
 }
}
2021/03/26 14:33:40 Creating ConfigMap B
2021/03/26 14:33:40 Displaying ConfigMap B json:
{
 "metadata": {
  "name": "cm-b",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-b",
  "uid": "a02a6c09-e5d3-4916-821a-7f78b5da8530",
  "resourceVersion": "45",
  "creationTimestamp": "2021-03-26T18:33:40Z"
 }
}
2021/03/26 14:33:40 Setting ConfigMap B to have ConfigMap A as its controller reference
2021/03/26 14:33:40 Displaying updated ConfigMap B json:
{
 "metadata": {
  "name": "cm-b",
  "namespace": "default",
  "selfLink": "/api/v1/namespaces/default/configmaps/cm-b",
  "uid": "a02a6c09-e5d3-4916-821a-7f78b5da8530",
  "resourceVersion": "46",
  "creationTimestamp": "2021-03-26T18:33:40Z",
  "ownerReferences": [
   {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "name": "cm-a",
    "uid": "dc3cb059-77ed-428b-b535-eed024a65935",
    "controller": true,
    "blockOwnerDeletion": true
   }
  ]
 }
}
2021/03/26 14:33:40 Deleting ConfigMap A
2021/03/26 14:33:40 Waiting for ConfigMap A to be deleted...
2021/03/26 14:33:53 Timed out waiting for ConfigMap A to be deleted
exit status 1
```
