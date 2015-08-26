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

获取 LeafServer

```
git clone https://github.com/name5566/leafserver
```

设置 leafserver 目录到 GOPATH 后获取相关依赖：

```
go get github.com/name5566/leaf
go get github.com/golang/protobuf/proto
```

编译 LeafServer：

```
go install server
```

如果一切顺利，运行 server 你可以获得以下输出：

```
2015/08/26 22:11:27 [release] Leaf starting up
```