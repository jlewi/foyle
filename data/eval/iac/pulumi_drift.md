Show any drift in the dev infrastructure

```sh {"id":"01HZ2J3RT6G1NY3YX4H0TW5VX8"}
pulumi -C /Users/jlewi/git_foyle/iac/dev refresh -y
pulumi -C /Users/jlewi/git_foyle/iac/dev preview --diff
```