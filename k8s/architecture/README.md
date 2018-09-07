k8s version：1.11

## Master components

* kube-apiserver。对外暴露k8s资源访问API。
* ETCD。存储k8s资源描述。
* kube-scheduler。监控未分配的pods资源，为其分配node。
* kube-controller-manager。包含：
  * Node Controller。负责监控node运行状态。
  * Replication Controller。负责正常运行了指定数量的pod。
  * Endpoints Controller。负责EP资源，以对应service与pod关系。
  * Service Account & Token Controllers。负责为namespace创建account和API access token
* cloud-controller-manager。1.6版本新增的alpha功能。k8s核心代码之前耦合了适配云平台差异的逻辑。以后，这部分代码需要云平台自己维护，以连接cloud-controller-manager。包括：
  * Node Controller。检测node是否被删除。
  * Route Controller。设置底层基础设施的路由。
  * Service Controller。在云平台中管理lb。
  * Volume Controller。管理volume，并与云平台交互实现volume编排。

没有cloud-controller-manager时的架构：

![无CCM的架构](images/pre-ccm-arch.png)

加入cloud-controller-manager之后的架构：

![有CCM的架构](images/post-ccm-arch.png))

  ## Node components

  * kubelet。每个node上运行的agent，确保pod中的containers正常运行。
  * kube-proxy。在主机中维护网络规则，转发请求。以实现service的抽象。
  * container运行时。Docker、rkt、runc等

  ## Addons

  * DNS
  * Web UI
  * Container resource monitoring。在中央存储中记录container运行的metrics时序数据，可通过UI浏览。
  * Cluster-level Logging。搜集日志，并保存到中央日志存储中，且可通过UI查询日志。