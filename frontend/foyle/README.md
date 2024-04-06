# foyle vscode 

This the vscode extension that is the frontend for foyle.

The extension was initialized using the [vscode extension generator](https://code.visualstudio.com/api/get-started/your-first-extension).

see [vscode_apis.md](vscode_apis.md) for explanation of the vscode apis.

## Developer Guide

* In vscode for web you can go to the extensions tab and search for "@builtin" to see if the extension is loaded

* The engines field in the [package.json manifest](https://code.visualstudio.com/api/references/extension-manifest) 
  controls the vscode version compatibility with the extension. If you install the extension in an incompatible
  version of vscode it won't load.

### Setup

### Setup your typsecript/npm/toolchain

You need npm and node installed. 

Install typescript and yarn

```
npm install -g typescript
npm install -g yarn
```

Install npm dependencies

```
yarn install
```

Compile it for web

```
yarn package-web
```

## Run the tests

```
yarn test
```

## Running the extension

```
yarn compile-web
yarn run-in-browser
```

**Important** You need to rerun compile-web to pick up the latest changes. I don't think you need to restart the server
you can just do referesh.

## vscode-test-web

This is the test server for vscode for web. You can just run it like so


```
yarn vscode-test-web --help
```

The source is in [https://github.com/microsoft/vscode-test-web](https://github.com/microsoft/vscode-test-web)

The extension ends up being served at

```
http://localhost:3000/static/devextensions/package.json
```

It doesn't look like the extension is active by default. However, you can go to the command pallet and search
for the Foyle commands. When you click one of the foyle comands it will activate the extension and you'll see a
message.

For more info see [Test Your Web Extensions using vscode-test-web](https://code.visualstudio.com/api/extension-guides/web-extensions#test-your-web-extension-in-a-browser-using-vscodetestweb)

### Limitations

When using `vscode-test-web` files aren't actually persisted to disk. I'm unclear whether this is an issue with the
server or the chrome instance that gets created. I suspect vscode-test-web might be launching a virtual filesystem
which is why we aren't seeing the files on disk.


### Connecting to the backend

You can use the development version of the extension that you run using `yarn run-in-browser` to connect to 
a backend running on your local machine. However, sine vscode is running on a different origin then the backend
you will need to configure the backend to allow cross origin requests. To do this edit `${HOME}/.foyle/config.yaml`
and add a section like this:

```yaml
server:
  cors:
    allowedOrigins: 
      - http://localhost:8080
    allowedHeaders:
      - Content-Type
      - Referrer
      - User-Agent
    vsCodeTestServerPort: 3000
```

This will enable CORS for the vscode test server running on port 3000. 

**important** The way `vscode-test-web` works, Chrome will end up generating a random referrer used as the origin.
Even though test server is running at `http://localhost:3000` the referrer will be something like 
`http://v--1f1nq97ha8fjnonnlusc36olp1p7do9ddp5bnr0apu6pt4phaoq0.localhost:3000`. Where the *hostname* appears to be
a randomly generated string. I suspect this a security feature of Chrome. The backend has special CORS handling
for this case to allow it to be used with the development server. This code path is enabled by setting 
`vsCodeTestServerPort` to a non zero value..

After updating your config, start the foyle server like so:

```
foyle serve
```

Then start the vscode extension as described in the previous section

```
yarn run-in-browser
```

If your foyle server isn't running on the default port of 8080 then in vscode you will need to open up settings
and configure foyle to use the server running on whatever port you are running on.

## WebViews, CORS, and `vscode.cdn.net`

Some assets still get pulled in from vscode.cdn.net. I haven't fully figured this out yet.
See [vscode-discussions/discussions](https://github.com/microsoft/vscode-discussions/discussions/985)

Per this [doc](https://code.visualstudio.com/api/extension-capabilities/extending-workbench#webview) I think the webview is where the markdown gets rendered

Per the  [webview extension docs](https://code.visualstudio.com/api/extension-guides/webview); I think they run in an iframe. I suspect the code they are pulling down maybe coming from "vscode-cdn.net". Therefore, when that code ends up trying to contact my server we hit a CORS issue.

I think the URL for that code might get set here
https://github.com/microsoft/vscode/blob/294eda9b7bd2fc70c174f9ce5e9bb87db5b4d1e6/product.json#L33

See the comment in https://vscodium.com/ about product.json.

# References

[vscode-discussions/discussions](https://github.com/microsoft/vscode-discussions/discussions/985)
  * Explains why for security reasons, WebView must be isolated from the host of vscode. 
  * WebView must have a different origin from the host. So, webview contents are hosted on *.vscode-cdn.net for VS Code Web.
  * Has an impact on how we setup CORS 
[VSCode Web Extensions](https://code.visualstudio.com/api/extension-guides/web-extensions)
