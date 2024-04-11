# This is based off of the devcontaienr for vscode.
# https://github.com/microsoft/vscode/blob/main/.devcontainer/Dockerfile
# The main difference is that we don't install vscode in the container; i.e. we don't call the install-vscode.sh script.
FROM mcr.microsoft.com/devcontainers/typescript-node:18-bookworm as build

# Build Args need to be after the FROM stage otherwise they don't get passed through to the RUN statment
ARG VERSION=unknown
ARG DATE=unknown
ARG COMMIT=unknown

# Not sure why this is needed since its not in the dev container.
RUN apt-get update -y && \
  apt-get install -y libxkbfile-dev

RUN npm install -g node-gyp

COPY /frontend/foyle /foyle
WORKDIR /foyle
RUN /foyle/build.sh

FROM scratch

COPY --from=build /foyle/package.json /foyle/package.json
COPY --from=build /foyle/package.nls.json /foyle/package.nls.json
# version.json gets builts in build.sh
COPY --from=build /foyle/version.json /foyle/version.json
COPY --from=build /foyle/dist /foyle/dist