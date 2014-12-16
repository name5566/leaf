package main

import (
	"github.com/name5566/leaf"
	"github.com/name5566/leaf/demo/server/echo/impl"
	"github.com/name5566/leaf/demo/server/gate/impl"
)

func main() {
	leaf.Run(
		new(echoimpl.Module),
		new(gateimpl.Module),
	)
}
