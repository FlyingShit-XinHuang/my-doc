#mycat

## 数据库连接

### schema.xml

修改schema.xml，配置逻辑schema和table，如：
```
	<schema name="demo" checkSQLschema="false" sqlMaxLimit="100" >
		<table name='whispir_users' primaryKey='ID' autoIncrement="true" rule="mod-long" dataNode="demo1,demo2" ></table>
	</schema>
```

配置schema和table所用的data node，一个data node对应一个db host上的一个database：
```
	<dataNode name="demo1" dataHost="mysql1" database="demo" />
	<dataNode name="demo2" dataHost="mysql2" database="demo" />
```

配置db host：

```
	<dataHost name="mysql1" maxCon="1000" minCon="2" balance="0"
	   writeType="0" dbType="mysql" dbDriver="native">
	   <heartbeat>select 1</heartbeat>
	   <writeHost host="mysql1M1" url="172.17.0.2:3306" user="root" password="123456" />
	</dataHost>

	<dataHost name="mysql2" maxCon="1000" minCon="2" balance="0"
	   writeType="0" dbType="mysql" dbDriver="native">
	   <heartbeat>select 1</heartbeat>
	   <writeHost host="mysql2M1" url="172.17.0.3:3306" user="root" password="123456" />
	</dataHost>
```

### server.xml

在server.xml中配置连接mycat的用户：

```
	<user name="demo">
		<property name="password">demo</property>
		<property name="schemas">demo</property>
	</user>
```

注：这里的user是客户端连接mycat需要的用户，与schema.xml的dataHost中设置的user不同，后者是连接mycat连接mysql时使用的用户。

## 建表

在schema.xml中配置逻辑schema和table，对应表并不会自动创建，需要连接到mycat再创建对应表。

## 自增主键

由于现在表分布在不同数据库中，所以不能依赖mysql自身的自增主键特性。mycat提供了对应功能，配置方法如下。

server.xml 配置，配置使用数据库表生成序列号：

```
<system><property name="sequnceHandlerType">1</property></system>
```

在某一个mysql数据库中创建MYCAT_SEQUENCE表，如在demo1对应的数据库中创建：

```
DROP TABLE IF EXISTS MYCAT_SEQUENCE;
-- name sequence 名称
-- current_value 当前 value
-- increment 增长步长! 可理解为 mycat 在数据库中一次读取多少个 sequence. 当这些用完后, 下次再从数据库中读取.

CREATE TABLE MYCAT_SEQUENCE (name VARCHAR(50) NOT NULL,current_value INT NOT NULL,increment INT NOT NULL DEFAULT 100, PRIMARY KEY(name)) ENGINE=InnoDB;

-- 插入sequence参数，name与表名一致
INSERT INTO MYCAT_SEQUENCE(name,current_value,increment) VALUES ('GLOBAL', 100000, 100);
```

在相同数据库中创建function

```
-- 获取当前 sequence 的值 (返回当前值,增量)
DROP FUNCTION IF EXISTS mycat_seq_currval;
DELIMITER ;;
CREATE FUNCTION mycat_seq_currval(seq_name VARCHAR(50)) RETURNS varchar(64) CHARSET 'utf8'
DETERMINISTIC
BEGIN
DECLARE retval VARCHAR(64);
SET retval="-999999999,null";
SELECT concat(CAST(current_value AS CHAR),",",CAST(increment AS CHAR)) INTO retval FROM MYCAT_SEQUENCE WHERE name = seq_name;
RETURN retval;
END
;;
DELIMITER ;

-- 设置 sequence 值
DROP FUNCTION IF EXISTS mycat_seq_setval;
DELIMITER ;;
CREATE FUNCTION mycat_seq_setval(seq_name VARCHAR(50),value INTEGER) RETURNS varchar(64) CHARSET 'utf8'
DETERMINISTIC
BEGIN
UPDATE MYCAT_SEQUENCE
SET current_value = value
WHERE name = seq_name;
RETURN mycat_seq_currval(seq_name);
END
;;
DELIMITER ;

-- 获取下一个 sequence 值
DROP FUNCTION IF EXISTS mycat_seq_nextval;
DELIMITER ;;
CREATE FUNCTION mycat_seq_nextval(seq_name VARCHAR(50)) RETURNS varchar(64) CHARSET 'utf8'
DETERMINISTIC
BEGIN
UPDATE MYCAT_SEQUENCE
SET current_value = current_value + increment WHERE name = seq_name;
RETURN mycat_seq_currval(seq_name);
END
;;
DELIMITER ;
```

接下来创建需要的表，建表时须指定自增主键，如：

```
create table whispir_users (
  id int not null auto_increment,
  name varchar(100) not null,
  primary key(id)
) engine InnoDB default charset='utf8';
```

schema.xml配置对应逻辑表，设置autoIncrement自增属性：

```
<schema name="demo" checkSQLschema="false" sqlMaxLimit="100" >
	<table name='whispir_users' primaryKey='ID' autoIncrement="true" rule="mod-long" dataNode="demo1,demo2" ></table>
</schema>
```

sequence_db_conf.properties配置表与sequence表所在data node的对应关系，注：表名须大写

```
WHISPIR_USERS=demo1
```

在sequence表中增加对应表的参数，如：

```
INSERT INTO MYCAT_SEQUENCE(name,current_value,increment) VALUES ('whispir_users', 100000, 100)
```

重启mycat。

## 分片规则

### 枚举

如按省份分片。须配置枚举值和data node映射，记录会根据枚举的值被分片到data node。可设置默认data node，未匹配到枚举值的记录被分到该node。

### 范围分片

规划某一范围数据分片到哪个data node。易于扩展。

```
<tableRule name="auto-sharding-long">
    <rule>
        <columns>user_id</columns>
        <algorithm>rang-long</algorithm>
        </rule>
</tableRule>
<function name="rang-long" class="org.opencloudb.route.function.AutoPartitionByLong">
    <property name="mapFile">autopartition-long.txt</property>
    <property name="defaultNode">0</property>
</function>
```

所有的节点配置都是从 0 开始，及 0 代表节点 1，此配置非常简单，即预先制定可能的 id 范围到某个分片
autopartition-long.txt:
0-500M=0
500M-1000M=1
1000M-1500M=2

### 取模分片

用id对指定的值求模，来决定data node。

### 取模-范围

结合了取模和范围分片，先对id取模，再根据取模后的值在范围配置中查找data node。

### 一致性hash

### 范围-取模

结合了范围和取模分片。先按范围分组，每组指定data node个数，再组内取模确定data node。

```
<tableRule name="auto-sharding-rang-mod">
    <rule>
        <columns>id</columns>
        <algorithm>rang-mod</algorithm>
    </rule>
</tableRule>
<function name="rang-mod"
 class="org.opencloudb.route.function.PartitionByRangeMod">
    <property name="mapFile">partition-range-mod.txt</property>
    <property name="defaultNode">21</property>
</function>
```

partition-range-mod.txt:
0-200M=5 //代表有 5 个分片节点
200M1-400M=1
400M1-600M=4
600M1-800M=4
800M1-1000M=6

### 冷热数据分片

最近一段时间分到一个库，超过的数据按一个间隔时间再分片。

```
<tableRule name="sharding-by-date">
    <rule>
        <columns>create_time</columns>
        <algorithm>sharding-by-hotdate</algorithm>
    </rule>
</tableRule>
<function name="sharding-by-hotdate" class="org.opencloudb.route.function.PartitionByHotDate">
    <property name="dateFormat">yyyy-MM-dd</property>
    <property name="sLastDay">10</property>
    <property name="sPartionDay">30</property>
</function>
```

### 其他

还有其他规则未介绍，如：截取子串，再对子串实施分片规则、其他日期相关分片规则。

## 扩容

扩容主要解决的问题是数据迁移，mycat暂时不支持扩容自动迁移数据。整体思路是将mysql数据导出再通过mycat导入来完成数据迁移。

使用一致性hash分片时，数据迁移量较小，且mycat提供了一些工具。

基于范围规则的分片在扩容时比较方便，只要规划的好，可以不用全量数据迁移甚至可以不用迁移现有数据。

## 实施指南

分析当前业务，具体内容包括如下几个方面：
* 数据模型：重点关注数据的增长模式（实时大量增长还是缓慢增长）和规律、数据之间的关联关系
* 数据访问模式：通过抓取系统中实际执行的 SQL，分析其频率、响应时间、对系统性能和功能的影响程度
* 数据可靠性的要求：系统中不同数据表的可靠性要求，以及操作模式
* 事务的要求：系统中哪些业务操作是严格事务的，哪些是普通事务或可以无事务的
* 数据备份和恢复问题：目前的备份模式，对系统的压力等

### 分库分表原则

原则一：能不分就不分，1000 万以内的表，不建议分片，通过合适的索引，读写分离等方式，可以很好的解 决性能问题。

原则二：分片数量尽量少，分片尽量均匀分布在多个 DataHost 上，因为一个查询 SQL 跨分片越多，则总体 性能越差，虽然要好于所有数据在一个分片的结果，只在必要的时候进行扩容，增加分片数量。

原则三：分片规则需要慎重选择，分片规则的选择，需要考虑数据的增长模式，数据的访问模式，分片关联 性问题，以及分片扩容问题，最近的分片策略为范围分片，枚举分片，一致性 Hash 分片，这几种分片都有利于 扩容

原则四：尽量不要在一个事务中的 SQL 跨越多个分片，分布式事务一直是个不好处理的问题

原则五：查询条件尽量优化，尽量避免 Select * 的方式，大量数据结果集下，会消耗大量带宽和 CPU 资源， 查询尽量避免返回大量结果集，并且尽量为频繁使用的查询语句建立索引。

### 数据拆分原则

1. 达到一定数量级才拆分（800 万） 
2. 不到 800 万但跟大表（超 800 万的表）有关联查询的表也要拆分，在此称为大表关联表 
3. 大表关联表如何拆：小于 100 万的使用全局表；大于 100 万小于 800 万跟大表使用同样的拆分策略；无 法跟大表使用相同规则的，可以考虑从 java 代码上分步骤查询，不用关联查询，或者破例使用全局表。
4. 破例的全局表：如 item_sku 表 250 万，跟大表关联了，又无法跟大表使用相同拆分策略，也做成了全局 表。破例的全局表必须满足的条件：没有太激烈的并发 update，如多线程同时 update 同一条 id=1 的记录。虽 有多线程 update，但不是操作同一行记录的不在此列。多线程 update 全局表的同一行记录会死锁。批量 insert 没问题。
5. 拆分字段是不可修改的 
6. 拆分字段只能是一个字段，如果想按照两个字段拆分，必须新建一个冗余字段，冗余字段的值使用两个字 段的值拼接而成（如大区+年月拼成 zone_yyyymm 字段）。
7. 拆分算法的选择和合理性评判：按照选定的算法拆分后每个库中单表不得超过 800 万 
8. 能不拆的就尽量不拆。如果某个表不跟其他表关联查询，数据量又少，直接不拆分，使用单库即可。

### Mycat限制

* 除了分片规则相同、ER 分片、全局表、以及 SharedJoin，其他表之间的 Join 问题目前还没有很好的解决，需要自己编写 Catlet 来处理。
* 不支持 Insert into 中不包括字段名的 SQL。
* insert into x select from y 的 SQL，若 x 与 y 不是相同的分片规则，则不被支持，此时会涉及到跨分片转移
* 跨分片的事务，目前只是弱 XA 模式，还没完全实现 XA 模式
* 分片的 Table，目前不能执行 Lock Table 这样的语句，因为这种语句会随机发到某个节点，也不会全部分片锁定，经常导致死锁问题，此类问题常常出现在 sqldump 导入导出 SQL 数据的过程中。
* 目前 sql 解析器采用 Druid,再某些 sql 例如 order，group，sum ，count 条件下，如果这类操作会出现兼容问题，比如：
```
select t.name as name1 from test order by t.name
```
    这条语句 select 列的别名与 order by 不一致解析器会出现异常，所以在对列加别名时候要注意这类操作异常，特别是由 jpa 等类似的框架生成的语句会有兼容问题。

开发框架方面，虽然支持 Hibernat，但不建议使用 Hibernat，而是建议 Mybatis 以及直接 JDBC 操作，原因 Hibernat 无法控制 SQL 的生成，无法做到对查询 SQL 的优化，导致大数量下的性能问题。此外，事务方面，
建议自己手动控制，查询语句尽量走自动提交事务模式，这样 Mycat 的读写分离会被用到，提升性能很明显。