# 日志目录

默认在/tmp/下，-log_dir=参数可设置

# 日志级别

通过-v=参数可设置，小于等于参数的日志会输出。

在stderr默认不输出Info日志，可通过-logtostderr=true、-alsologtostderr=true、-stderrthreshold=INFO三个参数来设置。