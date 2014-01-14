gorcon/track - Experimental
======

track is used for tracking player stats & chat messages. Utilizes gorcon
to monitor game server activity. Also has a snapshot system which stores a copy of
the current player list as JSON in the 'snapshot.json' file. Snapshots are used
to recover playerlist data (including session playtime) in the event of
application interruption/etc.

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
		track.Tracker(&r, "500ms")
	}