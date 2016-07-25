近期研究了一个kube-dns多次重启的问题，顺便了解了一下kube-dns的原理，在此与大家分享。

注：本文内容均基于k8s 1.2.2版本

## 原理简介

kube-dns用来为kubernetes service分配子域名，在集群中可以通过名称访问service。通常kube-dns会为service赋予一个名为“service名称.namespace.svc.cluster.local”的A记录，用来解析service的cluster ip。

在实际应用中，如果访问default namespace下的服务，则可以通过“service名称”直接访问。如果访问其他namespace下的服务，则可以通过“service名称.namespace”访问。

我们可以通过以下yaml文件来创建一个pod：

```
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - image: busybox
    command:
      - sleep
      - "3600"
    imagePullPolicy: IfNotPresent
    name: busybox
  restartPolicy: Always
```

查询kubernetes service：

```
$ kubectl exec busybox -- nslookup kubernetes.default

Server:    10.0.0.10
Address 1: 10.0.0.10

Name:      kubernetes.default
Address 1: 10.0.0.1
```


## 配置

kube-dns服务可以手动创建，参考https://github.com/kubernetes/kubernetes/tree/v1.2.2/cluster/addons/dns 下的skydns-rc.yaml.in和skydns-svc.yaml.in。在skydns-svc中设置kube-dns的cluster ip，在skydns-rc中设置本地域名。通常设置为10.0.0.10和cluster.local。之后通过kubectl create便可以创建kube-dns服务了。

kubelet相应的参数为：

```
--cluster-dns=<DNS service ip>
--cluster-domain=<default local domain>
```

k8s会为每个容器提供默认的/etc/resolv.conf配置，内容为：

```
search default.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.0.0.10
options ndots:5
```

当我们在容器中访问服务时，系统会将search记录与我们指定的名称拼接，并在nameserver中解析dns。如使用上述配置，当使用“service名称”访问服务时，最终会使用default.svc.cluster.local这条search记录拼接完整的服务名称，当使用“service名称.namespace”时，最终会使用svc.cluster.local这条search记录。cluster.local这条记录是为了兼容旧版本的服务，因为svc子域名在早期版本是没有的。

如果想修改默认的resolv.conf配置，可以通过kubelet的--resolv-conf参数来指定配置文件路径。

## pod组成

kube-dns pod由四个容器组成，skydns、etcd、kube2sky和healthz：

* skydns用来提供dns查询服务，它使用etcd作为存储。dns记录需要使用etcd api或etcdctl添加到etcd中。
* etcd用来为skydns提供存储服务。
* kube2sky是k8s与skydns之间的桥梁，它会监听k8s service的修改，并生成对应的dns记录存储到etcd中。此外还会在启动成功后监听8081端口，以便k8s readiness probe调用。
* healthz用来检测skydns的可用性，它会定期执行nslookup指令访问skydns。

skydns是一个开源项目，可参考https://github.com/skynetservices/skydns 。

kube2sky的源码只有一个go文件，一般在容器的根目录中可以找到当前使用的源码。也可以在k8s源码中找到，https://github.com/kubernetes/kubernetes/blob/v1.2.2/cluster/addons/dns/kube2sky/kube2sky.go 。原理为：
* kube2sky主要使用了k8s.io/kubernetes/pkg/controller/framework的NewInformer方法来监听service，并设置了添加、删除、修改事件的回调，在回调中修改etcd保存的dns记录。
* NewInformer方法的其中一个返回值是k8s.io/kubernetes/pkg/client/cache的Store实例，可看做是一个本地缓存。当监听的资源发生变更时，k8s.io/kubernetes/pkg/controller/framework会将Store实例自动更新，这样当事件触发时可以在回调中使用Store实例来查询资源而不用实时查询k8s API，从而提高效率。
* 事件监听是通过对应资源的watch API和定期调用list API来实现的，watch API提供了流式的通知功能，调用list API可以周期性的全量查询，避免遗漏通知。

healthz源码在github.com/kubernetes/contrib项目中，可以参考https://github.com/kubernetes/contrib/blob/master/exec-healthz/exechealthz.go 。它的原理是：

* 每隔一段时间（period参数，默认2秒）执行指定的shell指令（cmd参数）
* 提供一个http服务（port参数，默认监听8080端口）来查询最近一次shell指令的执行情况，执行失败或超时（latency参数，默认30秒）返回503错误。
* 查询地址为host:port/healthz
* 高版本的healthz还提供了quiet参数，可以减少不必要的日志输出。

kube-dns pod使用healthz提供的http服务设置了liveness probe来实现活性检测。这次碰到的重启问题就是由于healthz执行nslookup超时导致活性检查失败而触发了重启（最近测试服务器貌似资源不足了，经常卡出翔）。


## 调试

1. 我们可以在集群中使用busybox创建的pod来检查service的dns是否生效，可参考第一节中的示例。

2. 也可以访问kube-dns pod中的etcd容器来查看dns记录，使用以下指令进入到容器中：

```
$ kubectl exec -ti -c etcd <kube-dns-pod-name> --namespace=kube-system -- sh
```

在容器中运行etcdctl get或etcdctl ls来查看数据，dns所有记录都存在/skydns目录下。skydns会将域名顺序倒置并使用“/”替换“.”来生成etcd key的路径，例如kube-dns.kube-system.svc.cluster.local对应的路径为local/cluster/svc/kube-system/kube-dns/。在该路径下会有一个自动生成的随机名称的文件，查看该文件就可以看到dns记录了。

```
$ etcdctl ls /skydns/local/cluster/svc/kube-system/kube-dns

/skydns/local/cluster/svc/kube-system/kube-dns/21172049

$ etcdctl get /skydns/local/cluster/svc/kube-system/kube-dns/21172049

{"host":"10.0.0.10","priority":10,"weight":10,"ttl":30,"targetstrip":0}
```

3. 如需要调试kube2sky，可以修改源码之后新建一个带有kube2sky和etcd容器的pod，kube2sky会自动查询k8s集群信息并写入到etcd中，这样不会影响kube-system命名空间下的kube-dns服务，还可以查看etcd记录是否正确以及打印调试日志。