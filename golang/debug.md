# gdb debug

除通过日志方式调试外，Golang还支持使用gdb调试，gdb版本要求7.1以上版本。

使用go build -gcflags "-N -l"指令编译代码，使用“gdb 可执行文件”调试。

gdb相关调试指令参考http://golang.org/doc/gdb。

其中goroutine相关调试指令需要进行以下配置：
编辑$HOME/.gdbinit文件，添加add-auto-load-safe-path /usr/local/go/src/runtime/runtime-gdb.py