Docker version: v18.03

## Docker Engine

构成：
* server。即Docker daemon，dockerd指令。
* REST API。用以指示daemon执行什么操作。
* CLI client。即docker指令。

![Docker Engine](images/engine-components-flow.png)

## Docker architecture

Docker采用client-server，client通过REST API访问daemon。

![Docker架构](images/architecture.svg)

* Docker daemon。管理Docker对象（镜像、容器、网络和volumes）。daemon之间也可互相通信来实现Docker services的管理。
* Docker client。用户与Docker交互的主要工具。
* Docker registries。存储及分发镜像。
* Docker objects。包括：
  * 镜像。镜像是一个带有指令的只读模板，用以创建Docker容器。除使用registry中现有的镜像外，也可以通过Dockerfile创建自己需要的镜像。Dockerfile中的每条指令会创建一个layer。重新创建镜像时，只有变动的layer会重新构建。
  * 容器。容器是镜像的运行实例。通过API或CLI都可以管理容器。移除容器时，在持久存储外的任何修改都会随着消失。
  * Services。Services允许在多个daemons间scale容器，这些daemons组成了swarm。可通过service定义预期的状态（如容器副本数量）。

## The underlying technology

* Namespaces。Docker使用namespaces提供容器的隔离工作区，Docker会为一个容器创建多个namespaces，如：
  * pid namespace。隔离process。
  * net namespace。管理网络接口。
  * ipc namespace。管理IPC（InterProcess Communication）资源的访问。
  * mnt namespace。管理文件系统的挂载点。
  * uts namespace。隔离内核及版本标志。UTS: Unix Timesharing System
* Control groups。即cgroups。可用来限制应用使用的资源（如内存）。
* Union file systems。是用以创建layer并使其轻量快捷的文件系统。Docker用其提供容器的构建模块。
* Container format。Docker Engine将以上三者封装成为container format，即默认的libcontainer。未来会提供更多的container formats。