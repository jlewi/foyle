---
runme:
  document:
    relativePath: devcontainer.md
  session:
    id: 01JDDK3R7CDQP3ZBEM1TQRK466
    updated: 2024-11-23 15:10:17-08:00
---

# Building a dev instance of a container

* Use ko to build a docker image with some changes you want to test without doing a release

```bash {"id":"01JDDG5YXKP5JT14J1A8X6S6JB"}
# TODO(jeremy): Should we pute these in a different repository
cd ../app
# TODO(jeremy): WHy do we have to set ko_config_path; why isn't it picked up automatically?
KO_DOCKER_REPO=ghcr.io/jlewi/foyle \
    ko build --base-import-paths ./

# Ran on 2024-11-23 15:09:59-08:00 for 13.913s exited with 0
2024/11/23 15:10:01 Using base cgr.dev/chainguard/static:latest@sha256:5ff428f8a48241b93a4174dbbc135a4ffb2381a9e10bdbbc5b9db145645886d5 for github.com/jlewi/foyle/app
2024/11/23 15:10:02 Using build config foyle for github.com/jlewi/foyle/app
2024/11/23 15:10:04 git is in a dirty state
Please check in your pipeline what can be changing the following files:
AM app/.ko.yaml
 M app/Dockerfile
 M app/pkg/analyze/session_manager.go
 M manifests/statefulset.yaml
?? developer_guides/devcontainer-01JDDK3R7CDQP3ZBEM1TQRK466.md
?? developer_guides/devcontainer.md

2024/11/23 15:10:04 Building github.com/jlewi/foyle/app for linux/amd64
2024/11/23 15:10:10 Publishing ghcr.io/jlewi/foyle/app:latest
2024/11/23 15:10:10 existing blob: sha256:31395f960e91fef49bf76d6ef18ab0b06f538025931d1cc57d6e729487da71c4
2024/11/23 15:10:10 existing blob: sha256:250c06f7c38e52dc77e5c7586c3e40280dc7ff9bb9007c396e06d96736cf8542
2024/11/23 15:10:11 pushed blob: sha256:b63296218d0e42ae111b040e4f88e45416291572612eee7b905300f52ba7b92a
2024/11/23 15:10:11 pushed blob: sha256:80953368c6fb45c3f2761dde20c75785aa31343a7b2c1307ce36e4a7bd82da93
2024/11/23 15:10:11 pushed blob: sha256:e0be50e8b3b26f747fb8923cabac200955fcd31b2a710febd812226e3ce4b2ad
2024/11/23 15:10:12 ghcr.io/jlewi/foyle/app:sha256-9be786a0993fc946df13618a771ec90b5c7f9d082e37ba665dbf2a8541682ae0.sbom: digest: sha256:6e7f4b619c5d31f2495d701c9bd80c8fb06580efbd37be0a7b5f8324e7889bc5 size: 375
2024/11/23 15:10:12 Published SBOM ghcr.io/jlewi/foyle/app:sha256-9be786a0993fc946df13618a771ec90b5c7f9d082e37ba665dbf2a8541682ae0.sbom
2024/11/23 15:10:13 pushed blob: sha256:1733c90cb8f756f25ef803fb450f8c26c2be9cb63e629cd2cde5538e54dddd14
2024/11/23 15:10:13 ghcr.io/jlewi/foyle/app:latest: digest: sha256:9be786a0993fc946df13618a771ec90b5c7f9d082e37ba665dbf2a8541682ae0 size: 1337
2024/11/23 15:10:13 Published ghcr.io/jlewi/foyle/app@sha256:9be786a0993fc946df13618a771ec90b5c7f9d082e37ba665dbf2a8541682ae0
ghcr.io/jlewi/foyle/app@sha256:9be786a0993fc946df13618a771ec90b5c7f9d082e37ba665dbf2a8541682ae0

```

It looks like the output of your previous `ko build` command indicated that your Git repository is in a dirty state with untracked files. Before proceeding, it's a good idea to either commit or clean up these changes. Here are the next steps you can take:

1. Check the status of your Git repository to see what files are causing the dirty state.