---
description: Deploying Foyle on Kubernetes
title: Deploying Foyle on Kubernetes
weight: 3
---

{{< alert type="info" >}}
Deploying Foyle on Kubernetes is the recommended way for teams to have a shared instance that transfers knowledge across the organization.
{{< /alert >}}


## Installation

### Prerequisites: VSCode & RunMe 

Foyle relies on [VSCode](https://code.visualstudio.com/) and [RunMe.dev](https://runme.dev/)
to provide the frontend.

1. If you don't have VSCode visit the [downloads page](https://code.visualstudio.com/) and install it
1. Follow [RunMe.dev](https://docs.runme.dev/installation/installrunme#installing-runme-on-vs-code) instructions to install the RunMe.dev extension in vscode


### Deploy Foyle on Kubernetes

Create the namespace foyle

```
kubectl create namespace foyle
```

Clone the foyle repository to download a copy of the kustomize package

```
git clone https://github.com/jlewi/foyle.git
```

Deploy foyle

```
cd foyle
kustomize build manifests | kubectl apply -f -
```

Create a secret with your OpenAI API key

```
kubectl create secret generic openai-key --from-file=openai.api.key=/PATH/TO/YOUR/OPENAI/APIKEY -n foyle
```

### Verify its working

Check the pod is running

```
kubectl get pods -n foyle

NAME      READY   STATUS    RESTARTS   AGE
foyle-0   1/1     Running   0          58m
```

Port forward to the pod 

```
kubectl -n foyle port-forward foyle-0 8877:8877
```

Verify the service is accessible and healthy

```
curl http://localhost:8877/healthz

{"server":"foyle","status":"healthy"}
```

### Expose the service on your VPN

At this point you should customize the K8s service to make it accessible from employee machines over your VPN. 
While you could use it via `kubectl port-forward` that isn't a very good user experience.
If your looking for a VPN its hard to beat [Tailscale](https://tailscale.com/).

## Configure Runme
Inside VSCode configure Runme to use Foyle

1. Open the VSCode setting palette
1. Search for `Runme: Ai Base URL`
1. Set the address to `http://${HOST}:${PORT}/api`
  * The value of host and port will depend on how you make it accessible over vpn
  * The default port is 8877
  


