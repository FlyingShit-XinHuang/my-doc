#nodejs

## nodejs安装：

具体参考https://nodejs.org/en/download/package-manager/

执行下面两条命令：
1. curl -sL https://deb.nodesource.com/setup_4.x | sudo -E bash -
2. sudo apt-get install nodejs

## 调试方法：

具体参考http://www.cnblogs.com/moonz-wu/archive/2012/01/15/2322120.html

安装node-inspector：
sudo npm install -g node-inspector

启动node-inspector：
node-inspector --web-port=8081 &
其中--web-port用来指定监听端口，不指定时默认为8080

启动程序：
node --debug <your_nodejs_program>.js

在chrome浏览器中访问http://127.0.0.1:8081/debug?port=5858，即可在chrome浏览器中调试程序。

## profile方法：

具体参考https://nodejs.org/en/docs/guides/simple-profiling/

参考代码位于profile文件夹

须提前安装express和ab：
sudo npm install -g express
sudo apt-get install apache2-utils # for ab installation

运行程序时指定--prof参数：
NODE_ENV=production node --prof app.js

当前目录会生成isolate-0xnnnnnnnnnnnn-v8.log，使用以下指令生成可读文件：
node --prof-process isolate-0xnnnnnnnnnnnn-v8.log > processed.txt