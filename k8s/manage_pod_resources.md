k8s支持管理pod的cpu和memory两种计算资源，每种资源可以通过spec.container[].resources.requests和spec.container[].resources.limits两个参数管理

## 默认值

* requests未设置时，默认与limits相同。
* limits未设置时，默认值与集群配置相关。

## requests

* requests用于schedule阶段，k8s scheduler在调度pod时会保证所有pod的requests总和小于node能提供的计算能力。
* requests.cpu会被转换成docker run的--cpu-shares参数。该参数与cgroup cpu.shares功能相同，具体为：
    * 用来设置容器的cpu使用的相对权重。
    * 该参数只有在CPU资源不足时生效，系统会根据容器requests.cpu的比例来分配cpu资源给该容器。例如两个container设置了相同的requests.cpu，当两者抢占CPU资源时，系统会给两个container分配相同比例的CPU资源。
    * CPU资源充足时，requests.cpu不会限制单个container占用的最大值，即单个container可以独占CPU。
* requests.memory没有对应的docker run参数，只作为k8s调度依据。
* 可以使用requests来设置各容器需要的最小资源，使得k8s可以将更多的pod分配到一个node来实现超卖。

## limits

* limits用于限制运行时容器占用的资源。
* limits.cpu会被转换成docker run的--cpu-quota参数。该参数与cgroup cpu.cfs_quota_us功能相同，具体为：
    * 用来限制容器的最大CPU使用率。
    * cpu.cfs_quota_us参数与cpu.cfs_period_us通常结合使用，后者用来设置时间周期，前者设置在时间周期内可以使用的CPU时间。两者的比例即为CPU的最大使用率。
    * k8s将docker run的--cpu-period参数设置为100000，即100毫秒。该参数就对应着cgroup的cpu.cfs_period_us参数。
    * limits.cpu的单位使用m，意义为millicore，即千分之一核。250m就表示25%的最大cpu使用率，k8s会将其转换为25毫秒的-cpu-quota参数。
* limits.memory会被转换成docker run的–memory参数。用来限制容器使用的最大内存。
* 当容器申请内存超过limits时会被终止，并根据重启策略进行重启。
* 容器的CPU使用率不允许长时间超过limits，当超过limits时不会被终止。

