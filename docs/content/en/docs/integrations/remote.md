---
title: "Remote"
description: "Remote"
weight: 8
---

How to connect to a remote gRPC  runme on cloud workstations
Use grpcurl to test it out

```sh {"id":"01J1V9M6ZD4PJJ92HBRZ03M83V"}
CERTDIR=/Users/jlewi/.vscode/extensions/stateful.runme-3.5.9-darwin-arm64/tls 
grpcurl -insecure -cert ${CERTDIR}/cert.pem -key ${CERTDIR}/key.pem  localhost:7863 grpc.health.v1.Health/Check
```

In the server download and  unpack runme
 ./runme server --address localhost:7888 --runner --insecure

 * This runs it insecure mode

 * Start a tunnel from the local machine to the runme port on the remote machine

```sh {"id":"01J1VA5HCXZM7K0C7A5Q6YTAPW"}
 gcloud workstations start-tcp-tunnel --project=chainguard-workstations --region=us-central1 --cluster=work --config=work jeremy-lewi --local-host-port=localhost:7888 7888
```

Now connect it to; assuming we didn't enable TLS

```sh {"id":"01J1VA752VR18J815ETCJM0JVX"}
grpcurl --plaintext localhost:7888 grpc.health.v1.Health/Check
```

* Its not working.
* I updated vscode settings
* Specify custom address
* Disabled TLS 


* I take that back it is working
* I think the problem is its trying to run zsh on the server and that's not installed
{"level":"info","ts":1719977851.3049219,"caller":"runner/service.go:232","msg":"received initial request","_id":"01J1VAPE783Z66H6GK879WAMBY","knownID":"01J1VAAQ1CKE25TDPJCPKSYFPV","knownName":"echo-host","req":{"commandMode":"COMMAND_MODE_INLINE_SHELL","commands":["ls -la",""],"directory":"/Users/jlewi/git_triton","envs":["RUNME_ID=01J1VAAQ1CKE25TDPJCPKSYFPV","RUNME_RUNNER=v1","TERM=xterm-256color"],"knownId":"01J1VAAQ1CKE25TDPJCPKSYFPV","knownName":"echo-host","languageId":"sh","programName":"/bin/zsh","storeLastOutput":true,"tty":true,"winsize":{"cols":190,"rows":10}}}

* If I run it locally it works

* Internal failure executing runner: fork/exec /bin/zsh: no such file or directory


* One issue with setting a remote runner is that the remote runner is used for serialization so if its not available you can't open save which isn't great.