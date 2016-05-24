# pv与pvc

PersistentVolume（pv）和PersistentVolumeClaim（pvc）是k8s提供的两种API资源，用于抽象存储细节。管理员关注于如何通过pv提供存储功能而无需关注用户如何使用，同样的用户只需要挂载pvc到容器中而不需要关注存储卷采用何种技术实现。

pvc和pv的关系与pod和node关系类似，前者消耗后者的资源。pvc可以向pv申请指定大小的存储资源并设置访问模式。

## 生命周期

pv和pvc遵循以下生命周期：

* 供应准备。管理员在集群中创建多个pv供用户使用。
* 绑定。用户创建pvc并指定需要的资源和访问模式。在找到可用pv之前，pvc会保持未绑定状态。
* 使用。用户可在pod中像volume一样使用pvc。
* 释放。用户删除pvc来回收存储资源，pv将变成“released”状态。由于还保留着之前的数据，这些数据需要根据不同的策略来处理，否则这些存储资源无法被其他pvc使用。
* 回收。pv可以设置三种回收策略：保留（Retain），回收（Recycle）和删除（Delete）。
    - 保留策略允许人工处理保留的数据。
    - 删除策略将删除pv和外部关联的存储资源，需要插件支持。
    - 回收策略将执行清除操作，之后可以被新的pvc使用，需要插件支持。

## pv类型

pv支持以下类型：

* GCEPersistentDisk
* AWSElasticBlockStore
* NFS
* iSCSI
* RBD (Ceph Block Device)
* Glusterfs
* HostPath (single node testing only – local storage is not supported in any way and WILL NOT WORK in a multi-node cluster)

## pv属性

pv拥有以下属性：

* 容量。目前仅支持存储大小，未来可能支持IOPS和吞吐量等。
* 访问模式。ReadWriteOnce：单个节点读写。ReadOnlyMany：多节点只读。ReadWriteMany：多节点读写。挂载时只能使用一种模式。
* 回收策略。目前NFS和HostPath支持回收。 AWS、EBS、GCE、PD和Cinder支持删除。
* 阶段。分为Available（未绑定pvc）、Bound（已绑定）、Released（pvc已删除但资源未回收）、Failed（自动回收失败）

## pvc属性

* 访问模式。与pv的语义相同。在请求资源时使用特定模式。
* 资源。申请的存储资源数量

## nfs示例

### 安装nfs服务

以ubuntu为例，执行以下指令安装nfs server；

```
sudo apt-get install nfs-kernel-server
```

安装后修改/etc/exports，在文件末尾添加以下内容：

```
/ *(rw,insecure,no_root_squash)
```

重启nfs server：

```
$ sudo service rpcbind restart 
rpcbind stop/waiting
rpcbind start/running, process 10558

$ sudo service nfs-kernel-server restart
 * Stopping NFS kernel daemon                                                                                                                                            [ OK ] 
 * Unexporting directories for NFS kernel daemon...                                                                                                                      [ OK ] 
 * Exporting directories for NFS kernel daemon...                                                                                                                               exportfs: /etc/exports [2]: Neither 'subtree_check' or 'no_subtree_check' specified for export "*:/".
  Assuming default behaviour ('no_subtree_check').
  NOTE: this default has changed since nfs-utils version 1.0.x
                                                                                                                                                                         ]
 * Starting NFS kernel daemon
```

### pv和pvc创建与使用

本例中使用的yaml可参考[这里](./nfs.zip)

1. 首先创建pv和pvc

```
$ kubectl create -f nfs-pv.yaml 
persistentvolume "nfs" created

$ kubectl create -f nfs-pvc.yaml 
persistentvolumeclaim "nfs" created
```

2. 接下来创建nfs-busybox rc。对应的pod会向存储卷中不定期的更新index.html

```
$ kubectl create -f nfs-busybox-rc.yaml 
replicationcontroller "nfs-busybox" created
```

3. 再创建nfs-web rc和service。对应的pod会将存储卷挂载到nginx的静态目录中，这样我们可以通过service来访问index.html。

```
$ kubectl create -f nfs-web-rc.yaml 
replicationcontroller "nfs-web" created

$ kubectl create -f nfs-web-service.yaml 
service "nfs-web" created
```

4. 查看结果。可以看到index.html会不定期的更新显示时间。

```
# 查看nfs-web的ip
$ kubectl get svc
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.0.0.1     <none>        443/TCP   9h
nfs-web      10.0.0.18    <none>        80/TCP    2m

# 访问nfs-web
$ curl 10.0.0.18
Tue May 24 11:40:53 UTC 2016
nfs-busybox-c2vf2

$ curl 10.0.0.18
Tue May 24 11:40:59 UTC 2016
nfs-busybox-c2vf2
```
