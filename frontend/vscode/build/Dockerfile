# Container for building the vscode web assets and publishing them in a container.
# This is based off of the devcontaienr for vscode.
# https://github.com/microsoft/vscode/blob/main/.devcontainer/Dockerfile
# The main difference is that we don't install vscode in the container; i.e. we don't call the install-vscode.sh script.
FROM mcr.microsoft.com/devcontainers/typescript-node:18-bookworm

# Not sure why this is needed since its not in the dev container.
RUN apt-get update -y && \
  apt-get install -y libxkbfile-dev

# N.B. If we switch to user node here we won't have permission to create the vscode directory.
# Should we create a "/workspace" dir and do everything in there and give it permission vscode?
# USER node
RUN npm install -g node-gyp

COPY /frontend/vscode/build/build.sh /build.sh
RUN /build.sh