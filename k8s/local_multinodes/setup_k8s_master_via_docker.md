# 通过docker在本地创建k8s master

参考http://kubernetes.io/docs/getting-started-guides/docker-multinode/master/

## 导出环境变量

export MASTER_IP=10.0.2.15 #本地ip
export K8S_VERSION=1.2.2
export ETCD_VERSION=2.3.1
export FLANNEL_VERSION=0.5.5
export FLANNEL_IFACE=eth0
export FLANNEL_IPMASQ=true

## 安装flanneld和etcd

### 创建隔离的docker daemon

执行以下命令：
```
sudo sh -c 'docker daemon -H unix:///var/run/docker-bootstrap.sock -p /var/run/docker-bootstrap.pid --iptables=false --ip-masq=false --bridge=none --graph=/var/lib/docker-bootstrap 2> /var/log/docker-bootstrap.log 1> /dev/null &'
```

创建后须使用-H unix:///var/run/docker-bootstrap.sock选项执行docker指令来查看该daemon下容器运行情况，如：
```
sudo docker -H unix:///var/run/docker-bootstrap.sock ps
```

### 运行etcd容器

执行下列两个指令：

```
sudo docker -H unix:///var/run/docker-bootstrap.sock run -d \
    --net=host \
    index.tenxcloud.com/coreos/etcd-amd64:${ETCD_VERSION} \
    /usr/local/bin/etcd \
        --listen-client-urls=http://127.0.0.1:4001,http://${MASTER_IP}:4001 \
        --advertise-client-urls=http://${MASTER_IP}:4001 \
        --data-dir=/var/etcd/data
```

```
sudo docker -H unix:///var/run/docker-bootstrap.sock run \
    --net=host \
    index.tenxcloud.com/coreos/etcd-amd64:${ETCD_VERSION} \
    etcdctl set /coreos.com/network/config '{ "Network": "10.1.0.0/16" }'
```

### 运行flannel容器

先停止docker服务，如：
```
sudo service docker stop
```

启动flanneld容器
```
sudo docker -H unix:///var/run/docker-bootstrap.sock run -d \
    --net=host \
    --privileged \
    -v /dev/net:/dev/net \
    index.tenxcloud.com/coreos/flannel:${FLANNEL_VERSION} \
        /bin/sh -c "/opt/bin/flanneld --ip-masq=${FLANNEL_IPMASQ} \
        --iface=${FLANNEL_IFACE}"
```

执行以下操作查看环境变量：
```
sudo docker -H unix:///var/run/docker-bootstrap.sock exec $(sudo docker -H unix:///var/run/docker-bootstrap.sock ps -lq) cat /run/flannel/subnet.env
```

修改/etc/default/docker配置文件，将以上两个环境变量值替换对应的参数：
```
DOCKER_OPTS="$DOCKER_OPTS --bip=${FLANNEL_SUBNET} --mtu=${FLANNEL_MTU}"
```

删除docker bridge
```
apt-get install bridge-utils #首次执行一次即可
sudo ifconfig docker0 down
sudo brctl delbr docker0
```

重启docker
```
sudo service docker start
```

## 启动k8s master

sudo docker run \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:ro \
    --volume=/var/lib/docker/:/var/lib/docker:rw \
    --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --privileged=true \
    --pid=host \
    -d \
    index.tenxcloud.com/google_containers/hyperkube-amd64:v${K8S_VERSION} \
    /hyperkube kubelet \
        --allow-privileged=true \
        --api-servers=http://localhost:8080 \
        --v=2 \
        --address=0.0.0.0 \
        --enable-server \
        --hostname-override=127.0.0.1 \
        --config=/etc/kubernetes/manifests-multi \
        --containerized \
        --cluster-dns=10.0.0.10 \
        --cluster-domain=cluster.local