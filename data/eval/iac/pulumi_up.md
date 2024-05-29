```bash {"id":"01HZ2GTQAA8ECXR36P75BTCPB9"}
Sync the dev infra
```

```bash {"id":"01HZ2GTQAA8ECXR36P75EJBGK4"}
pulumi -C /Users/jlewi/git_foyle/iac/dev refresh -y
pulumi -C /Users/jlewi/git_foyle/iac/dev up --diff --skip-preview --non-interactive
```
