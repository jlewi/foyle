Get the logs for building the image carabou
```bash
gcloud logging read 'logName="projects/foyle-dev/logs/hydros" jsonPayload.image="carabou"' --freshness=1d  --project=foyle-dev
```
