## 机制

PersistentVolumeClaimBinder以单例模式运行在master上，它会监控所有pvc并将它们绑定到与请求资源最接近的足量的pv上。也就是说pvc占用的实际存储资源可能会大于请求的资源。k8s不保证底层存储的高可用，需要提供者负责。

NewPersistentVolumeOrderedIndex用来索引不同访问模式的pv并基于容量排序。PersistentVolumeClaimBinder根据pvc的需求来查找索引。

pv是全局的，pvc可以指定namespace。

默认情况下pv使用的回收策略为Retain，此时如果绑定的pvc删除后，pv将处于“Released”状态，需要手动处理pv或自定义回收脚本。

## binder运行机制实践

binder是周期性检测并执行绑定操作的，我们可以通过下面的示例来验证。

修改nfs-pv.yaml内容为：

```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs
spec:
  capacity:
    storage: 10Mi
  accessModes:
    - ReadWriteMany
  nfs:
    server: localhost
    path: "/"
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs2
spec:
  capacity:
    storage: 5Mi
  accessModes:
    - ReadWriteMany
  nfs:
    server: localhost
    path: "/"
```

首先创建一个pvc，再创建两个pv。pvc需要的资源为1Mi，两个pv的容量分别为10Mi和5Mi。

```
$ kubectl create -f nfs-pvc.yaml
persistentvolumeclaim "nfs" created
 
$ kubectl create -f nfs-pv.yaml
persistentvolume "nfs" created
```

查看pv和pvc可以看到此时尚未进行绑定。如果先创建pv再创建pvc则会立即绑定。

```
$ kubectl get pvc
NAME      STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
nfs       Pending                                      11s
 
$ kubectl get pv
NAME      CAPACITY   ACCESSMODES   STATUS      CLAIM     REASON    AGE
nfs       10Mi       RWX           Available                       8s
nfs2      5Mi        RWX           Available                       6s
```

几分钟之后再进行查看，此时会发现绑定已完成，且pvc绑定到了5Mi容量的pv上。

```
$ kubectl get pv
NAME      CAPACITY   ACCESSMODES   STATUS      CLAIM         REASON    AGE
nfs       10Mi       RWX           Available                           7m
nfs2      5Mi        RWX           Bound       default/nfs             7m
```

## 问题

目前发现以下问题：

* 已经绑定的pv可以被删除，删除后对应的pvc仍为Bound状态。如有pod正在使用则仍可使用。如无pod使用，则创建pod时会出现错误。
* pv可以在绑定后被编辑（如访问模式，容量），导致信息与对应的pvc不一致。
* 使用Recycle策略回收nfs pv时，k8s会创建一个pod，该pod使用gcr.io/google_containers/busybox镜像和command ["/bin/sh", "-c", 'test -e /scrub && rm -rf /scrub/..?* /scrub/.[!.]* /scrub/*  && test -z "$(ls -A /scrub)" || exit 1']来清除数据。目前没找到官方的修改该pod的方法，或许可通过kube-controller-manager的--pv-recycler-pod-template-filepath-nfs选项来设置。

## 参考资料

https://github.com/kubernetes/kubernetes/blob/v1.2.4/docs/design/persistent-storage.md

https://github.com/kubernetes/kubernetes/blob/v1.2.4/docs/admin/kube-controller-manager.md