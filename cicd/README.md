# Foyle CICD

* We use a cronjob to regularly run [hydros](https://github.com/jlewi/hydros) to release Foyle

```bash
kustomize build releaser | kubectl apply -f -
```

## Create one off job

* You can fire off a job from the cron job

```bash {"id":"01JAX9F0VQW4RWK80WNCNANSWD","interactive":"true"}
kubectl create job --from=cronjob/release-cron one-off-release -n foyle-cicd
```

```bash {"id":"01JAX9G3MYCXPPHJP06G0CWDPY","interactive":"false"}
# 1. Check the status of the one-off job and its pods to ensure everything is running correctly.
kubectl -n foyle-cicd get jobs
kubectl -n foyle-cicd get pods -n foyle-cicd
```

```bash
kubectl -n foyle-cicd get pods -w
```

```bash {"id":"01JAX9ENHC429F63VT3RC11A49","interactive":"false"}
kubectl create cronjob foyle-releaser --image=us-west1-docker.pkg.dev/foyle-public/images/releaser --schedule="0 * * * *" -- /bin/sh -c "kustomize build releaser | kubectl apply -f -"
```