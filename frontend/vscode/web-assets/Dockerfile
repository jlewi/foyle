# Should we pin this
FROM us-west1-docker.pkg.dev/foyle-public/images/vscode as build

COPY /frontend/vscode/web-assets /workspace
RUN /workspace/build.sh

FROM scratch
COPY --from=build /assets /assets

