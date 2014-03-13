gorcon/track - Experimental
======

track is used for tracking player stats & chat messages. Utilizes gorcon
to monitor game server activity. Also has a snapshot system which stores a copy of
the current player list as JSON in the 'snapshot.json' file.

Example:

Rcon with AutoReconnect and Tracker usage.

	package main
	
	import (
		"fmt"
		"github.com/lee8oi/gorcon"
		"github.com/lee8oi/gorcon/track"
		"time"
	)
	
	var config = gorcon.Config{
		Admin:    "Gorcon",
		Address:  "123.123.123.123",
		Pass:     "SeCrEtPaSsWoRd",
		RconPort: "18666",
	}
	
	func main() {
		var t track.Tracker
		if err := t.Rcon.Connect(config.Address + ":" + config.Port); err != nil {
			fmt.Println(err)
			return
		}
		if err := t.Rcon.Login(config.Pass); err != nil {
			fmt.Println(err)
			return
		}
		t.Rcon.AutoReconnect("30s")
		t.Start("500ms")
	}