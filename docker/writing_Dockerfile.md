# Dockerfile编写技巧

这个sprint研究了如何将私有云dashboard和admin集成进CI和CD。最主要的工作就是为两个项目创建Dockerfile，使其能在我们的CI环境中构建出运行服务的镜像。经过几天的实践，发现了几个编写Dockerfile的小技巧，在此与大家分享。

## 参考资料

不久之前，本司的赵帅龙攻城狮做了一次分享，为大家讲述了一些构建镜像的优化方法，可参考这里：http://blog.tenxcloud.com/?p=1313
本司blog中也有一篇介绍构建微型镜像的文章，也颇为有借鉴意义，可参考这里：http://blog.tenxcloud.com/?p=1302
我要给大家分享的内容跟Docker cache相关，可参考官方的介绍：https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/#build-cache

## 分享内容

开始分享之前大家先看两个Dockerfile：

Dockerfile1
```
FROM index.tenxcloud.com/docker_library/alpine:edge
 
## install nodejs
RUN echo '@edge http://nl.alpinelinux.org/alpine/edge/main' >> /etc/apk/repositories
RUN apk update && apk upgrade
 
WORKDIR /home/tenxcloud
ADD package.json .
ENV NODE_PATH /home/tenxcloud/node_modules/
 
RUN apk add nodejs-lts@edge
RUN apk add git
 
### WARN: installing cnpm make size bigger
RUN npm install cnpm -g --registry=https://registry.npm.taobao.org
RUN cnpm install --registry=https://registry.npm.taobao.org && rm -f package.json
 
## clear unused packages and files
RUN apk del git
RUN npm uninstall -g cnpm \
  && npm uninstall -g npm \
  && rm -rf /tmp/* \
  && rm -rf /root/.npm/
```

Dockerfile2
```
FROM index.tenxcloud.com/docker_library/alpine:edge
MAINTAINER XinHuang<huangxin@tenxcloud.com>
 
RUN echo '@edge http://nl.alpinelinux.org/alpine/edge/main' >> /etc/apk/repositories
RUN apk update && apk upgrade
 
WORKDIR /home/tenxcloud
ADD package.json .
ENV NODE_PATH /home/tenxcloud/node_modules/
 
## install nodejs
RUN apk add nodejs-lts@edge \
  ## install git
  && apk add git \
  ## install cnpm
  && npm install cnpm -g --registry=https://registry.npm.taobao.org \
  ## install packages
  && cnpm install --registry=https://registry.npm.taobao.org \
  ## clear unused packages and files
  && rm -f package.json \
  && apk del git \
  && npm uninstall -g cnpm \
  && npm uninstall -g npm \
  && rm -rf /tmp/* \
  && rm -rf /root/.npm/
```

这两个Dockerfile的功能是一样的，都是先按照node.js和git，然后使用package.json安装依赖包，最后做一些清理工作。

从开发角度来讲Dockerfile1更优。首先是因为它可读性更好，配合注释能够很方便的了解有哪些操作。其次因为cache机制的作用，我们在调试Dockerfile时更为方便：因为Dockerfile中每条指令都会创建一个layer作用在父镜像上生成新的镜像，并形成cache。当我们修改Dockerfile中的错误后，由于之前的指令已经形成了cache，Docker在后续构建时会直接使用cache，加速执行构建步骤。这样可以避免每次构建时重复的下载操作，提高调试效率。

但从实际使用的角度来说，我更倾向使用Dockerfile2，因为它比前者生成的镜像要小。使用两者分别构建镜像可以发现Dockerfile2构建的镜像仅有70+M，而Dockerfile1有140+M。原因就是Dockerfile的每条指令都是在父镜像上创建一个layer，而不会修改父镜像本身。也就是说在Dockerfile1执行“RUN cnpm install ...”之后的镜像大小已经达到140+M，后续的清理操作虽然会减小容器运行环境的大小，但并不会减小镜像本身的大小。

## 结论

结合自身的实践，个人觉得在创建和调试Dockerfile的时候最好使用Dockerfile1这种形式，尽量将耗时的操作（如下载、安装等）拆分到多个RUN指令中，这样在后续docker build时可以利用cache机制加速构建过程。在确定Dockerfile最终的操作序列之后，使用Dockerfile2这种形式，将创建和清除操作合到一个RUN指令中，这样构建的镜像会小很多，使得镜像的push/pull过程更快。
