# Docker in Docker

前段时间在研究Jenkins Docker插件时了解到了Docker in Docker（下文简称dind）相关知识，特与大家分享。

## 简介

dind顾名思义就是在一个Docker容器中运行Docker。其最初是为了简化Docker自身开发的步骤，因为修改in程序之后需要重启Docker Daemon。而有了dind之后，开发人员就可以避免重启主机的Docker Daemon，而是使用最新代码直接在容器中运行新的Docker Daemon。

后来人们开始尝试在CI中使用dind，使用Docker创建一个具有构建环境的容器，并在容器中构建镜像或运行应用容器。这样可以保证构建环境的统一和干净，而且创建构建节点非常方便和快速。

目前发现有两种实现思路，各有利弊。可参考这篇文章https://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/

## 方法一

这种方式的资料可以参考Docker Hub官方镜像的说明https://hub.docker.com/_/docker/。这种方法的大概思路是：

* 运行在主机的容器需要设置--privileged参数，同时需要为容器中的/var/lib/docker目录设置存储卷。
* 使用的镜像包含了完整的Docker运行环境，包括client和daemon。
* 官方镜像是将client和daemon分成了两个镜像，另一种类似的镜像（如jpetazzo/dind）是将client和daemon都运行在一个容器中。

这种方法的坏处在文章中也有所提及：

* 一是使用--privileged参数会使得容器内的Docker在设置一些安全配置时对主机上的Docker产生影响。
* 第二是cache的使用问题，如果容器内的Docker想要共享cache或共享主机cache，那么这种方法是不能解决的。

## 方法二

这种方式不是真正意义上的dind，其思路是：

* 将主机的docker.sock挂载到容器中
* 容器中只需使用client就可以访问daemon，不过访问的是主机的daemon

这种方式避免使用--privileged参数，并且所有容器内的Docker可以共享主机的cache来提高构建效率。但这种方法也有缺点，就是隔离性的问题，例如不同容器内的Docker构建同名镜像会存在覆盖的问题。

两种方法各有优劣，大家可根据实际需求进行选择。