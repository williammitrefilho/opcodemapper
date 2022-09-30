package main

import (
	"modgo.com/server"
	"modgo.com/whatserver"
)

func main(){

	server.Listen(whatserver.Handler)
}