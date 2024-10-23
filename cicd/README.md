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

* Fetch the logs for the K8s job one-off-release
* Use gcloud to fetch them from gcloud
* I noticed that in some k8s the labels k8s-pod weren't attached. I wonder if that happens because the pod and VM didn't live long enough?

```bash {"id":"01JAX9PE1ZQATY0R0RTMVAWSEQ","interactive":"false"}
# Fetch the logs for the one-off K8s job using gcloud
gcloud logging read "resource.type=\"k8s_container\" AND labels.\"k8s-pod/batch_kubernetes_io/job-name\"=\"one-off-release\"" --limit=100
```