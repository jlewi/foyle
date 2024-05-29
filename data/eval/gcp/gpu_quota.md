Check for preemptible A100 quota in us-central1

```bash {"id":"01HZ2GRHWRZG8BT9CANA3G2RH6"}
gcloud compute regions describe us-central1 --format=json | jq '.quotas[] | select(.metric | contains("NVIDIA_A100"))'
```
