# Protocol buffers

We use protocol buffers as a way of defining APIs and data resources.

## Developer Guide

Language bindings are generated using [buf](https://buf.build/docs/introduction)

```
buf generate
```

In the past we've run into issues generating grpc bindings for python; 
see [bufbuild/buf#1344](https://github.com/bufbuild/buf/issues/1344). As of right
now we aren't using grpc in python so hopefully we don't need to worry about this and
if we do hopefully the issue has been resolved by now.


## GRPC-Gateway

We [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to generate RESTful services from our grpc services.
