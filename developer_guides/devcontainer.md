# Building a dev instance of a container

* Use ko to build a docker image with some changes you want to test without doing a release

```bash {"id":"01JDDG5YXKP5JT14J1A8X6S6JB"}
# TODO(jeremy): Should we pute these in a different repository
cd ../app
# TODO(jeremy): WHy do we have to set ko_config_path; why isn't it picked up automatically?
KO_DOCKER_REPO=ghcr.io/jlewi/foyle \
    ko build --base-import-paths ./
```

It looks like the output of your previous `ko build` command indicated that your Git repository is in a dirty state with untracked files. Before proceeding, it's a good idea to either commit or clean up these changes. Here are the next steps you can take:

1. Check the status of your Git repository to see what files are causing the dirty state.