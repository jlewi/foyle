Get the cloud build jobs for commit abc1234

```bash {"id":"01HZ2GPZR77SNCQ83KV7HB9Z2M"}
gcloud builds list --limit=10 --format="value(ID,createTime,duration,tags,logUrl,status)" --project=foyle-public --filter="tags:commit-abc1234"
```
