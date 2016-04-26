# 使用k8s搭建spark集群

Spark的组件介绍可参考[官方文档](http://spark.apache.org/docs/latest/cluster-overview.html)
大数据生态圈简介可参考[这篇文章](http://www.36dsj.com/archives/23504)

## 基于k8s官方示例

具体参考[github k8s](https://github.com/kubernetes/kubernetes/tree/master/examples/spark)

### FAQ

#### 镜像拉取问题

该方法需要访问gcr.io下载镜像（国内一般需要vpn），需要注意的是gcr.io/google_containers/spark:1.5.2_v1镜像不能使用index.tenxcloud.com/google_containers/spark替换，替换后拉取镜像时会出现“docker: filesystem layer verification failed”错误。

可将zeppelin-controller.yaml使用的镜像修改为index.tenxcloud.com/google_containers/zeppelin:v0.5.6_v1

#### webui service使用问题

文档中的kubectl proxy --port=8001指令只能监听127.0.0.1的代理请求，不适用于测试环境和虚拟机环境，因为使用的ip地址不是127.0.0.1。
此时使用kubectl proxy --port=8001 --address=\<ip address\> --disable-filter指令启动proxy，在浏览器中访问就没问题了。

#### pyspark示例运行错误

示例中的数据源存在问题，可使用本地文件运行，例如“sc.textFile("/opt/spark/licenses/*").map(lambda s: len(s.split())).sum()”

#### Zeppelin webui使用问题

同样只能通过localhost或127.0.0.1访问，目前尚未找到解决方法。


## 基于tenxcloud镜像库搭建

需要根据k8s源码中的examples/spark/下的yaml文件搭建，将所有yaml文件复制到工作目录下。

修改spark-master-controller.yaml和spark-worker-controller.yaml：
* spec.template.spec.containers.command均修改为“/start.sh”
* spec.template.spec.containers.images分别修改为index.tenxcloud.com/google_containers/spark-master:1.5.2_v1和index.tenxcloud.com/google_containers/spark-worker:1.5.2_v1

zeppelin-controller.yaml使用的镜像修改为index.tenxcloud.com/google_containers/zeppelin:v0.5.6_v1

修改完成后，按k8s官方示例的步骤启动即可。

### 简易的spark-driver


由于zeppelin镜像非常大，拉取会消耗很多时间。可以使用下面的spark-driver.yaml创建一个简易的spark-driver：

```
kind: ReplicationController
apiVersion: v1
metadata:
  name: spark-driver 
spec:
  replicas: 1
  selector:
    component: spark-driver
  template:
    metadata:
      labels:
        component: spark-driver
    spec:
      containers:
        - name: spark-driver
          image: index.tenxcloud.com/google_containers/spark-driver:1.5.2_v1
          resources:
            requests:
              cpu: 100m

```

创建后，使用kubectl exec \<spark-driver-podname\> -it pyspark即可访问。