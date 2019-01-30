# captains-log

Captain's Log is a Kubernetes Operator for deploying [hugo](https://gohugo.io)-based static sites to Kubernetes and managing site content via the Kubernetes API.

## Getting Started

The recommended way to configure Captains Log is to use [Replicated Ship](https://github.com/replicatedhq/ship):

```shell
brew install ship
ship init https://github.com/dexhorthy/captains-log/blob/master/install.yaml
```

Ship will download and give you an opportunity to review the Kubernetes manifests included to run Captain's Log. You can create patches and overlays to make any changes necessary for your environment. Once finished, follow the instructions in Ship and `kubectl apply -f rendered.yaml`.

You can then use `ship watch && ship update` to watch and configure updates as they are shipped here.

## Creating A Blog


```sh
echo 'apiVersion: blogging.dexhorthy.com/v1alpha1
kind: Blog
metadata:
  name: my-blog
spec:
  title: My Blog
  serviceType: LoadBalancer # optional
  ' | kubectl apply -f -
```

```sh
echo 'apiVersion: blogging.dexhorthy.com/v1alpha1
kind: BlogPost
metadata:
  name: first-post
spec:
  blog: my-blog # matches Blog name above
  content: |
      ---
      title: My First Post
      date: 2018-01-27T14:53:18-08:00
      draft: false
      ---
      Captain's Log: I've created my first post!
  ' | kubectl apply -f -
```

If you're using Docker For Mac, the `LoadBalancer` service will let you view your Blog on [localhost:1313](http://localhost:1313). For other Kubernetes providers, use `kubectl get svc` to get the service IP, or use `CluterIP` + `Ingress` to connect.

## Contributing

Fork and clone this repo, and you can run it locally on a Kubernetes cluster with [tilt](https://github.com/windmilleng/tilt):

```shell
make install  # this will install the CRDs to your cluster
tilt up  # this will start the manager and controllers in your cluster, and watch for file changes and redeploy
```

