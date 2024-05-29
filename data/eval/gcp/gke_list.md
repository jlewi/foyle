Describe the dev cluster?
```bash
gcloud container clusters describe --region=us-west1 --project=foyle-dev dev
```
```output
exitCode: 1
```
```output
stderr:
ERROR: (gcloud.container.clusters.describe) ResponseError: code=404, message=Not found: projects/foyle-dev/locations/us-west1/clusters/dev.
No cluster named 'dev' in foyle-dev.
```
