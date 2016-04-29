# 基于k8s构建可访问cassandra的spark集群

## 前提条件

* 已搭建好kubernetes集群，且开启kube-dns

## 集群搭建

1. 启动spark-master
  ```
  kubectl create -f yaml/spark-master-service.yaml
  kubectl create -f yaml/spark-master-controller.yaml
  ```

2. 启动spark-worker
  ```
  kubectl create -f yaml/spark-worker-controller.yaml
  ```

3. 启动cassandra
  ```
  kubectl create -f yaml/cassandra-service.yaml
  kubectl create -f yaml/cassandra.yaml
  ```

4. 启动spark-driver
  ```
  kubectl create -f yaml/spark-driver.yaml
  ```

## 使用示例

1. 创建keyspace和table
  ```
  #进入cql命令行
  kubectl exec -ti cassandra -- /usr/bin/cqlsh cassandra

  #在cql中创建keyspace和table
  cqlsh> CREATE KEYSPACE test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1 };
  cqlsh> CREATE TABLE test.kv(key text PRIMARY KEY, value int);
  cqlsh> INSERT INTO test.kv(key, value) VALUES ('key1', 1);
  cqlsh> INSERT INTO test.kv(key, value) VALUES ('key2', 2);
  cqlsh> select * from test.kv;
   key  | value
  ------+-------
   key1 |     1
   key2 |     2

  ```

2. 使用spark-driver访问cassandra
  ```
  #进入spark-shell
  kubectl exec -ti <spark-pod-name> spark-shell

  #访问cassandra，获取test.kv表的大小
  scala> sc.stop
  scala> import com.datastax.spark.connector._
  scala> import org.apache.spark._
  scala> val conf = new SparkConf()
  scala> conf.set("spark.cassandra.connection.host", "cassandra")
  scala> val sc = new SparkContext("local[2]", "Cassandra Connector Test", conf)
  scala> val table = sc.cassandraTable("test", "kv")
  scala> table.count
  res2: Long = 2
  ```

## 镜像构建说明

* spark镜像构建参考image-build/README.md
* cassandra镜像可参考[kubernetes示例](https://github.com/kubernetes/kubernetes/tree/release-1.2/examples/cassandra/image)。