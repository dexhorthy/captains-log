我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# captains-log

Captain's Log is a Kubernetes Operator for deploying [hugo](https://gohugo.io)-based static sites to Kubernetes and managing site content via the Kubernetes API.

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

The recommended way to configure Captains Log is to use [Ship](https://github.com/replicatedhq/ship):

```shell
brew install ship
ship init https://github.com/dexhorthy/captains-log/blob/master/operator.yaml
```

Ship will download and give you an opportunity to review the Kubernetes manifests included to run Captain's Log. You can create patches and overlays to make any changes necessary for your environment. Once finished, follow the instructions in Ship and `kubectl apply -f rendered.yaml`.

You can then use `ship watch && ship update` to watch and configure updates as they are shipped here.

#### Raw Install

```
kubectl apply -f https://raw.githubusercontent.com/dexhorthy/captains-log/master/operator.yaml
```

## Contributing

Fork and clone this repo, and you can run it locally on a Kubernetes cluster with [tilt](https://github.com/windmilleng/tilt):

```shell
make install  # this will install the CRDs to your cluster
tilt up  # this will start the manager and controllers in your cluster, and watch for file changes and redeploy
```

