gorcon
======

gorcon package contains the essential functions needed for, connecting to &
running commands on, BF2CC based Rcon servers.

EXAMPLES:

Basic:

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
		if err := r.Login(PASSWORD); err != nil {
			fmt.Println(err)
			return
		}
		result, err := r.Send("RCON COMMAND")
		if err != nil {
			fmt.Println(err)
		}
		if len(result) > 1 {
			fmt.Println(result)
		}
	}

Command console including Reconnect & Config:

	package main
	
	import (
		"bufio"
		"fmt"
		"github.com/lee8oi/gorcon"
		"os"
	)
	
	var config = gorcon.Config{
		Admin:    "Gorcon",
		Address:  "123.123.123.123",
		Pass:     "SeCrEtPaSsWoRd",
		RconPort: "18666",
	}
	
	func main() {
		var r gorcon.Rcon
		if err := r.Connect(config.Address + ":" + config.Port); err != nil {
			fmt.Println(err)
			return
		}
		if err := r.Login(config.Pass); err != nil {
			fmt.Println(err)
			return
		}
		r.AutoReconnect(true)
		go r.Writer()
		go r.Reader()
		for {
			in := bufio.NewReader(os.Stdin)
			line, err := in.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(line) > 1 {
				r.Write(line)
			}
		}
	}