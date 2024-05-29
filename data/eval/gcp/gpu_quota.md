Check for preemptible A100 quota in us-central1
```bash
gcloud compute regions describe us-central1 --format=json | jq '.quotas[] | select(.metric | contains("NVIDIA_A100"))'
```
