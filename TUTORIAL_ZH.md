Leaf 游戏服务器框架简介
==================

Leaf 是一个开发效率和执行效率并重的开源游戏服务器框架。Leaf 的关注点：

* 良好的使用体验。Leaf 总是尽可能的提供简洁和易用的接口，尽可能的提升开发的效率
* 稳定性。Leaf 总是尽可能的恢复运行过程中的错误，避免崩溃
* 多核支持。Leaf 通过模块机制和 [leaf/go](https://github.com/name5566/leaf/tree/master/go) 尽可能的利用多核资源，同时又尽量避免各种副作用
* 模块机制。

Leaf 的模块机制
---------------

一个 Leaf 开发的游戏服务器由多个模块组成（例如 [LeafServer](https://github.com/name5566/leafserver)），模块有以下特点：

* 每个模块运行在一个单独的 goroutine 中
* 模块间通过一套轻量的 RPC 机制通讯（[leaf/chanrpc](https://github.com/name5566/leaf/tree/master/chanrpc)）

Leaf 不建议在游戏服务器中设计过多的模块。

游戏服务器在启动时进行模块的注册，例如：

```go
leaf.Run(
	game.Module,
	gate.Module,
	login.Module,
)
```

这里按顺序注册了 game、gate、login 三个模块。每个模块都需要实现接口：

```go
type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan bool)
}
```

Leaf 首先会在同一个 goroutine 中按模块注册顺序执行模块的 OnInit 方法，等到所有模块 OnInit 方法执行完成后则为每一个模块启动一个 goroutine 并执行模块的 Run 方法。最后，游戏服务器关闭时（Ctrl + C 关闭游戏服务器）将按模块注册相反顺序在同一个 goroutine 中执行模块的 OnDestroy 方法。

Leaf 源码概览
---------------

* leaf/chanrpc 提供了一套基于 channel 的 RPC 机制，用于游戏服务器模块间通讯
* leaf/db 数据库相关，目前支持 [MongoDB](https://www.mongodb.org/)
* leaf/gate 网关模块，负责游戏客户端的接入
* leaf/go 用于创建能够被 Leaf 管理的 goroutine
* leaf/log 日志相关
* leaf/network 网络相关，使用 TCP 协议，可自定义消息格式，目前 Leaf 提供了基于 [protobuf](https://developers.google.com/protocol-buffers) 和 JSON 的消息格式
* leaf/recordfile 用于管理游戏数据
* leaf/timer 定时器相关
* leaf/util 辅助库

使用 Leaf 开发游戏服务器
---------------

[LeafServer](https://github.com/name5566/leafserver) 是一个基于 Leaf 开发的游戏服务器，我们以 LeafServer 作为起点。

获取 LeafServer：

```
git clone https://github.com/name5566/leafserver
```

设置 leafserver 目录到 GOPATH 后获取相关依赖：

```
go get github.com/name5566/leaf
go get github.com/golang/protobuf/proto
go get gopkg.in/mgo.v2
```

编译 LeafServer：

```
go install server
```

如果一切顺利，运行 server 你可以获得以下输出：

```
2015/08/26 22:11:27 [release] Leaf starting up
```

敲击 Ctrl + C 关闭游戏服务器，服务器正常关闭输出：

```
2015/08/26 22:12:30 [release] Leaf closing down (signal: interrupt)
```

### Hello Leaf

现在，在 LeafServer 的基础上，我们来看看游戏服务器如何接收和处理网络消息。

首先定义一个 JSON 格式的消息（protobuf 类似）。打开 LeafServer msg/msg.go 文件可以看到如下代码：

```go
package msg

import (
	"github.com/name5566/leaf/network/json"
	"github.com/name5566/leaf/network/protobuf"
)

var (
	JSONProcessor     = json.NewProcessor()
	ProtobufProcessor = protobuf.NewProcessor()
)

func init() {

}
```

我们尝试添加一个名字为 Hello 的消息（msg/msg.go 文件中未改动部分这里略去）：

```go
func init() {
	// 这里我们注册了一个 JSON 消息 Hello
	// 我们也可以使用 ProtobufProcessor 注册 protobuf 消息（同时注意修改配置文件 conf/conf.go 中的 Encoding）
	JSONProcessor.Register(&Hello{})
}

// 一个结构体定义了一个 JSON 消息的格式
// 消息名为 Hello
type Hello struct {
	Name string
}
```

客户端发送到游戏服务器的消息需要通过 gate 模块路由，简而言之，gate 模块决定了某个消息具体交给内部的哪个模块来处理。这里，我们将 Hello 消息路由到 game 模块中。打开 LeafServer gate/router.go，敲入如下代码：

```go
package gate

import (
	"server/game"
	"server/msg"
)

func init() {
	// 这里指定消息 Hello 路由到 game 模块
	// 模块间使用 ChanRPC 通讯，消息路由也不例外
	msg.JSONProcessor.SetRouter(&msg.Hello{}, game.ChanRPC)
}
```

一切就绪，我们现在可以在 game 模块中处理 Hello 消息了。打开 LeafServer game/internal/handler.go，敲入如下代码：

```go
package internal

import (
	"github.com/name5566/leaf/log"
	"reflect"
	"server/msg"
)

func init() {
	// 向当前模块（game 模块）注册 Hello 消息的消息处理函数 handleHello
	handler(&msg.Hello{}, handleHello)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleHello(args []interface{}) {
	// 收到的 Hello 消息
	m := args[0].(*msg.Hello)
	// 消息的发送者
	a := args[1].(gate.Agent)

	// 输出收到的消息的内容
	log.Debug("hello %v", m.Name)

	// 给发送者回应一个 Hello 消息
	a.WriteMsg(&msg.Hello{
		Name: "client",
	})
}
```

到这里，一个简单的范例就完成了。为了更加清楚的了解消息的格式，我们从 0 编写一个最简单的测试客户端。

Leaf 中，在网络中传输的消息都会使用以下格式：

```
--------------
| len | data |
--------------
```

其中：

1. len 表示了 data 部分的长度（字节数）。len 本身也有长度，默认为 2 字节（可配置），len 本身的长度决定了单个消息的最大大小
2. data 部分使用 JSON 或者 protobuf 编码（也可自定义其他编码方式）

测试客户端同样使用 Go 语言编写：
```go
package main

import (
	"encoding/binary"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3563")
	if err != nil {
		panic(err)
	}

	// Hello 消息（JSON 格式）
	// 对应游戏服务器 Hello 消息结构体
	data := []byte(`{
		"Hello": {
			"Name": "leaf"
		}
	}`)

	// len + data
	m := make([]byte, 2+len(data))

	// 默认使用大端序
	binary.BigEndian.PutUint16(m, uint16(len(data)))

	copy(m[2:], data)

	// 发送消息
	conn.Write(m)
}
```

执行此测试客户端，服务器输出：

```
2015/08/26 23:26:23 [debug  ] hello leaf
```
