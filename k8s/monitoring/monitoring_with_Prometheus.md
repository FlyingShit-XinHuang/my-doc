Prometheus是一款开源的监控工具，支持k8s metrics的数据格式，同时也支持通过k8s api进行服务发现从而实现对自定义的metrics进行监控。下面通过一个示例来介绍如何将Prometheus集成到k8s集群中。

# 参考资料

文中的示例具体参考了这篇文章：https://coreos.com/blog/monitoring-kubernetes-with-prometheus.html

Prometheus官方文档：https://prometheus.io/docs

镜像：quay.io/prometheus/prometheus:v1.0.1

# 示例

编排需要两个文件：[prom-deployment.yml](prom-deployment.yml)和[prom-service.yml](prom-service.yml)。前者用于创建deployment并运行一个pod，还创建了一个configmap用于配置Prometheus。后者用于对外提供Prometheus服务。

下面介绍一下configmap中的几个关键配置：

* scrape_configs用来配置数据源（Prometheus称之为target），每个target可以用job_name来命名。Prometheus会定期向每个target发送http请求来获取metrics数据，默认path为/metrics。
* 示例中以kubernetes开头的job都是k8s相关的target，每个target配置的kubernetes_sd_configs就是告诉Prometheus如何通过k8s api发现target服务。
* 大多数情况下我们只需要配置示例中的kubernetes-nodes，其中kubernetes_sd_configs.role=node是必须的。配置之后Prometheus会自动发现各k8s node，并通过kubelet的api来获取metrics，如其中一个node target为192.168.1.97:10250/metrics。
* kubelet提供的metrics通常已满足大多数需求，它能提供各容器的CPU、内存、流量、FS IO等监控数据。
* 我们也可以提供自定义的metrics数据，包括service维度和pod维度。这需要我们自己实现/metrics接口供Prometheus获取数据。配置方式可参考示例中的kubernetes-services和kubernetes-pods
* 每个target的relabel_configs用来处理数据对应的label，__meta开头的label是Prometheus为我们封装好的label，可以用它来生成我们需要的label。label可以理解为索引，在Prometheus查询中扮演着重要角色。
* 大多数情况下，只需要设置kubernetes_sd_configs.in_cluster=true配置就可以完成k8s api的认证方式配置（该方式自动使用serviceaccount）。示例运行的环境有一些特殊设置，为此使用了tls_config和bearer_token进行了配置。

通过两个yaml文件就可以完成Prometheus的部署，在浏览器中访问service可以进入Prometheus的控制台。通常在Prometheus pod启动几分钟之后才能看到监控数据。

相比于heapster api，Prometheus提供的查询API功能更加强大，可以基于label来实现复杂查询。文档也相对完善。但Prometheus从kubelet api中查询到的metrics种类很多，每种metric含义需要进一步查找文档甚至看源码来确认。