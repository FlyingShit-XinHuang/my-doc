Deployment是新一代用于Pod管理的对象，与Replication Controller相比，它提供了更加完善的功能，使用起来更加简单方便。

注：本文进行的相关操作是基于k8s 1.2.2版本执行的。

## Deployment相关操作

### 创建

我们可以使用下面的yaml文件来创建一个Deployment：

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: nginx
        track: stable
    spec:
      containers:
      - name: nginx
        image: index.tenxcloud.com/docker_library/nginx:1.7.9
        ports:
        - containerPort: 80
```

从上面的例子中可以发现Deployment与RC的定义基本相同，需要注意的是apiVersion和kind是有差异的。

### 状态查询

使用kubectl get可以查询Deployment当前状态：

```
$ kubectl get deployment
NAME               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3         3         3            3           2h
```

其中DESIRED为期望的Pod数量，CURRENT为当前的数量，UP-TO-DATE为已更新的数量，AVAILABLE为已运行的数量。通过这四个数量我们可以了解到Deployment目前的状态。Deployment会自动处理直到四个数量达到一致，而在Deployment更新过程中CURRENT、UP-TO-DATE和AVAILABLE会根据不同情况发生变化。

### 更新

Deployment更新分为两种情况：

* rolling-update。只有当Pod template发生变更时，Deployment才会触发rolling-update。此时Deployment会自动完成更新，且会保证更新期间始终有一定数量的Pod为运行状态。
* 其他变更，如暂停/恢复更新、修改replica数量、修改变更记录数量限制等操作。这些操作不会修改Pod参数，只影响Deployment参数，因此不会触发rolling-update。

通过kubectl edit指令更新Deployment，可以将例子中的nginx镜像版本改成1.9.1来触发一次rolling-update。期间通过kubectl get来查看Deployment的状态，可以发现CURRENT、UP-TO-DATE和AVAILABLE会发生变化。

### 删除

kubectl delete指令可以用来删除Deployment。需要注意的是通过API删除Deployment时，对应的RS和Pods不会自动删除，需要依次调用删除Deployment的API、删除RS的API和删除Pods的API。

## 特性

### 使用RS管理Pod

Replica Set（简称RS）是k8s新一代的Pod controller。与RC相比仅有selector存在差异，RS支持了set-based selector（可以使用in、notin、key存在、key不存在四种方式来选择满足条件的label集合）。Deployment是基于RS实现的，我们可以使用kubectl get rs命令来查看Deployment创建的RS：

```
$ kubectl get rs
NAME                          DESIRED   CURRENT   AGE
nginx-deployment-1564180365   3         3         6s
nginx-deployment-2035384211   0         0         36s
```

由Deployment创建的RS的命名规则为“&lt;Deployment名称&gt;-&lt;pod template摘要值&gt;”。由于之前的操作中我们触发了一次rolling-update，因此会查看到两个RS。更新前后的RS都会保留下来。

### 弹性伸缩

与RC相同，只需要修改.spec.replicas就可以实现Pod的弹性伸缩。

### 重新部署

如果设置Deployment的.spec.strategy.type==Recreate时，更新时会将所有已存在的Pod杀死后再创建新Pod。与RC不同的是，修改Deployment的Pod template后更新操作将会自动执行，无需手动删除旧Pod。

### 更完善的rolling-update

与RC相比，Deployment提供了更完善的rolling-update功能：

* Deployment不需要使用kubectl rolling-update指令来触发rolling-update，只需修改pod template内容即可。这条规则同样适用于使用API来修改Deployment的场景。这就意味着使用API集成的应用，无须自己实现一套基于RC的rolling-udpate功能，Pod更新全都交给Deployment即可。
* Deployment会对更新的可用性进行检查。当使用新template创建的Pod无法运行时，Deployment会终止更新操作，并保留一定数量的旧版本Pod来提供服务。例如我们更新nginx镜像版本为1.91（一个不存在的版本），可以看到以下结果：

```
$ kubectl get rs
NAME                          DESIRED   CURRENT   AGE
nginx-deployment-1564180365   2         2         25s
nginx-deployment-2035384211   0         0         36s
nginx-deployment-3066724191   2         2         6s
 
$ kubectl get pods
NAME                                READY     STATUS             RESTARTS   AGE
nginx-deployment-1564180365-70iae   1/1       Running            0          25s
nginx-deployment-1564180365-jbqqo   1/1       Running            0          25s
nginx-deployment-3066724191-08mng   0/1       ImagePullBackOff   0          6s
nginx-deployment-3066724191-eocby   0/1       ImagePullBackOff   0          6s
```

* 此外Deployment还支持在rolling-update过程中暂停和恢复更新过程。通过设置.spec.paused值即可暂停和恢复更新过程。暂停更新后的Deployment可能会处于与以下示例类似的状态：

```
$ kubectl get deployment
NAME               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3         4         2            4           2h
 
$ kubectl get rs
NAME                          DESIRED   CURRENT   AGE
nginx-deployment-1569886762   2         2         2h
nginx-deployment-2041090608   0         0         2h
nginx-deployment-956404068    2         2         2h
```

* 支持多重更新，在更新过程中可以执行新的更新操作。Deployment会保证更新结果为最后一次更新操作的执行结果。
* 影响更新的一些参数：
    * .spec.minReadySeconds参数用来设置确认运行状态的最短等待时间。更新Pod之后，Deployment至少会等待配置的时间再确认Pod是否处于运行状态。也就是说在更新一批Pod之后，Deployment会至少等待指定时间再更新下一批Pod。
    * .spec.strategy.rollingUpdate.maxUnavailable用来控制不可用Pod数量的最大值，从而在删除旧Pod时保证一定数量的可用Pod。如果配置为1，且replicas为3。则更新过程中会保证至少有2个可用Pod。默认为1。
    * .spec.strategy.rollingUpdate.maxSurge用来控制超过期望数量的Pod数量最大值，从而在创建新Pod时限制总量。如配置为1，且replicas为3。则更新过着中会保证Pod总数量最多有4个。默认为1。
    * 后两个参数不能同时为0。

### 更新回退

除了提供完善的更新功能外，Deployment还支持回退到历史版本（曾经更新过的版本）。Deployment的更新回退是基于RS和revision号来实现的：

* 在之前的示例中我们了解到每次更新都会有对应的RS，这些RS用来记录Pod template。使用相同Pod template的更新操作只会创建一个RS。
* 每个RS会对应一个revision版本号，revision是一个递增的正整数。
* 在回退Deployment时指定对应的revision即可完成回退操作，指定0可以回退到上一版本。
* 通过kubectl rollout history deployment/<Deployment名称>指令来查询可用的revision。
* 通过CLI或API查询RS详情时可以从.metadata.annotations的deployment.kubernetes.io/revision来查看RS对应的revision。
* 由于前两条规则，我们在查看所有revision时可能不会看到递增的值。因为如果使用相同的Pod template更新Deployment时，对应的RS只会保存最新的revision，旧的revsion值将被丢弃。

通过修改Deployment的.spec.rollbackTo.revision值来触发更新回退。回退后，RS对应的revision也将会更新（也就是说更新回退是使用历史记录的Pod template执行的更新操作）。

默认情况下，Deployment会保存所有的更新历史。可以通过.spec.revisionHistoryLimit来限制更新历史记录的最大数量，也就是RS的数量。

### 参考资料

Deployment用户手册：

http://kubernetes.io/docs/user-guide/deployments/

Replica Set用户手册：

http://kubernetes.io/docs/user-guide/replicasets/

set-based selector：

http://kubernetes.io/docs/user-guide/labels/#label-selectors

Deployment API：

http://kubernetes.io/docs/api-reference/extensions/v1beta1/operations/

canary部署示例：

http://kubernetes.io/docs/user-guide/managing-deployments/#canary-deployments
