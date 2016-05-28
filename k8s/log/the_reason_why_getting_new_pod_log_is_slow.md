# 首次通过k8s elasticsearch获取新建Pod的日志缓慢的原因

使用公有云查看容器服务的时候偶尔会遇到这样的情况，在创建完容器并运行后去查看日志的时候总是加载不出来，需要等待十几秒甚至一分钟才能加载。我“有幸”被分配来解决这个问题，经过一天的努力终于发现这个问题的原因，特与大家分享。

## k8s日志记录工作原理

关于k8s日志的介绍可以参考[这篇文章](http://kubernetes.io/docs/getting-started-guides/logging/)，这里简单总结几点：

* 通过命令行和API查询的日志的生命周期与Pod相同，也就是说Pod重启之后之前的日志就不见了。
* 为了解决上述问题，可以通过“集群级别日志”技术来获取Pod的历史日志。
* k8s支持多种类型的“集群级别日志”，我们使用的是Elasticsearch。

关于k8s Elasticsearch日志的相关介绍可以看[这篇](http://kubernetes.io/docs/getting-started-guides/logging-elasticsearch/)，这里简单总结几点：

* k8s会在每个node上启动一个Fluentd agent Pod来收集当前节点的日志。
* k8s会在整个集群中启动一个Elasticsearch Service来汇总整个集群的日志。
* 上述内容可以通过kubectl get pod/rc/service --namespace=kube-system来查看。
* 我们可以通过Elasticsearch API来按条件查询日志。
* Elasticsearch URL可以通过kubectl cluster-info指令查看

## 问题定位

Elasticsearch只是用来汇总和查询日志的，系统压力很小时Elasticsearch API查询不到日志基本是由于Fluentd agent没有将日志准时发送到Elasticsearch。所以现在要来了解Fluentd agent的工作机制。k8s github代码库中的Fluentd配置介绍了很多细节，可以参考[这里](https://github.com/kubernetes/kubernetes/blob/dae5ac482861382e18b1e7b2943b1b7f333c6a2a/cluster/addons/fluentd-elasticsearch/fluentd-es-image/td-agent.conf)。

简单说就是Fluentd会在node的/var/log/containers/目录中监控日志，这些文件名称包含了namespace、pod名、rc名等信息，文件内容是时间、日志级别和日志内容等信息。Fluentd以tail形式监控这些文件并将修改发送给Elasticsearch。

在Fluentd配置中可以看到“flush\_interval 5s”这样的信息，理论上在系统压力很小时应该几秒内就能看到日志，那为何会有文章开头提到的问题呢？原来是由于创建容器后，/var/log/containers/下会创建新的日志文件，而Fluentd对于新生成的日志不会立即进行监控，而是有一个时间间隔，具体可以参考[这里](http://docs.fluentd.org/articles/in_tail#)。其中讲述的是Fluentd tail Input Plugin的配置，对该问题产生影响的就是refresh\_interval配置。由于k8s使用的Fluentd配置文件中没有指定refresh_interval，因此使用的是60s的默认配置。这样就有可能导致在创建容器服务且运行之后最多要等待1分钟才能查到日志。

## 验证方法

验证方法有很多，我的方法比较笨拙：

* 首先修改/etc/profile，添加export HISTTIMEFORMAT="%Y-%m-%d %H:%M:%S  "变量，并重新打开shell。该操作目的是让history命令可以显示历史操作的执行时间。
* 然后准备好Elasticsearch的查询方法，如下所示，其中我用的rc名称是aaa和bbb。

```
curl http://192.168.1.81:8080/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging/logstash-2016.05.26/_search?pretty -d '{"sort":[{"time_nano":{"order":"desc"}}],"query":{"bool":{"must":[{"match":{"kubernetes.namespace_name":"wangleiatest"}}, {"match": {"kubernetes.container_name":"rc名称"}}]}}}'
```

* 创建aaa rc，并使用上述查询方法查看日志。我是间隔5秒进行一次查询。
* 查询到aaa日志之后，马上创建bbb rc，并用上述方法查询bbb的日志。依然用5秒间隔查询，直到看到日志。
* 这时用history 20来查看历史记录就可以看到最后一条查询aaa日志的操作与最后一条查询bbb日志操作的时间间隔约为1min。

```
2010  2016-05-26 18:45:45  curl http://192.168.1.81:8080/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging/logstash-2016.05.26/_search?pretty -d '{"sort":[{"time_nano":{"order":"desc"}}],"query":{"bool":{"must":[{"match":{"kubernetes.namespace_name":"wangleiatest"}}, {"match": {"kubernetes.container_name":"aaa"}}]}}}'
2011  2016-05-26 18:46:45  curl http://192.168.1.81:8080/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging/logstash-2016.05.26/_search?pretty -d '{"sort":[{"time_nano":{"order":"desc"}}],"query":{"bool":{"must":[{"match":{"kubernetes.namespace_name":"wangleiatest"}}, {"match": {"kubernetes.container_name":"bbb"}}]}}}'
```

## 解决方法

问题定位到了，解决就很简单了，只需要修改Fluentd配置，添加指定refresh_interval并重做镜像，再使用新的镜像来创建Fluentd agent Pod就好了。
