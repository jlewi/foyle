Sync the manifests to the dev cluster

```sh {"id":"01HZ315CWSZPNMJCRJJFMB4S8S"}
flux reconcile kustomization dev-cluster --with-source
```