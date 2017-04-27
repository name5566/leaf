Brief introduction to Leaf
==========================

Leaf, written in Go, is a open source game server framework aiming to boost the efficiency both in development and runtime.

Leaf champions below philosophies:

* Simple APIs. Leaf tends to provide simple and plain interfaces which are always best for use.
* Self-healing. Leaf always tries to salvage the process from runtime errors instead of leaving it to crash.
* Multi-core support. Leaf utilize its modules and [leaf/go](https://github.com/name5566/leaf/tree/master/go) to make use of CPU resouces at maximum while avoiding varieties of side effects may be caused.

* Module-based.

Leaf's Modules
--------------

A game server implemented with Leaf may include many modules (e.g. [LeafServer](https://github.com/name5566/leafserver)) which all share below traits:

* Each module runs inside a separate goroutine
* Modules communicate with one another via a light weight RPC channel([leaf/chanrpc](https://github.com/name5566/leaf/tree/master/chanrpc))

Leaf suggests not to take in too many modules in your game server implementation.

Modules are registered at the beginning of program as below

```go
leaf.Run(
    game.Module,
    gate.Module,
    login.Module,
)
```

The modules of `game`, `gate` and `gate` are registered consecutively. They are required to implement a `Module` interface.

```go
type Module interface {
    OnInit()
    OnDestroy()
    Run(closeSig chan bool)
}
```

Leaf follows below steps to manage modules:

1. Takes turns (FIFO) to register the given modules by calling `OnInit()`' in a parent goroutine
2. Starts a new goroutine for each module to run `Run()`
3. When the parent goroutine is being closed (like by a SIGINT), the modules will be unregistered by calling `OnDestroy()` in the reverse order when they get registered.

Leaf source code directories
----------------------------

* leaf/chanrpc : RPC channel for inter-modules communication
* leaf/db : Database Utilities with [MongoDB](https://www.mongodb.org/) support
* leaf/gate : Gate module that connects to client
* leaf/go : Factory of goroutine that manageable for Leaf
* leaf/log : Logging
* leaf/network : Networking through TCP or WebSocket with a customized message encoding. There are two built-in encodings, [protobuf](https://developers.google.com/protocol-buffers) and JSON.
* leaf/recordfile : To manage game related data.
* leaf/timer : Timer
* leaf/util : Utilities

How to use Leaf
---------------

[LeafServer](https://github.com/name5566/leafserver) is a game server developped with Leaf. Let's start with it.

Download the source code of LeafServer：

```
git clone https://github.com/name5566/leafserver
```

Download and install leafserver to GOPATH:

```
go get github.com/name5566/leaf
```

Compile LeafServer：

```
go install server
```

Run `server` you will get below screen output if everything is successful.

```
2015/08/26 22:11:27 [release] Leaf 1.1.2 starting up
```

Press Ctrl + C to terminate the process, you'll see

```
2015/08/26 22:12:30 [release] Leaf closing down (signal: interrupt)
```

### Hello Leaf

Now with the acknowledge of LeafServer, we come to see how server receives and handles messages.


Firstly we define a JSON-encoded message(likely the protobuf). Open LeafServer msg/msg.go then you will see below:

```go
package msg

import (
    "github.com/name5566/leaf/network"
)

var Processor network.Processor

func init() {

}
```

Processor is the message handler. Here we use the handler of JSON, the default message encoding, and create a Hello message.

```go
package msg

import (
    "github.com/name5566/leaf/network/json"
)

// Create a JSON Processor（or protobuf if you like）
var Processor = json.NewProcessor()

func init() {
    // Register message Hello
    Processor.Register(&Hello{})
}

// One struct for one message
// Contains a string member
type Hello struct {
    Name string
}
```

Every message sent from client to server will be flown to `gate` module for routing. Just in brief, `gate` determines which message will be handled by which modules. We want to feed `game` module with `Hello` here, so open LeafServer gate/router.go and write below:

```go
package gate

import (
    "server/game"
    "server/msg"
)

func init() {
    // Route Hello to game
    // All communication are through ChanRPC including the management messages
    msg.Processor.SetRouter(&msg.Hello{}, game.ChanRPC)
}
```

It is ready to handle `Hello` message in `game` module. Open LeafServer game/internal/handler.go and write:

```go
package internal

import (
    "github.com/name5566/leaf/log"
    "github.com/name5566/leaf/gate"
    "reflect"
    "server/msg"
)

func init() {
    // Register the handler of `Hello` message to `game` module handleHello
    handler(&msg.Hello{}, handleHello)
}

func handler(m interface{}, h interface{}) {
    skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleHello(args []interface{}) {
    // Send "Hello"
    m := args[0].(*msg.Hello)
    // The receiver
    a := args[1].(gate.Agent)

    // The content of the message
    log.Debug("hello %v", m.Name)

    // Reply with a `Hello`
    a.WriteMsg(&msg.Hello{
        Name: "client",
    })
}
```

By here we've finished a simplest example for server. Now we will write a client from scratch for testing to understand the message structure better.

When we choose TCP over the others, the message in transition will be all formated like below:

```
--------------
| len | data |
--------------
```

To be more specific:

1. len means the size of data by bytes. len itself takes space(2 bytes by default, configurable). The minimum size of len equals the the minimum size of a single message.
2. data part is encoded with JSON or protobuf (or a customized one)

Write the client:
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

    // Hello message (JSON-encoded)
    // The structure of the message
    data := []byte(`{
        "Hello": {
            "Name": "leaf"
        }
    }`)

    // len + data
    m := make([]byte, 2+len(data))

    // BigEndian encoded
    binary.BigEndian.PutUint16(m, uint16(len(data)))

    copy(m[2:], data)

    // Send message
    conn.Write(m)
}
```

Run the client to send the message, then server will display it as received

```
2015/09/25 07:41:03 [debug  ] hello leaf
2015/09/25 07:41:03 [debug  ] read message: read tcp 127.0.0.1:3563->127.0.0.1:54599: wsarecv: An existing connection was forcibly closed by the remote host.
```

Client will exit after send out the message, and then disconnect with server. Thus server displays the event message of disconnection(the second, the event message might be dependant on the version of Go environment).

Beside TCP, WebSocket is another choice of protocol and ideal for HTML5 web game. Leaf uses TCP or WebSocket separately or jointly. In other words, server can handle TCP messages and WebSocket messages at the same time. They are "transparent" for developers. From now on we will demonstrate how to use a client based on WebSocket:
```html
<script type="text/javascript">
var ws = new WebSocket('ws://127.0.0.1:3653')

ws.onopen = function() {
    // Send Hello message
    ws.send(JSON.stringify({Hello: {
        Name: 'leaf'
    }}))
}
</script>
```

Save above to a HTML file and open it in a browser (with WebSocket support). Before that, we still have to update the configuration for LeafServer in bin/conf/server.json by adding WebSocket listenning address：
```json
{
    "LogLevel": "debug",
    "LogPath": "",
    "TCPAddr": "127.0.0.1:3563",
    "WSAddr": "127.0.0.1:3653",
    "MaxConnNum": 20000
}
```

Restart server then we get the first WebSocket message:

```
2015/09/25 07:50:03 [debug  ] hello leaf
```

Please to be noted: Within WebSocket, Leaf always send binary messages rather text messages.

### Leaf in details

LeafServer includes three modules, they are:

* gate module: for management of connection
* login module: for the user authentication
* game module: for the main business

The structure of a Leaf module is suggested (but not forced) to:

1. Be located in a separate directory
2. Have its internal implementation located under `./internal`
3. Have a file external.go to expose its interfaces. For instance of external.go:

```go
package game

import (
    "server/game/internal"
)

var (
    // Instantiate game module
    Module = new(internal.Module)
    // Expose ChanRPC
    ChanRPC = internal.ChanRPC
)
```

Instantiation of game module must be done before its registration to Leaf framework(detailed in LeafServer main.go). Besides ChanRPC needs to be exposed for inter-module communication.

Enter into game module's internal（LeafServer game/internal/module.go）：

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

skeleton is the key which implements `Run()` and provides:

* ChanRPC
* goroutine
* Timer

### Leaf ChanRPC

Since in Leaf, every module runs in a separate goroutine, a RPC channel is needed to support the communication between modules. The representing object ChanRPC needs to be registered when the game server is being started and actually it is not safe. For example, in LeafServer, game module registers two ChanRPC objects: NewAgent and CloseAgent.

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

skeleton is used to register ChanRPC. RegisterChanRPC's first parameter is the string name of ChanRPC and the second is the function that implements ChanRPC. NewAgent and CloseAgent will be called by gate module respectively when connection is set up or broken. The calling of ChanRPC includes 3 modes:

1. Synchronous mode : called waiting for ChanRPC is yielded
2. Asynchronous mode : called with a callback function where you can handle the returned ChanRPC
3. "Go mode" : return immediately ignoring any return values or errors

This is how gate module call game module's NewAgent ChanRPC (This snippet is simplified for demonstration):

```go
game.ChanRPC.Go("NewAgent", a)
```

Here NewAgent will be called with a parameter a which can be retrieved from args[0], the rest can be done in the same manner.

More references are at [leaf/chanrpc](https://github.com/name5566/leaf/blob/master/chanrpc). Please be noted, no matter how delicate the encapsulation is, calling function across goroutines cannot be that straight. Try not to create too many modules and interactions. Modules designed in Leaf are supposed to decouple the businesses from others rather make most use of CPU cores. The correct way to make most use of CPU cores is to use goroutine properly.

### Leaf Go

Use goroutine properly can make better use of CPU cores. Leaf implements its own Go() for below reasons:

* Errors within goroutine can be handled
* Game server needs to wait for all goroutines' execution
* The results of goroutine can be obtained more easily
* goroutine will follow the order to be exercised. It is very important in some occasion

Here is an example which can be tested in OnInit() in LeafServer's module.

```go
log.Debug("1")

// Define res to make the result watchable
var res string

skeleton.Go(func() {
    // Simulate a slow operation
    time.Sleep(1 * time.Second)

    // res is modified
    res = "3"
}, func() {
    log.Debug(res)
})

log.Debug("2")
```

The result are:

```go
2015/08/27 20:37:17 [debug  ] 1
2015/08/27 20:37:17 [debug  ] 2
2015/08/27 20:37:18 [debug  ] 3
```

skeleton.Go() accepts two function parameters, first one will be exercised in a separate goroutine and afterwards the second be exercised within the same goroutine. And res can only be used by one goroutine at one moment so nothing more need to be done for synchronization. This implementation makes CPU can be fully used while no need to block goroutines. It is quite convenient when shared resources are used.

More references are at [leaf/go](https://github.com/name5566/leaf/blob/master/go)。

### Leaf timer

Go has a built-in implementation in its standard library:

```go
func AfterFunc(d Duration, f func()) *Timer
```

AfterFunc() will wait for a duration of d then exercises f() in a separate goroutine. Leaf also implement AfterFunc(), and in this version f() will be exercised but within the same goroutine. It will prevent synchronization from happening.

```go
skeleton.AfterFunc(5 * time.Second, func() {
    // ...
})
```

Besides, Leaf timer support [cron expressions](https://en.wikipedia.org/wiki/Cron) to support scheduled jobs like start at 9am daily or Sunday 6pm weekly.

More references are at [leaf/timer](https://github.com/name5566/leaf/blob/master/timer)。

### Leaf log

Leaf support below log level:

1. Debug level: Not critical
2. Release level: Critical
3. Error level: Errors
4. Fatal level: Fatal errors

Debug < Release < Error < Fatal (In priority level)

For LeafServer, bin/conf/server.json is used to configure log level which will filter out the lower level log information. Fatal level log is sort of different and comes only when the game server exit. Usually it records the information when the game server is failed to start up.

Set LogFlag (LeafServer conf/conf.go) to output the file name and the line number:

```
LogFlag = log.Lshortfile
```

LogFlag：[https://golang.org/pkg/log/#pkg-constants](https://golang.org/pkg/log/#pkg-constants)


More references are at [leaf/log](https://github.com/name5566/leaf/blob/master/log).

### Leaf recordfile

Leaf recordfile is formatted in CSV([Example](https://github.com/name5566/leaf/blob/master/recordfile/test.txt)). recordfile is to manage the configuration for game. The usage of recordfile in LeafServer is quite simple:

1. Create a CSV file under bin/gamedata
2. Call readRf() to read it in gamedata module

Samples:

```go
// Make sure Test.txt is located in bin/gamedata
// The file name must match the name of the struct, and all characters are case sensitive.
// Every instance of defined struct maps to one specific row in recordfile
type Test struct {
    // The type of first column is int
    // "index" means this column will be indexed(exclusively)
    Id  int "index"
    // The type of second column is an array of int with a length of 4
    Arr [4]int
    // The type of third column is string
    Str string
}

// Load recordfile Test.txt into memory
// RfTest is the object that represents Test.txt in memory
var RfTest = readRf(Test{})

func init() {
    // Search in index
    // Fetch the row with id equals 1 in Test.txt
    r := RfTest.Index(1)

    if r != nil {
        row := r.(*Test)

        // Log this row
        log.Debug("%v %v %v", row.Id, row.Arr, row.Str)
    }
}
```

Refer to [leaf/recordfile](https://github.com/name5566/leaf/blob/master/recordfile) for more details.

Learn more
----------

More references are at Wiki [https://github.com/name5566/leaf/wiki](https://github.com/name5566/leaf/wiki)
