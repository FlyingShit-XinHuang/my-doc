## MyISAM和InnoDB区别

* InnoDB支持事务，MyISAM不支持
* MyISAM适合查询以及插入为主的应用，InnoDB适合频繁修改以及涉及到安全性较高的应用
* InnoDB支持外键，MyISAM不支持
* InnoDB不支持FULLTEXT类型的索引
* InnoDB中不保存表的行数，如select count(*) from table时，InnoDB需要扫描一遍整个表来计算有多少行，但是MyISAM只要简单的读出保存好的行数即可。注意的是，当count(*)语句包含where条件时MyISAM也需要扫描整个表
* 清空整个表时，InnoDB是一行一行的删除，效率非常慢。MyISAM则会重建表
* InnoDB支持行锁，MyISAM只支持表锁

## 参考资料

* https://www.cnblogs.com/kevingrace/p/5685355.html