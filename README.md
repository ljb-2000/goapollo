# goapollo

Golang 实现的Apollo 客户端。

支持从命令行、配置文件、环境变量启动，将服务器端变动的配置定时拉取并保存到本地文件中。

## 命令行启动

命令行启动只支持监控一个应用。

```bash
./goapollo run -file=./runtime/nginx.conf -port=8081 -addr= -app_id=app123456 -cluster=default -namespace=application -server_url=https://dev.xx.com/ -long_interval=60 -full_interval=30
```

参数：

`-file` 将配置保存的文件路径
`-port` 启动服务监听的HTTP端口
`-addr` 如果设置了监听地址，则使用设置的监听地址开放HTTP服务，否则监听所有地址
`-app_id` 应用的APPID
`-cluster` 集群名称
`-namespace` 命名空间
`-server_url` Apollo 服务器地址
`-long_interval` 通知轮询时间间隔，单位是秒
`-full_interval` 全量拉取轮询时间间隔，单位是秒

## 环境变量启动

环境变量启动也仅支持监控一个应用。

```bash
./goapollo run 
```

环境变量值：

`APOLLO_SAVE_PATH` 将配置保存的文件路径
`APOLLO_HTTP_PORT` 启动服务监听的HTTP端口
`APOLLO_HTTP_ADDR` 如果设置了监听地址，则使用设置的监听地址开放HTTP服务，否则监听所有地址
`APOLLO_APP_ID` 应用的APPID
`APOLLO_CLUSTER` 集群名称
`APOLLO_NAMESPACE_NAME` 命名空间
`APOLLO_SERVER_URL` Apollo 服务器地址
`APOLLO_LONG_INTERVAL` 通知轮询时间间隔，单位是秒
`APOLLO_FULL_INTERVAL`  全量拉取轮询时间间隔，单位是秒

## 配置文件

通过配置文件可以监控多个应用，启动命令如下：

```bash
./goapollo run -c=conf/app.conf
```

配置文件如下：

```ini
addr=
port=8081

[app:6e77bd897fe903ac:nginx]
;应用ID
appId=6e77bd897fe903ac

;命名空间
namespace=TEST1.nginx

;服务器地址
serverUrl=http://dev.config.baidu.com/

;默认集群
cluster=default

;保存的文件地址
saveFile=./runtime/conf/nginx.params

;通知轮换时间间隔
longPollInterval=70

[app:6e77bd897fe903ac:json]
;应用ID
appId=6e77bd897fe903ac

;命名空间
namespace=TEST1JSON.json

;服务器地址
serverUrl=http://dev.config.baidu.com/

;默认集群
cluster=default

;保存的文件地址
saveFile=./runtime/conf/nginx.json

;通知轮换时间间隔
longPollInterval=60

[app:6e77bd897fe903ac:ini]
;应用ID
appId=6e77bd897fe903ac

;命名空间
namespace=TEST1.ini

;服务器地址
serverUrl=http://dev.config.baidu.com/

;默认集群
cluster=default

;保存的文件地址
saveFile=./runtime/conf/nginx.ini

;通知轮换时间间隔
longPollInterval=60
```

