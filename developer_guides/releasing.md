# Releasing foyle

* We use [hydros](https://github.com/jlewi/hydros/blob/main/docs/continuous_delivery.md) to release foyle

```sh {"id":"01HYPCGWG3JK9XC72V5D9YATJ0"}
cd ~/git_foyle
hydros apply releasing.yaml
```

### Verify the release

* Use the gh CLI to list the releases

```bash {"id":"01HYPDMFHKRSD8E2C0RFEK8007"}
gh release list
```

Describe the latest release to see its artifacts

To describe the latest release and see its artifacts, you can use the following command with the `gh` CLI:

```bash {"id":"01HYPDPKN79GV0WSWH1570MWPR"}
gh release view v0.0.10
```

* List the images in the GHCR repository

```bash {"id":"01HYPDS28M0P2VDWSB7MRBBJT1"}
gcrane ls ghcr.io/jlewi/foyle-vscode-ext
```