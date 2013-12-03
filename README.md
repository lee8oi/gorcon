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

Console with reconnection & Config usage.

	package main
	
	import (
		"bufio"
		"fmt"
		"github.com/lee8oi/gorcon"
		"os"
		"time"
	)
	
	var config = gorcon.Config{
		Admin:    "Gorcon",
		Address:  "123.123.123.123",
		Pass:     "SeCrEtPaSsWoRd",
		RconPort: "18666",
	}
	
	func main() {
		var r gorcon.Rcon
		r.Start(&config)
		str, err := r.SetAdmin(config.Admin)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(str)
		for {
			fmt.Printf("Command: ")
			in := bufio.NewReader(os.Stdin)
			line, err := in.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(line) > 1 {
				str, err := r.Send(line)
				if err != nil {
					fmt.Println(err)
					err = r.Reconnect(30 * time.Second)
					if err != nil {
						fmt.Println(err)
						return
					}
					str, _ = r.Send(line)
				}
				fmt.Println(str)
			}
		}
	}
	
	
	