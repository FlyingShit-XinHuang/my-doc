本次迭代的功能是基于k8s job实现的。与大家分享一下job相关内容。

## 相关资料
官方介绍：
http://kubernetes.io/docs/user-guide/jobs/
API：
http://kubernetes.io/docs/api-reference/batch/v1/operations/

## 分享内容

job与rc不同，包含的pod多用于执行一次性任务、批处理工作等，执行完成后pod便会停止（status.phase变为Succeeded）。这时通过kubectl get pods是看不到pod的，需要加“-a”参数。

### RestartPolicy
job pod的template的RestartPolicy只能指定Never或OnFailure，当job未完成的情况下：
* 如果RestartPolicy指定Never，则job会在pod出现故障时创建新的pod，且故障pod不会消失。.status.failed加1。
* 如果RestartPolicy指定OnFailure，则job会在pod出现故障时其内部重启容器，而不是创建pod。.status.failed不变。

#### Never策略示例：
job.yaml如下所示，其中command故意将perl指令写错，这样pod将出现故障。
```
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    metadata:
      name: pi
    spec:
      containers:
      - name: pi
        image: index.tenxcloud.com/sdvdxl/perl
        command: ["perl1",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
```

创建job并查看job信息：
```
$ kubectl create -f job.yaml 
job "pi" created

$ kubectl describe job/pi
Name:       pi
Namespace:  default
Image(s):   index.tenxcloud.com/sdvdxl/perl
Selector:   controller-uid=cb44e05e-18f7-11e6-b739-0800278856e6
Parallelism:    1
Completions:    1
Start Time: Fri, 13 May 2016 10:45:22 +0000
Labels:     controller-uid=cb44e05e-18f7-11e6-b739-0800278856e6,job-name=pi
Pods Statuses:  1 Running / 0 Succeeded / 41 Failed
No volumes.
Events:
  FirstSeen LastSeen    Count   From            SubobjectPath   Type        Reason          Message
  --------- --------    -----   ----            -------------   --------    ------          -------
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-fvbql
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-f8qjn
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-83rn2
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-j525m
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-msjtr
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-rnqh4
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-vvp02
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-f4rs4
  1m        1m      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-gazi9
  1m        1s      33  {job-controller }           Normal      SuccessfulCreate    (events with common reason combined)
```

可以发现job在不停的创建新的pod，但只有一个pod在执行任务。

```
$ kubectl get pods -a |grep pi|wc -l
71

$ kubectl get pods
NAME       READY     STATUS              RESTARTS   AGE
pi-9d1iv   0/1       ContainerCreating   0          0s
```

执行kubectl delete时可以看到删除了所有pod，耗时较长。

#### OnFailure策略示例：
job.yaml如下所示，其中command故意将perl指令写错，这样pod将出现故障。
```
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    metadata:
      name: pi
    spec:
      containers:
      - name: pi
        image: index.tenxcloud.com/sdvdxl/perl
        command: ["perl1",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: OnFailure
```

创建job并查看job信息：
```
$ kubectl create -f job.yaml
job "pi" created

$ kubectl describe job/pi
Name:       pi
Namespace:  default
Image(s):   index.tenxcloud.com/sdvdxl/perl
Selector:   controller-uid=36422362-18fa-11e6-b739-0800278856e6
Parallelism:    1
Completions:    1
Start Time: Fri, 13 May 2016 11:02:41 +0000
Labels:     controller-uid=36422362-18fa-11e6-b739-0800278856e6,job-name=pi
Pods Statuses:  1 Running / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen LastSeen    Count   From            SubobjectPath   Type        Reason          Message
  --------- --------    -----   ----            -------------   --------    ------          -------
  11s       11s     1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-2n6pa
```

可以看到job只创建了一个pod，查看pod描述可以发现该pod一直在重启容器：

```
$ kubectl describe pod/pi-2n6pa
...
Events:
  FirstSeen LastSeen    Count   From            SubobjectPath       Type        Reason      Message
  --------- --------    -----   ----            -------------       --------    ------      -------
  2m        2m      1   {default-scheduler }                Normal      Scheduled   Successfully assigned pi-2n6pa to 127.0.0.1
  2m        2m      1   {kubelet 127.0.0.1} spec.containers{pi} Normal      Created     Created container with docker id 310aac97cc9a
  2m        2m      1   {kubelet 127.0.0.1} spec.containers{pi} Warning     Failed      Failed to start container with docker id 310aac97cc9a with error: API error (404): Container command 'perl1' not found or does not exist.

  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Normal  Created Created container with docker id 62cab877f860
  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Warning Failed  Failed to start container with docker id 62cab877f860 with error: API error (404): Container command 'perl1' not found or does not exist.

  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Normal  Created Created container with docker id b1cda2dbee87
  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Warning Failed  Failed to start container with docker id b1cda2dbee87 with error: API error (404): Container command 'perl1' not found or does not exist.

  2m    2m  1   {kubelet 127.0.0.1}     Warning FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "pi" with CrashLoopBackOff: "Back-off 20s restarting failed container=pi pod=pi-2n6pa_default(3643a8cc-18fa-11e6-b739-0800278856e6)"

  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Warning Failed  Failed to start container with docker id fc88d454f0c2 with error: API error (404): Container command 'perl1' not found or does not exist.

  2m    2m  1   {kubelet 127.0.0.1} spec.containers{pi} Normal  Created     Created container with docker id fc88d454f0c2
  1m    1m  3   {kubelet 127.0.0.1}             Warning FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "pi" with CrashLoopBackOff: "Back-off 40s restarting failed container=pi pod=pi-2n6pa_default(3643a8cc-18fa-11e6-b739-0800278856e6)"

  2m    1m  5   {kubelet 127.0.0.1} spec.containers{pi} Normal  Pulling pulling image "index.tenxcloud.com/sdvdxl/perl"
  2m    1m  5   {kubelet 127.0.0.1} spec.containers{pi} Normal  Pulled  Successfully pulled image "index.tenxcloud.com/sdvdxl/perl"
  1m    1m  1   {kubelet 127.0.0.1} spec.containers{pi} Warning Failed  Failed to start container with docker id 6a6464d1f31b with error: API error (404): Container command 'perl1' not found or does not exist.

  2m    1m  5   {kubelet 127.0.0.1}     Warning FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "pi" with RunContainerError: "runContainer: API error (404): Container command 'perl1' not found or does not exist.\n"

  1m    1m  1   {kubelet 127.0.0.1} spec.containers{pi} Normal  Created     Created container with docker id 6a6464d1f31b
  2m    3s  12  {kubelet 127.0.0.1} spec.containers{pi} Warning BackOff     Back-off restarting failed docker container
  1m    3s  8   {kubelet 127.0.0.1}             Warning FailedSync  Error syncing pod, skipping: failed to "StartContainer" for "pi" with CrashLoopBackOff: "Back-off 1m20s restarting failed container=pi pod=pi-2n6pa_default(3643a8cc-18fa-11e6-b739-0800278856e6)"
```

### 设置超时

job执行超时时间可以通过spec.activeDeadlineSeconds来设置，超过指定时间未完成的job会以DeadlineExceeded原因停止：

```
$ kubectl describe job pi
Name:               pi
Namespace:          default
Image(s):           index.tenxcloud.com/sdvdxl/perl
Selector:           controller-uid=8c172a2d-18fb-11e6-b739-0800278856e6
Parallelism:            1
Completions:            1
Start Time:         Fri, 13 May 2016 11:12:14 +0000
Active Deadline Seconds:    1s
Labels:             controller-uid=8c172a2d-18fb-11e6-b739-0800278856e6,job-name=pi
Pods Statuses:          0 Running / 0 Succeeded / 1 Failed
No volumes.
Events:
  FirstSeen LastSeen    Count   From            SubobjectPath   Type        Reason          Message
  --------- --------    -----   ----            -------------   --------    ------          -------
  7s        7s      1   {job-controller }           Normal      SuccessfulCreate    Created pod: pi-eq3yt
  4s        4s      1   {job-controller }           Normal      SuccessfulDelete    Deleted pod: pi-eq3yt
  4s        4s      1   {job-controller }           Normal      DeadlineExceeded    Job was active longer than specified deadline
```

### 删除job

通过API删除job时其对应的pod并不会自动被删除，需要手动调用删除pod的API，删除时指定query参数labelSelector=job-name=&lt;job名称>。

### 停止job

设置.spec.parallelism为0可以将job停止。此时job不会创建Pod，且会删除所有已运行的Pod（运行结束的不会被删除）。

### pod selector

job同样可以指定selector来关联pod。需要注意的是job目前可以使用两个API组来操作，batch/v1和extensions/v1beta1。当用户需要自定义selector时，使用两种API组时定义的参数有所差异。

* 使用batch/v1时，用户需要将jod的spec.manualSelector设置为true，才可以定制selector。默认为false。
* 使用extensions/v1beta1时，用户不需要额外的操作。因为extensions/v1beta1的spec.autoSelector默认为false，该项与batch/v1的spec.manualSelector含义正好相反。换句话说，使用extensions/v1beta1时，用户不想定制selector时，需要手动将spec.autoSelector设置为true。

### 多容器

如果Job中定义了多个容器，则Job的状态将根据所有容器的执行状态来变化。如果Job定义的容器中存在http server、mysql等长期的容器和一些批处理容器，则Job状态不会发生变化（因为长期运行的容器不会主动结束）。此时可以通过Pod的.status.containerStatuses获取指定容器的运行状态。