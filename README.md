gorcon
======

Golang based Rcon package for connecting to BF2CC admin servers.

Basic usage (replace <adminname>, <address>, <port>, <pass>, and <command> if copying):

	package main
	
	import (
		"fmt"
		"github.com/lee8oi/gorcon"
	)
	
	func main() {
		var r gorcon.Rcon
		r.Admin = "<adminname>"
		if err := r.Connect("<address>:<port>"); err != nil {
			fmt.Println(err)
			return
		}
		if err := r.Login("<pass>"); err != nil {
			fmt.Println(err)
			return
		}
		data, err := r.Send("<command>")
		if err != nil {
			fmt.Println("Error", err)
		}
		fmt.Println(data)
	}

	
	
	