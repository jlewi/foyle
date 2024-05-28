# Developing the Runme Extension

* The Runme extension is in the [vscode-runme](https://github.com/stateful/vscode-runme) repository
* The service is defined in [runme/pkg/api/proto/runme/ai](https://github.com/stateful/runme/tree/main/pkg/api/proto/runme/ai)
* Follow [RunMe's vscode contributing.md](https://github.com/stateful/vscode-runme/blob/main/CONTRIBUTING.md)
* If you need nvm you can brew install it

```sh {"id":"01HY2569DM0SR533BT4ZJTD2WV"}
brew install nvm
```

* The command inside Runme's contributing guide assumed vscode's binary was on the path; for me it wasn't so I had to execut
   the command using the full path.

```sh {"id":"01HY2584G3Q0A89TK1NRWVH0ZN"}
jq -r ".recommendations[]" .vscode/extensions.json | xargs -n 1 /Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code --force --install-extension
```

## Building and installing the extension from source

* [VSCode Extension Packaging & Publishing](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
* It looks like the package has a `bundle` command that will build the extension and package it into a `.vsix` file

```sh {"id":"01HY25HEG7CR7QCGJSERF3BB4K"}
cd ~/git_vscode-runme
npm run bundle
```

```sh {"id":"01HY25KVHCN2P1W9NV0ECD1TW0"}
ls -la ~/git_vscode-runme/
```

Now we can install the extension using the vscode binary

* I had to uninstall the RunMe extension first before installing the new one
* If you search for the extension in the extensions view, you can click on the arrow next to the update button and uncheck auto-update
   if you don't do that it may continue to auto update
* The exact file will be named `runme-X.Y.Z.vsix` so it will change as the version changes
* You can bump the version in `package.json` to something like `"version": "3.5.7-dev.0",` and then do the build and install
   * **Note**: Your version number should be higher than whats in the vscode marketplace otherwise vscode
      does some odd version magic
   * The advantage of this is that then you can tell which version of the extension you have installed
   * It also seemed like when I didn't bump the version I might have actually been using an old version of the extension

```bash {"id":"01HYZVG8KZKYSTFS4R1RJZDS7P"}
/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code --force --install-extension ~/git_vscode-runme/runme-3.5.9.vsix
```

```sh {"id":"01HY264KZTS4J9NHJASJT1GYJ7"}
ls -la ~/git_vscode-runme/dist
```

So it looks like my runme install is messed up
Lets try installing and reinstalling it

* Now everything is working; I can generate completions using Foyle

```bash {"id":"01HY74YTEZDZVJYPMB0VMCE84S"}
/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code --uninstall-extension stateful.runme

```

```bash {"id":"01HY75KYKE3SFAM5EXMDAVJDTQ"}
echo "hello world"
```

## Debugging the Runme Extension in vscode

* It seems like you may need to run `yarn build` for changes to get picked up; running `F5` doesn't always seem to work
* Console logs will show up in the `debug console` in the development workspace; not the instance of vscode that gets launched to run
   your extension