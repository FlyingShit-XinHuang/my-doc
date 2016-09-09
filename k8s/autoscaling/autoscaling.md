由于工作需要，近期对k8s自动伸缩的源码进行了阅读，在此与各位分享。

源码位置：pkg/controller/podautoscaler/horizontal.go

NewHorizontalController方法创建hpa controller，其中调用了framework.NewInformer（源码位于pkg/controller/framework/controller.go）。Informer实现了一种通知机制，当监听的资源发生变更时，会调用相应回调进行处理：

* NewInformer第一个参数为ListerWatcher接口的实例对象，对象必须实现List和Watch方法。这两个方法本质上是分别调用k8s的list和watch API。
* Informer在Run方法中创建了一个Reflector（源码位于pkg/client/cache/reflector.go）对象，Reflector来负责执行List和Watch方法（在另一个goroutine中），并将List和Watch返回的k8s对象保存到一个fifo的Queue中。这个Queue由Informer创建，并传递给Reflector。Reflector具体工作原理为：

    * 首先调用List方法，获取所有满足条件的k8s对象并更新Queue中的数据，同时使用接口返回的resourceVersion调用Watch接口来监听当前resourceVersion之后的变更。
    * Reflector会对Watch到的Added、Modified和Deleted三种事件进行处理，对Queue做相应的增删改操作。
    * Reflector会定期终止Watch操作，并重新执行步骤1和2。通过定期调用List可以获取到最新的全量数据，避免网络问题导致Watch可能遗漏一些事件而产生的问题。

* Queue的实现比较有参考意义：队列由slice存储，顺序保存key。k8s对象由map存储，key从队列中读取。这里使用了一种增量队列，map的value也为slice，其中顺序保存了一个k8s对象的增量操作{Type, Object}，Type的值有Added、Updated、Sync和Deleted，Object为每次增量操作的k8s对象。Sync类型的而增量操作对应的是List方法保存的对象，Added、Updated和Deleted对应的是Watch方法保存的对象。
* Informer循环地对Queue执行pop操作，判断队首增量操作的类型并相应的调用增删改回调。
* NewInformer第二个参数用来指定监听的k8s资源，这里指定了HorizontalPodAutoscaler，即hpa资源。第三个参数为List操作的执行间隔。第四个参数用来注册回调函数，这里只注册了Add和Update回调。

Add和Update回调最终均调用了HorizontalController的reconcileAutoscaler方法。该方法主要负责执行弹性伸缩流程：

* 首先会查询hpa资源引用的scale资源来获取当前状态（副本数和pod selector）。scale资源是deployment、rc和rs的子资源，用于封装伸缩操作相关细节。HorizontalController通过scale完成伸缩操作。
* 然后通过scale的selector查询所有pods，并计算出cpu request的均值。再通过heapster api获取所有pods（1分钟内）的metrics并计算均值。
* request均值/metrics均值为当前的利用率，使用当前利用率与hpa设置的利用率的比值作为弹性伸缩的依据，如果比值大于1.1则扩容，如果小于0.9则缩容。
* 弹性伸缩时还会保证伸缩后的副本数量符合hpa设置的最大和最小值的限制。此外还会保证短时间内不会频繁的弹性伸缩：上次伸缩的5min内不能发生缩容，3min内不能发生扩容。
* 具体的伸缩操作有两个步骤：首先通过k8s api修改scale资源的副本数（.spec.replicas），然后再通过k8s api修改hpa的status（lastScaleTime、currentReplicas、desiredReplicas和currentCPUUtilizationPercentage）

目前custom metrics与cpu metrics不同点在于：

* custom metrics的定义和status都在hpa的annotation中
* custom metrics目前须指定具体值而非比例
* 访问heapster api时custom metrics会在url path中加入custom/
* 容器需要自行实现metrics数据收集，并需要对pod进行配置，使其对cAdvisor暴露采集端口。配置方式比较复杂，难于通过程序进行控制，且custom metrics目前尚为alpha阶段，后续可能发生较大变更，不建议使用。

弹性伸缩会由两个地方触发，一个由Watch监听到的hpa变更触发，二为定期执行List操作触发。
