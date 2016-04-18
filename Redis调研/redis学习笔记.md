# redis学习笔记

<!-- toc -->

## 安装

<pre>
    tar xzf redis-X.tar.gz #解压
    cd redis-X #进入代码目录
    make PREFIX=xxxx install #安装到xxxx目录
</pre>

执行redis-server启动服务，执行redis-cli访问服务。

<pre>
[root@localhost redis]# redis-server --daemonize yes
[root@localhost redis]# redis-cli
127.0.0.1:6379> set demo "hello world"
OK
127.0.0.1:6379> get demo
"hello world"
</pre>

## 环境设置

* overcommit memory系统参数设置为1：在/etc/sysctl.conf中修改vm.overcommit\_memory = 1，然后重启。或者执行命令sysctl vm.overcommit_memory=1
* 禁用内核transparent huge pages特性：echo never > /sys/kernel/mm/transparent\_hugepage/enabled
* 设置Redis的maxmemory配置。这样当内存占用过多时，Redis会报错，但不会出现故障。
* Redis被用于频繁写的应用时，保存RDB文件或重写AOF日志可能会消耗2倍内存。
* 使用主从复制时，保证主节点使用持久化，或者保证主节点不会自动重启。从节点会复制主节点数据，如果主节点重启且没有任何数据存储时，从节点数据也会被清空

## 持久化

[Redis持久化](Redis持久化.md)

## Redis集群

[Redis集群](Redis集群.md)

## 配置

配置文件通常命名为redis.conf，配置格式为keyword argument1 argument2 ... argumentN。如：

<pre>
    slaveof 127.0.0.1 6380
</pre>

启动时可指定配置文件，如redis-server redis.conf

可通过命令行指定配置，如redis-server --daemonize yes。指定配置会覆盖配置文件的配置。

可通过CONFIG SET和CONFIG GET指令在线修改配置。可通过“CONFIG GET *”指令查看可修改的配置列表