# Redis延迟监控

Redis 2.8.13版本引入了延迟监控功能，帮助用户检测可能的延迟问题。延迟监控逻辑上由以下部分组成：

* hooks。在不同影响延迟的代码中取样。
* 不同事件的延迟峰值的时间序列记录
* 报告引擎。从时间序列中抓取元数据。
* 分析引擎。根据测量结果提供可读的报告和提示。

## 事件和时间序列

被监控的不同代码路径有不同的名字，它们被称为事件。如command是测量可能执行缓慢的指令的延迟峰值的事件，fast-command是O(1)和O(log N)指令的监控事件名称。延迟峰值是运行时间超过配置的延迟阈值的事件。每个监控事件都有对应的时间序列，时间序列工作原理是：

* 每次延迟峰值发生时，它会被记录到合适的时间序列。
* 每个时间序列由160个元素组成。
* 每个元素都有两部分：时间峰值测量的时间戳，事件执行消耗的时间（毫秒）。
* 在同一秒内发生的同一事件的延迟峰值会被合并（使用最大延迟）。
* 每个元素都会记录最大的延迟。

## 延迟监控开关

第一步是设置延迟阈值（毫秒），只有超过阈值的事件会被记录，用户要根据具体需求设置。使用以下指令可以在运行的生产环境中开启延迟监控：

<pre>
CONFIG SET latency-monitor-threshold 100
</pre>

监控默认是关闭的（阈值为0）。

## 使用LATENCY指令

### LATENCY LATEST

LATENCY LATEST指令会报告最后一次记录的延迟事件，每个事件有以下字段：

* 事件名称
* 最后一次延迟峰值的时间戳
* 延迟时间（毫秒）
* 该事件最大的延迟时间。

以下是输出内容的示例：

<pre>
127.0.0.1:6379> debug sleep 1
OK
(1.00s)
127.0.0.1:6379> debug sleep .25
OK
127.0.0.1:6379> latency latest
1) 1) "command"
   2) (integer) 1405067976
   3) (integer) 251
   4) (integer) 1001
</pre>

### LATENCY RESET

LATENCY RESET指令不指定参数时将重置所有事件，丢弃目前已记录的内容，重置最大延迟时间。参数中指定事件名称时，将重置这些事件。该指令返回重置的事件时间序列数量。

### LATENCY GRAPH

该指令会生成指定事件的ASCII码风格的图像，例如：

<pre>
127.0.0.1:6379> latency reset command
(integer) 0
127.0.0.1:6379> debug sleep .1
OK
127.0.0.1:6379> debug sleep .2
OK
127.0.0.1:6379> debug sleep .3
OK
127.0.0.1:6379> debug sleep .5
OK
127.0.0.1:6379> debug sleep .4
OK
127.0.0.1:6379> latency graph command
command - high 500 ms, low 101 ms (all time high 500 ms)
--------------------------------------------------------------------------------
   #_
  _||
 _|||
_||||

11186
542ss
sss
</pre>

每一列最下面的垂直字符表示事件发生在多久之前。如例中第一列表示事件发生在15秒前。每一列上面的图表示延迟的度量，下划线表示最小，#号表示最大，o表示中等。

graph子命令用于快速查看延迟事件的趋势而不需使用额外的工具，也不需解析LATENCY HISTORY提供的原数据。

### LATENCY DOCTOR

LATENCY DOCTOR命令是延迟监控最强大的分析工具，它可以提供额外的统计数据（如延迟峰值间的平均周期，中位数，可读的事件分析）。

<pre>
127.0.0.1:6379> latency doctor

Dave, I have observed latency spikes in this Redis instance.
You don't mind talking about it, do you Dave?

1. command: 5 latency spikes (average 300ms, mean deviation 120ms,
   period 73.40 sec). Worst all time event 500ms.

I have a few advices for you:

- Your current Slow Log configuration only logs events that are
  slower than your configured latency monitor threshold. Please
  use 'CONFIG SET slowlog-log-slower-than 1000'.
- Check your Slow Log to understand what are the commands you are
  running which are too slow to execute. Please check
  http://redis.io/commands/slowlog for more information.
- Deleting, expiring or evicting (because of maxmemory policy)
  large objects is a blocking operation. If you have very large
  objects that are often deleted, expired, or evicted, try to
  fragment those objects into multiple smaller objects.
  </pre>