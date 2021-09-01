# 介绍

主要用于腾讯云物联网平台设备调试，有些时候虚拟设备不一定能满足需求，需要 mock 真实设备的情况。
# 安装

右侧 Release 处下载自己对应平台的压缩包。
MacOS Intel 版对应 darwin-amd64
MacOS M1 版对应 darwin-arm64
# 使用方法

1. 「访达」直接打开，或使用`tar -zxvf ./xxxx.tar.gz -C ./target_dir` 将下载到的 `tar.gz` 文件解压到 `target_dir` 目录(该目录必须存在)。
2. 运行
   - 参照 config_sample.yaml 内的内容，修改文件名为 `config.yaml`，将二进制文件与配置放到同一个文件夹下，运行二进制文件即可。
   - 使用 `fake_mqtt_device -c ./config.yaml` 指定配置文件。

# Q & A

## MacOS 打开提示未受信任

请参考: https://support.apple.com/zh-cn/guide/mac-help/mh40616/mac