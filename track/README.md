gorcon/track - Experimental
======

gorcon/track package contains the PlayerList types and the Tracker methods
needed to track player connections & stats.

Notes:

Tracker currently logs connection status to standard output. Includes session
playtime on disconnect.

Snapshot system writes PlayerList to file as JSON (current dir. 'snapshot.json'). Tracker loads the snapshot prior to tracking.

Example:

Rcon with 30 second AutoReconnect and Tracker.

	package main
	
	import (
		"fmt"
		"github.com/lee8oi/gorcon"
		"github.com/lee8oi/gorcon/track"
	)
	
	var config = gorcon.Config{
		Admin:    "Gorcon",
		Address:  "123.123.123.123",
		Pass:     "SeCrEtPaSsWoRd",
		RconPort: "18666",
	}
	
	func main() {
		var (
			r  gorcon.Rcon
			pl track.PlayerList
		)
		if err := r.Connect(config.Address + ":" + config.Port); err != nil {
			fmt.Println(err)
			return
		}
		if err := r.Login(config.Pass); err != nil {
			fmt.Println(err)
			return
		}
		r.AutoReconnect("30s")
		pl.Tracker(&r)
	}