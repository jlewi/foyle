List the GCB jobs that build image backend/caribou
```bash
gcloud builds list --limit=10 --filter='tags="us-west1-docker.pkg.dev_foyle-public_images_backend_caribou"' --format="value(ID,createTime,duration,tags,logUrl,status)" --project=foyle-public
```
```output
exitCode: 0
```
