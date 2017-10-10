Leaf 游戏服务器框架简介
==================

Leaf 是一个由 Go 语言（golang）编写的开发效率和执行效率并重的开源游戏服务器框架。Leaf 适用于各类游戏服务器的开发，包括 H5（HTML5）游戏服务器。

Leaf 的关注点：

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
* leaf/network 网络相关，使用 TCP 和 WebSocket 协议，可自定义消息格式，默认 Leaf 提供了基于 [protobuf](https://developers.google.com/protocol-buffers) 和 JSON 的消息格式
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

设置 leafserver 目录到 GOPATH 环境变量后获取 Leaf：

```
go get github.com/name5566/leaf
```

编译 LeafServer：

```
go install server
```

如果一切顺利，运行 server 你可以获得以下输出：

```
2015/08/26 22:11:27 [release] Leaf 1.1.2 starting up
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
	"github.com/name5566/leaf/network"
)

var Processor network.Processor

func init() {

}
```

Processor 为消息的处理器（可由用户自定义），这里我们使用 Leaf 默认提供的 JSON 消息处理器并尝试添加一个名字为 Hello 的消息：

```go
package msg

import (
	"github.com/name5566/leaf/network/json"
)

// 使用默认的 JSON 消息处理器（默认还提供了 protobuf 消息处理器）
var Processor = json.NewProcessor()

func init() {
	// 这里我们注册了一个 JSON 消息 Hello
	Processor.Register(&Hello{})
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
	msg.Processor.SetRouter(&msg.Hello{}, game.ChanRPC)
}
```

一切就绪，我们现在可以在 game 模块中处理 Hello 消息了。打开 LeafServer game/internal/handler.go，敲入如下代码：

```go
package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/gate"
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

Leaf 中，当选择使用 TCP 协议时，在网络中传输的消息都会使用以下格式：

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

执行此测试客户端，游戏服务器输出：

```
2015/09/25 07:41:03 [debug  ] hello leaf
2015/09/25 07:41:03 [debug  ] read message: read tcp 127.0.0.1:3563->127.0.0.1:54599: wsarecv: An existing connection was forcibly closed by the remote host.
```

测试客户端发送完消息以后就退出了，此时和游戏服务器的连接断开，相应的，游戏服务器输出连接断开的提示日志（第二条日志，日志的具体内容和 Go 语言版本有关）。

除了使用 TCP 协议外，还可以选择使用 WebSocket 协议（例如开发 H5 游戏）。Leaf 可以单独使用 TCP 协议或 WebSocket 协议，也可以同时使用两者，换而言之，服务器可以同时接受 TCP 连接和 WebSocket 连接，对开发者而言消息来自 TCP 还是 WebSocket 是完全透明的。现在，我们来编写一个对应上例的使用 WebSocket 协议的客户端：
```html
<script type="text/javascript">
var ws = new WebSocket('ws://127.0.0.1:3653')

ws.onopen = function() {
    // 发送 Hello 消息
    ws.send(JSON.stringify({Hello: {
        Name: 'leaf'
    }}))
}
</script>
```

保存上述代码到某 HTML 文件中并使用（任意支持 WebSocket 协议的）浏览器打开。在打开此 HTML 文件前，首先需要配置一下 LeafServer 的 bin/conf/server.json 文件，增加 WebSocket 监听地址（WSAddr）：
```json
{
    "LogLevel": "debug",
    "LogPath": "",
    "TCPAddr": "127.0.0.1:3563",
    "WSAddr": "127.0.0.1:3653",
    "MaxConnNum": 20000
}
```

重启游戏服务器后，方可接受 WebSocket 消息：

```
2015/09/25 07:50:03 [debug  ] hello leaf
```

在 Leaf 中使用 WebSocket 需要注意的一点是：Leaf 总是发送二进制消息而非文本消息。

### Leaf 模块详解

LeafServer 中包含了 3 个模块，它们分别是：

* gate 模块，负责游戏客户端的接入
* login 模块，负责登录流程
* game 模块，负责游戏主逻辑

一般来说（而非强制规定），从代码结构上，一个 Leaf 模块：

1. 放置于一个目录中（例如 game 模块放置于 game 目录中）
2. 模块的具体实现放置于 internal 包中（例如 game 模块的具体实现放置于 game/internal 包中）

每个模块下一般有一个 external.go 的文件，顾名思义表示模块对外暴露的接口，这里以 game 模块的 external.go 文件为例：

```go
package game

import (
	"server/game/internal"
)

var (
	// 实例化 game 模块
	Module  = new(internal.Module)
	// 暴露 ChanRPC
	ChanRPC = internal.ChanRPC
)
```

首先，模块会被实例化，这样才能注册到 Leaf 框架中（详见 LeafServer main.go），另外，模块暴露的 ChanRPC 被用于模块间通讯。

进入 game 模块的内部（LeafServer game/internal/module.go）：

```go
package internal

import (
	"github.com/name5566/leaf/module"
	"server/base"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
}

func (m *Module) OnDestroy() {

}
```

模块中最关键的就是 skeleton（骨架），skeleton 实现了 Module 接口的 Run 方法并提供了：

* ChanRPC
* goroutine
* 定时器

### Leaf ChanRPC

由于 Leaf 中，每个模块跑在独立的 goroutine 上，为了模块间方便的相互调用就有了基于 channel 的 RPC 机制。一个 ChanRPC 需要在游戏服务器初始化的时候进行注册（注册过程不是 goroutine 安全的），例如 LeafServer 中 game 模块注册了 NewAgent 和 CloseAgent 两个 ChanRPC：

```go
package internal

import (
	"github.com/name5566/leaf/gate"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {

}

func rpcCloseAgent(args []interface{}) {

}
```

使用 skeleton 来注册 ChanRPC。RegisterChanRPC 的第一个参数是 ChanRPC 的名字，第二个参数是 ChanRPC 的实现。这里的 NewAgent 和 CloseAgent 会被 LeafServer 的 gate 模块在连接建立和连接中断时调用。ChanRPC 的调用方有 3 种调用模式：

1. 同步模式，调用并等待 ChanRPC 返回
2. 异步模式，调用并提供回调函数，回调函数会在 ChanRPC 返回后被调用
3. Go 模式，调用并立即返回，忽略任何返回值和错误

gate 模块这样调用 game 模块的 NewAgent ChanRPC（这仅仅是一个示例，实际的代码细节复杂的多）：

```go
game.ChanRPC.Go("NewAgent", a)
```

这里调用 NewAgent 并传递参数 a，我们在 rpcNewAgent 的参数 args[0] 中可以取到 a（args[1] 表示第二个参数，以此类推）。

更加详细的用法可以参考 [leaf/chanrpc](https://github.com/name5566/leaf/blob/master/chanrpc)。需要注意的是，无论封装多么精巧，跨 goroutine 的调用总不能像直接的函数调用那样简单直接，因此除非必要我们不要构建太多的模块，模块间不要太频繁的交互。模块在 Leaf 中被设计出来最主要是用于划分功能而非利用多核，Leaf 认为在模块内按需使用 goroutine 才是多核利用率问题的解决之道。

### Leaf Go

善用 goroutine 能够充分利用多核资源，Leaf 提供的 Go 机制解决了原生 goroutine 存在的一些问题：

* 能够恢复 goroutine 运行过程中的错误
* 游戏服务器会等待所有 goroutine 执行结束后才关闭
* 非常方便的获取 goroutine 执行的结果数据
* 在一些特殊场合保证 goroutine 按创建顺序执行

我们来看一个例子（可以在 LeafServer 的模块的 OnInit 方法中测试）：

```go
log.Debug("1")

// 定义变量 res 接收结果
var res string

skeleton.Go(func() {
	// 这里使用 Sleep 来模拟一个很慢的操作
	time.Sleep(1 * time.Second)

	// 假定得到结果
	res = "3"
}, func() {
	log.Debug(res)
})

log.Debug("2")
```

上面代码执行结果如下：

```go
2015/08/27 20:37:17 [debug  ] 1
2015/08/27 20:37:17 [debug  ] 2
2015/08/27 20:37:18 [debug  ] 3
```

这里的 Go 方法接收 2 个函数作为参数，第一个函数会被放置在一个新创建的 goroutine 中执行，在其执行完成之后，第二个函数会在当前 goroutine 中被执行。由此，我们可以看到变量 res 同一时刻总是只被一个 goroutine 访问，这就避免了同步机制的使用。Go 的设计使得 CPU 得到充分利用，避免操作阻塞当前 goroutine，同时又无需为共享资源同步而忧心。

更加详细的用法可以参考 [leaf/go](https://github.com/name5566/leaf/blob/master/go)。

### Leaf timer

Go 语言标准库提供了定时器的支持：

```go
func AfterFunc(d Duration, f func()) *Timer
```

AfterFunc 会等待 d 时长后调用 f 函数，这里的 f 函数将在另外一个 goroutine 中执行。Leaf 提供了一个相同的 AfterFunc 函数，相比之下，f 函数在 AfterFunc 的调用 goroutine 中执行，这样就避免了同步机制的使用：

```go
skeleton.AfterFunc(5 * time.Second, func() {
	// ...
})
```

另外，Leaf timer 还支持 [cron 表达式](https://en.wikipedia.org/wiki/Cron)，用于实现诸如“每天 9 点执行”、“每周末 6 点执行”的逻辑。

更加详细的用法可以参考 [leaf/timer](https://github.com/name5566/leaf/blob/master/timer)。

### Leaf log

Leaf 的 log 系统支持多种日志级别：

1. Debug 日志，非关键日志
2. Release 日志，关键日志
3. Error 日志，错误日志
4. Fatal 日志，致命错误日志

Debug < Release < Error < Fatal（日志级别高低）

在 LeafServer 中，bin/conf/server.json 可以配置日志级别，低于配置的日志级别的日志将不会输出。Fatal 日志比较特殊，每次输出 Fatal 日志之后游戏服务器进程就会结束，通常来说，只在游戏服务器初始化失败时使用 Fatal 日志。

我们还可以通过配置 LeafServer conf/conf.go 的 LogFlag 来在日志中输出文件名和行号：

```
LogFlag = log.Lshortfile
```

可用的 LogFlag 见：[https://golang.org/pkg/log/#pkg-constants](https://golang.org/pkg/log/#pkg-constants)


更加详细的用法可以参考 [leaf/log](https://github.com/name5566/leaf/blob/master/log)。

### Leaf recordfile

Leaf 的 recordfile 是基于 CSV 格式（范例见[这里](https://github.com/name5566/leaf/blob/master/recordfile/test.txt)）。recordfile 用于管理游戏配置数据。在 LeafServer 中使用 recordfile 非常简单：

1. 将 CSV 文件放置于 bin/gamedata 目录中
2. 在 gamedata 模块中调用函数 readRf 读取 CSV 文件

范例：

```go
// 确保 bin/gamedata 目录中存在 Test.txt 文件
// 文件名必须和此结构体名称相同（大小写敏感）
// 结构体的一个实例映射 recordfile 中的一行
type Test struct {
	// 将第一列按 int 类型解析
	// "index" 表明在此列上建立唯一索引
	Id  int "index"
	// 将第二列解析为长度为 4 的整型数组
	Arr [4]int
	// 将第三列解析为字符串
	Str string
}

// 读取 recordfile Test.txt 到内存中
// RfTest 即为 Test.txt 的内存镜像
var RfTest = readRf(Test{})

func init() {
	// 按索引查找
	// 获取 Test.txt 中 Id 为 1 的那一行
	r := RfTest.Index(1)

	if r != nil {
		row := r.(*Test)

		// 输出此行的所有列的数据
		log.Debug("%v %v %v", row.Id, row.Arr, row.Str)
	}
}
```

更加详细的用法可以参考 [leaf/recordfile](https://github.com/name5566/leaf/blob/master/recordfile)。

了解更多
---------------

阅读 Wiki 获取更多的帮助：[https://github.com/name5566/leaf/wiki](https://github.com/name5566/leaf/wiki)
