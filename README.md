# goapollo

Golang 实现的Apollo 客户端。

支持从命令行、配置文件、环境变量启动，将服务器端变动的配置定时拉取并保存到本地文件中。

保存的物理路径支持多文件，一般格式为：

```bash
//文件格式:文件路径;文件格式:文件路径
fastcgi:./runtime/conf/nginx.params;ini:./runtime/conf/nginx.conf
```

目前支持的保存的文件格式有：`fastcgi`、`ini`、 `yml`、`yaml`、`json`。

## 关联命名空间

目前，键值对格式的配置支持一个主命名空间和多个关联命名空间，关联命名空间必须和主命名空间在一个应用内。

如果存在主命名空间的配置与关联的配置冲突，则主配置会覆盖关联配置。

如果需要关联多个命名空间，需要使用英文下的";"分隔。


## 命令行启动

命令行启动只支持监控一个应用。

```bash
./goapollo run -file=./runtime/nginx.conf -port=8081 -addr= -app_id=app123456 -cluster=default -namespace=application -related=dev_business.nginx -server_url=https://dev.xx.com/ -long_interval=60 -full_interval=30
```

参数：

- `-file` 将配置保存的文件路径
- `-port` 启动服务监听的HTTP端口
- `-addr` 如果设置了监听地址，则使用设置的监听地址开放HTTP服务，否则监听所有地址
- `-app_id` 应用的APPID
- `-cluster` 集群名称
- `-namespace` 命名空间
- `-server_url` Apollo 服务器地址
- `-long_interval` 通知轮询时间间隔，单位是秒
- `-full_interval` 全量拉取轮询时间间隔，单位是秒
- `log_path` 日志输出目录
- `related` 关联命名空间，多个用";"分隔

## 环境变量启动

环境变量启动也仅支持监控一个应用。

```bash
./goapollo run 
```

环境变量值：

- `APOLLO_SAVE_PATH` 将配置保存的文件路径
- `APOLLO_HTTP_PORT` 启动服务监听的HTTP端口
- `APOLLO_HTTP_ADDR` 如果设置了监听地址，则使用设置的监听地址开放HTTP服务，否则监听所有地址
- `APOLLO_APP_ID` 应用的APPID
- `APOLLO_CLUSTER` 集群名称
- `APOLLO_NAMESPACE_NAME` 命名空间
- `APOLLO_SERVER_URL` Apollo 服务器地址
- `APOLLO_LONG_INTERVAL` 通知轮询时间间隔，单位是秒，建议设置为 1s,过大的间隔会导致通知延迟
- `APOLLO_FULL_INTERVAL`  全量拉取轮询时间间隔，单位是秒
- `APOLLO_LOG_PATH` 日志输出目录
- `APOLLO_RELATED` 关联的辅助命名空间，拉取配置时，如果与主命名空间冲突，会被主命名空间的配置覆盖，多个用";"分隔

## 配置文件

通过配置文件可以监控多个应用，启动命令如下：

```bash
./goapollo run -c=conf/app.conf
```

配置文件如下：

```ini
addr=
port=8081
log_path=

[app:6e77bd897fe903ac:nginx]
;应用ID
appId=6e77bd897fe903ac

;命名空间
namespace=TEST1.nginx

;相关联的命名空间
related=dev_business.nginx

;服务器地址
serverUrl=http://dev.config.baidu.com/

;默认集群
cluster=default

;保存的文件地址
saveFile=nginx:./runtime/conf/nginx.params;ini:./runtime/conf/nginx.conf

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

