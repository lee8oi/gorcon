gorcon
======

gorcon package contains the essential functions needed for, connecting to &
running commands on, BF2CC based Rcon servers.

Basic usage (replace ALLCAPS elements with appropriate values):

	package main
	
	import (
		"fmt"
		"github.com/lee8oi/gorcon"
	)
	
	func main() {
		var r gorcon.Rcon
		r.Admin = "ADMINNAME"
		if err := r.Connect("ADDRESS:PORT"); err != nil {
			fmt.Println(err)
			return
		}
		if err := r.Login("PASS"); err != nil {
			fmt.Println(err)
			return
		}
		data, err := r.Send("COMMAND")
		if err != nil {
			fmt.Println("Error", err)
			return
		}
		fmt.Println(data)
	}

	
	
	