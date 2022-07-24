# 暴露边缘服务 demo 代码

## 设计

TODO: https://www.rectcircle.cn/

## 运行

### 启动相关服务

```bash
# 运行位于边缘设备的服务 demo1 和 demo2 (守护进程)
go run ./cmd/edgeservice/demo1
go run ./cmd/edgeservice/demo2

# 运行位于机房的 Exposer Server  (集群)
go run ./cmd/exposer/server

# 运行位于边缘设备的 Exposer Client (守护进程)
go run ./cmd/exposer/client

# 运行位于机房的协议转换器服务 (集群) http 和 tcp
go run ./cmd/protoconv/http
go run ./cmd/protoconv/tcp
```

### 测试和输出

```bash
# 通过 http 转换器服务，访问 demo1 和 demo2
curl localhost:9000 -H 'X-Edge-Device-ID: DEVICE-0000' -H 'X-Edge-Service-ID: demo1'
# 输出: Hello, world! service id is demo1,  port is 8081
curl localhost:9000 -H 'X-Edge-Device-ID: DEVICE-0000' -H 'X-Edge-Service-ID: demo2'
# 输出: Hello, world! service id is demo2,  port is 8082

# 通过 tcp 转换器服务，访问 demo2
curl localhost:9001
# 输出: Hello, world! service id is demo2,  port is 8082
```
