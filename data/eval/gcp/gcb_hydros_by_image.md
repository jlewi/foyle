List the GCB jobs that build image backend/caribou

```bash {"id":"01HZ2GP8BFCQ5BKGHEYRFW44YD"}
gcloud builds list --limit=10 --filter='tags="us-west1-docker.pkg.dev_foyle-public_images_backend_caribou"' --format="value(ID,createTime,duration,tags,logUrl,status)" --project=foyle-public
```
