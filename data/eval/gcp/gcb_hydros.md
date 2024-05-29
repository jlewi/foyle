Get the cloud build jobs for commit abc1234
```bash
gcloud builds list --limit=10 --format="value(ID,createTime,duration,tags,logUrl,status)" --project=foyle-public --filter="tags:commit-abc1234"
```
```output
exitCode: 0
```
```output
stdout:
54e441b3-1bb4-46d4-8d5a-72f115b2c4cb	2024-05-03T18:10:24+00:00		us-west1-docker.pkg.dev_foyle-public_images_hydros_hydros;commit-e4d040b;version-v20240503T111024	https://console.cloud.google.com/cloud-build/builds/54e441b3-1bb4-46d4-8d5a-72f115b2c4cb?project=593472949942	SUCCESS
```
