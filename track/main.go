/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.14 (lee8oi)

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

type Tracker struct {
	players  playerList
	messages chat
	Rcon     gorcon.Rcon
}

/*
Start runs the Tracker which monitors player stats & chat messages via Rcon connection.
Runs in iterations. Sleeps for the specified wait time at the end of each iteration.
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (see time.ParseDuration doc).
*/
func (t *Tracker) Start(wait string) {
	dur, err := time.ParseDuration(wait)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.players.load("snapshot.json")
	for {
		str, err := t.Rcon.Send("bf2cc clientchatbuffer")
		if err != nil {
			fmt.Println(err)
			break
		}
		t.messages.add(str)
		com := make(chan string)
		go t.messages.parse(com)
		for i := range com {
			fmt.Println(i)
		}
		t.messages.clear()
		str, err = t.Rcon.Send("bf2cc pl")
		if err != nil {
			fmt.Println(err)
			break
		}
		t.players.track(str)
		t.players.save("snapshot.json")
		time.Sleep(dur)
	}
}

func (t *Tracker) command(cmd string) {

}
