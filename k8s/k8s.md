# k8s学习笔记

## 本地环境搭建

[参考个人博客](http://blog.csdn.net/xts_huangxin/article/details/51118523)

## 构建服务

使用以下yaml创建两个副本的Deployment

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 2
  template:
    metadata:
      labels:
        run: my-nginx
    spec:
      containers:
      - name: my-nginx
        image: nginx
        ports:
        - containerPort: 80
```

使用以下yaml创建Service

```
apiVersion: v1
kind: Service
metadata:
  name: my-nginx
  labels:
    run: my-nginx
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    run: my-nginx
```

以上会将Label为run=my-nginx的所有Pod的80端口创建服务。注意Service的spec可以定义targetPort，该端口为访问容器的端口。port为在集群中访问Service的端口。

每个Pod通过Endpoints暴露端口，使用“kubectl get ep Service名称”来查看Endpoints。

通过“kubectl get svc Service名称”可以查看到CLUSTER-IP，在任意节点中使用<CLUSTER-IP>:<PORT>可以访问Service。

## DNS

k8s提供了一个DNS集群插件Service，可以自动将DNS名称赋给Service。

使用以下指令可以查看dns服务是否启动：

```
kubectl get services kube-dns --namespace=kube-system
```

如果未启动，可以修改启动脚本的ENABLE_CLUSTER_DNS变量，并重启集群。

开启kube-dns之后，每个Service名称会作为dns名称，在任意pod中通过Service名称即可访问对应服务，而无需知道Service的IP。

如 kubectl run curl --image=radial/busyboxplus:curl -i --tty进入后执行nslookup \<service name\>

## 端口对外

k8s提供两种将端口暴露在外网的方法：NodePort和LoadBalancer。在Service的spec中指定type为NodePort时，k8s将port映射到当前节点的一个端口，客通过节点IP和NodePort访问集群内的服务。

LoadBalancer待验证。

## 辅助容器

同一个pod中可以部署多个容器，一个主容器和一些有辅助性质的容器，通常这些容器通过文件系统交互。例如下面的yaml将创建一个Web服务容器和一个拉取git代码库更新的辅助容器。

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      volumes:
      - name: www-data
        emptyDir: {}
      containers:
      - name: nginx
        image: nginx
        # This container reads from the www-data volume
        volumeMounts:
        - mountPath: /srv/www
          name: www-data
          readOnly: true
      - name: git-monitor
        image: myrepo/git-monitor
        env:
        - name: GIT_REPO
          value: http://github.com/some/repo.git
        # This container writes to the www-data volume
        volumeMounts:
        - mountPath: /data
          name: www-data
```

## 资源管理

以下示例显示如何设置资源

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: redis
spec:
  template:
    metadata:
      labels:
        app: redis
        tier: backend
    spec:
      containers:
      - name: redis
        image: kubernetes/redis:v1
        ports:
        - containerPort: 80
        resources:
          limits:
            # cpu units are cores
            cpu: 500m
            # memory units are bytes
            memory: 64Mi
          requests:
            # cpu units are cores
            cpu: 500m
            # memory units are bytes
            memory: 64Mi
```

## 活性检测与就绪检测

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            # Path to probe; should be cheap, but representative of typical behavior
            path: /index.html
            port: 80
          initialDelaySeconds: 30
          timeoutSeconds: 1
```

## 终止通知

k8s支持两种通知：

* 向应用发送SIGTERM信号，如果未能停止则默认30秒（通过spec.terminationGracePeriodSeconds控制）之后发送SIGKILL。
* pre-stop lifecycle hook。在SIGTERM信号发送之前执行，例如：

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx
spec:
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
        lifecycle:
          preStop:
            exec:
              # SIGTERM triggers a quick exit; gracefully terminate instead
              command: ["/usr/sbin/nginx","-s","quit"]
```

## 终止消息

容器终止时默认将终止消息写入/dev/termination-log，以下指令用来查看终止消息。

```
$ kubectl get pods/pod-w-message -o go-template="{{range .status.containerStatuses}}{{.lastState.terminated.message}}{{end}}"
Sleep expired
$ kubectl get pods/pod-w-message -o go-template="{{range .status.containerStatuses}}{{.lastState.terminated.exitCode}}{{end}}"
0
```

## 组织资源配置文件

可将多个资源配置写入一个YAML文件中，使用“---”分割

kubectl create也可指定多个-f选项，或者指定一个目录。

推荐做法是将与一个微服务或应用层相关的资源配置到一个文件中，将一个应用相关的配置文件放在在一个目录下。

## kubectl的批量操作

kubectl delele -f可指定配置文件或目录，用来删除对应名称的资源。

kubectl delete resource1/name1 resource2/name2也可以批量删除资源。

kubectl delete deployment,services -l app=nginx可以通过资源label进行批量删除

## label更新

kubectl label pods -l app=nginx tier=fe可以更新app为nginx的pods的label

## 应用扩容与缩容

kubectl scale deployment/my-nginx --replicas=1可以将pods副本数量改为1

kubectl autoscale deployment/my-nginx --min=1 --max=3可以让Deployment自动在1~3个副本直接扩容或缩容

## 资源在线更新

建议将配置文件使用源代码管理工具管理起来，在配置更新后使用kubectl apply指令将更新应用到集群中。