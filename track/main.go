/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.13 (lee8oi)

The main track file contains the key Tracking functions that actively retrieve
data from game server and perform tracking functions.
*/

/*
track is used for tracking player stats & chat messages. Utilizes gorcon
to monitor game server activity. Also has a snapshot system which stores a copy of
the current player list as JSON in the 'snapshot.json' file.
*/
package track

import (
	"fmt"
	"github.com/lee8oi/gorcon"
	"time"
)

/*
Tracker monitors player stats & chat messages via single gorcon.Rcon connection.
Runs in iterations. Sleeps for the specified wait time at the end of each iteration.
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (see time.ParseDuration doc).
*/
func Tracker(r *gorcon.Rcon, wait string) {
	var (
		pl playerList
		c  chat
	)
	dur, err := time.ParseDuration(wait)
	if err != nil {
		fmt.Println(err)
		return
	}
	pl.load("snapshot.json")
	for {
		cstr, err := r.Send("bf2cc clientchatbuffer")
		if err != nil {
			fmt.Println(err)
			break
		}
		c.new(cstr)
		c.parse()
		pstr, err := r.Send("bf2cc pl")
		if err != nil {
			fmt.Println(err)
			break
		}
		list := pl.new(pstr)
		pl.track(&list)
		pl.updateAll(&list)
		pl.snapshot("snapshot.json")
		time.Sleep(dur)
	}
}
