# Build a docker image With vscode Web Assets

We build a datacontainer that contains just the vscode web assets. Needed for serving.
We do this in two steps

1. Build a docker container containing the compiled vscode web assets
2. Build a docker container that copies the compiled vscode web assets from the first container

The reason for doing this in two steps is because building vscode is slow. It takes 45 minutes on a 
E2_HIGHCPU_32 machine.

Furthermore, its been a bit of a trial and error process to figure out what assets we actually need to include and
serve. Caching the assets in a container makes it easier to iterate on determining which assets are actually needed.
