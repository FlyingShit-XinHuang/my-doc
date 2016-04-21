# docker

docker安装参考https://docs.docker.com/linux/，其中还会有简单的操作示例
  
安装后，执行docker指令需要root权限，为避免执行docker指令时添加sudo，可以将当前用户添加到docker用户组中：
sudo usermod -aG docker <user>

执行后重启终端