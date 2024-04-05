# foyle vscode 

This the vscode extension that is the frontend for foyle.

The extension was initialized using the [vscode extension generator](https://code.visualstudio.com/api/get-started/your-first-extension).

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

# References
[VSCode Web Extensions](https://code.visualstudio.com/api/extension-guides/web-extensions)

