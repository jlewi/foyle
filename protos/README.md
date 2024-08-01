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

## GRPC-Gateway

We [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to generate RESTful services from our grpc services.

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

# To Install the connect typescript plugin