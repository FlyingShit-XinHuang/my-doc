# docker

## 安装

docker安装参考https://docs.docker.com/linux/ ，其中还会有简单的操作示例。支持Linux、Mac OS X、Windows

## 避免使用sudo的方法

安装后，执行docker指令需要root权限，为避免执行docker指令时添加sudo，可以将当前用户添加到docker用户组中：
sudo usermod -aG docker <user>

执行后重启终端

## 在docker中运行docker的方法

在docker容器中安装docker之后是无法启动docker daemon的，是因为docker默认会回收一些权限，以避免容器的运行对kernel造成潜在威胁。

在执行docker run指令时添加--privileged选项可以赋予容器所有权限，此时在容器中即可运行docker了。