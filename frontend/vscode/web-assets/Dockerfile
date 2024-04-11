# Should we pin this
FROM us-west1-docker.pkg.dev/foyle-public/images/vscode as build

# Build Args need to be after the FROM stage otherwise they don't get passed through to the RUN statment
ARG VERSION=unknown
ARG DATE=unknown
ARG COMMIT=unknown

COPY /frontend/vscode/web-assets /workspace
RUN /workspace/build.sh

FROM scratch
COPY --from=build /assets /assets

