# captains-log

Captain's Log is a Kubernetes Operator for deploying [hugo](https://gohugo.io)-based static sites to Kubernetes and managing site content via the Kubernetes API.

If you prefer a sandwich-themed demo, check out the [talk from Kubernetes LA](https://youtu.be/fWUd31TIEfY?t=1656)!

## Usage


```shell
echo 'apiVersion: blogging.dexhorthy.com/v1alpha1
kind: Blog
metadata:
  name: my-blog
spec:
  title: My Blog
  serviceType: LoadBalancer
  ' | kubectl apply -f -
```

```shell
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

Edit your post with

```shell
kubectl edit blogpost first-post
```

## Getting Started

```
kubectl apply -f https://raw.githubusercontent.com/dexhorthy/captains-log/master/operator.yaml
```

## Contributing

Fork and clone this repo, and you can run it locally on a Kubernetes cluster with [tilt](https://github.com/windmilleng/tilt):

```shell
make install  # this will install the CRDs to your cluster
tilt up  # this will start the manager and controllers in your cluster, and watch for file changes and redeploy
```

