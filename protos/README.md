# Protocol buffers

We use protocol buffers as a way of defining APIs and data resources.

## Developer Guide

Language bindings are generated using [buf](https://buf.build/docs/introduction)

```sh {"id":"01J2H55T7BEVFCJV9A9QX55FTR"}
buf generate
```

In the past we've run into issues generating grpc bindings for python;
see [bufbuild/buf#1344](https://github.com/bufbuild/buf/issues/1344). As of right
now we aren't using grpc in python so hopefully we don't need to worry about this and
if we do hopefully the issue has been resolved by now.

In order to be able to generate typescript protos using protobuf es I had to run

```sh {"id":"01J2H55T7BEVFCJV9A9VN0YK7G"}
npm install -g @bufbuild/protoc-gen-es 
```

See [protobuf-es](https://github.com/bufbuild/protobuf-es). I think I had to add the "-g" option because we aren't
in a typescript project in our buf project. We didn't need the other commands because we installed buf via
homebrew or some other means.

I ran

```bash {"id":"01J2H55T7BEVFCJV9A9Y1VSEVA"}
npm install @bufbuild/protobuf @bufbuild/protoc-gen-es @bufbuild/buf
```

inside `frontend/foyle` because I think that adds needed packages to my extension project

### Installing the Connect Typescript Plugin

I tried

```bash {"id":"01J2H5K0PBVEZ8SJ5KBPET0M88"}
npm install  -g @bufbuild/protoc-gen-connect-es
```

* That ran and seems to have put the plugin in `/Users/jlewi/.nvm/versions/node/v18.19.0/bin/protoc-gen-connect-es`
* It looks like `protoc-gen-es` is in /opt/homebrew/bin/protoc-gen-es
   * How did that happen?

## Typescript Protos

We are using [protobuf-es](https://github.com/bufbuild/protobuf-es)  which is from the makers of buf.
This is a different implementation of protobuf  for typescript meant to solve a lot of the problems with protos
and typescript see [blog](https://buf.build/blog/protobuf-es-the-protocol-buffers-typescript-javascript-runtime-we-all-deserve).

More documentation can be found [here](https://github.com/bufbuild/protobuf-es/blob/main/docs/generated_code.md)

```bash {"id":"01JEVJYRQKKF8Q8DX7SSX6KGW0"}
# To confirm the installation of the required packages and plugins, run the following commands:
npm list @bufbuild/protobuf @bufbuild/protoc-gen-es @bufbuild/buf
npm list -g @bufbuild/protoc-gen-connect-es
```

# GoLang

* Install the plugin below to get `protoc-gen-go` 

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

## connect-go

To install the connect-go plugin

```sh {"id":"01J2H55T7BEVFCJV9AA06C6GR5"}
 go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

## Zap Marshalers

We use [zap marshalers](https://pkg.go.dev/go.uber.org/zap#hdr-JSON) to generate MarshalLogObject methods for our types.
We can then log protos doing

```go {"id":"01J2H55T7BEVFCJV9AA1SB6ACP"}
logger.Info("message", zap.Object("proto", proto))
```

This ensures the proto is logged in the json encoding of the proto. If we don't then I'm not sure what the schema is

See also [This SO](https://stackoverflow.com/questions/68411821/correctly-log-protobuf-messages-as-unescaped-json-with-zap-logger)

To install the plugin

```bash {"id":"01J2H55T7BEVFCJV9AA5QCWGPX"}
go install github.com/kazegusuri/go-proto-zap-marshaler/protoc-gen-zap-marshaler@latest
```

Ensure the plugin is in your path; otherwise buf won't be able to find it.

## TODO GRPC-connect

We should look into [connect-rpc](https://connectrpc.com/). That might simplify things

# Buf Schema Registry

Refer to the [docs](https://buf.build/docs/bsr/module/publish#module-and-repository-setup)

1. Login if you haven't already

```bash {"id":"01J463MR2W87MNYGHKV6XANBM7"}
buf registry login
```

```bash {"id":"01J464214VYQ224BYDJZEQFAGQ"}
buf dep update
buf build
buf push

```

## Push a development version

* To push a development version add the label flag
* In the UI this labels show up in the upper right corner of the window

```bash {"id":"01J6CP47NS6RA0MXEME8CX5AFV"}
buf build
buf push --label=dev
```

# Developing The VSCode Extension

There are two ways I think we can iterate on the vscode extension when making changes to the proto

1. Publish to the [BSR using labels](https://buf.build/docs/bsr/module/publish#pushing-from-a-local-workspace)

   * I think we can use labels to create the equivalent of a development branch of the SDK

2. We can update `buf.gen.yaml` to output to a local path inside the vscode extension directory

   * See [buf.gen.yaml](https://github.com/jlewi/foyle/blob/9663fb81a36ab63876c33873cf4726dc8ef80092/protos/buf.gen.yaml#L28)

# Troubleshooting

## Error: In Typescript Type Is Not Assignable

In vscode-runme I'd gotten

```ts {"id":"01J6FX93E4T2E0TPG5KX7GBT1Q"}
ERROR in ./src/extension/ai/events.ts:66:33
TS2345: Argument of type 'LogEventsRequest' is not assignable to parameter of type 'PartialMessage<LogEventsRequest>'.
  Types of property 'events' are incompatible.
    Type 'LogEvent[]' is not assignable to type 'PartialMessage<LogEvent>[]'.
      Type 'LogEvent' is not assignable to type 'PartialMessage<LogEvent>'.
        Types of property 'type' are incompatible.
          Type 'LogEventType' is not assignable to type 'LogEventType | undefined'.
    64 |     const req = new LogEventsRequest()
    65 |     req.events = events
  > 66 |     await this.client.logEvents(req).catch((e) => {
       |                                 ^^^
    67 |       this.log.error(`Failed to log event; error: ${e}`)
    68 |     })
    69 |   }

```

* I had updated the bufbuild/es package but not the connect/es package
* When Up updated the connect/es package the error went away

```bash {"id":"01J6CQT3CYQ76FRD1GX258JAW2"}
cd /path/to/vscode-runme
npm update @buf/stateful_runme.community_timostamm-protobuf-ts
npm install
```