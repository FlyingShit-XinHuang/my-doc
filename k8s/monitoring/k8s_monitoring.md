# k8x monitoring

## 概述

![整体架构图](monitoring-architecture.png)

* Heapster是一个整个集群监控和事件数据的整合器。
* Heapster在集群中作为一个pod运行。
* Heapster pod通过每个node的Kubelet发现所有node，并查询使用信息。
* Kubelet从cAdvisor获取数据。
* Heapster通过pod和相关的label将信息组织在一起
* 数据将推送到一个可配置的后台，用以存储和可视化。

## cAdvisor

* cAdvisor是一款开源的，用于分析容器资源使用率和性能的代理软件。
* Kubelet程序中集成了cAdvisor
* cAdvisor自动发现机器中的所有容器，并收集CPU、内存、文件系统和网络的统计数据。
* cAdvisor使用4194端口为容器提供一个简单的UI访问页面。

## Kubelet

* Kubelet桥接了k8s master和其他节点。
* Kubelet管理一个机器中运行的pods和容器
* Kubelet会找到每个pod对应的所有容器，并从cAdvisor获取单独容器使用率的统计数据。
* Kubelet通过REST API对外提供集成好的pod资源使用率统计数据。

## 存储后台

### InfluxDB and Grafana

* Grafana与InfluxDB的监控组合在开源界非常流行。InfluxDB为读写时间序列数据提供易用的API。
* Heapster默认使用两者作为存储后台。
* InfluxDB和Grafana运行在Pod中，并作为Service提供服务，以便被Heapster发现。
* Grafana容器提供UI服务，一个便于配置的仪表盘接口。
* 默认的k8s仪表盘包含一个监控集群和pods的资源使用率的示例，仪表盘的自定义以及扩展非常容易。
