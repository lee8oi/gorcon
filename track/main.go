/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.15 (lee8oi)

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
	admins   group
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

	if err := t.admins.load("admins.json"); err != nil {
		t.admins.Power = 100
		t.admins.Members = make(map[string]user)
		t.admins.add("2318009192", "Vegabruda")
		t.admins.save("admins.json")
	}

	for {
		str, err := t.Rcon.Send("bf2cc clientchatbuffer")
		if err != nil {
			fmt.Println(err)
			break
		}
		t.messages.add(str)
		com := make(chan *player)
		go t.messages.parse(&t.players, com)
		t.command(com)

		str, err = t.Rcon.Send("bf2cc pl")
		if err != nil {
			fmt.Println(err)
			break
		}
		mon := make(chan player)
		go t.players.track(str, mon)
		t.monitor(mon)

		time.Sleep(dur)
	}
}

//monitor channel data from t.players.track(). Used to monitor player connection states.
func (t *Tracker) monitor(mon chan player) {
	for i := range mon {
		//fmt.Println(i.State)
		switch i.State {
		case "connected":
			fmt.Printf("%s has connected\n", i.Name)
		case "initial":
			fmt.Printf("%s is connecting\n", i.Name)
		case "disconnected":
			if i.Joined == *new(time.Time) {
				fmt.Printf("%s has disconnected\n", i.Name)
			} else {
				fmt.Printf("%s has disconnected (%s)\n", i.Name, i.playtime())
			}
		}
	}
}

//command monitors com channel for commands sent from chat.parse(). Used to handle
//in-game commands typed by players.
func (t *Tracker) command(com chan *player) {
	for i := range com {
		if t.admins.member(i.Nucleus) {
			fmt.Printf("%s is an admin! ", i.Name)
			fmt.Printf("profileid: %s nucleusid: %s\n", i.Profileid, i.Nucleus)
		}
		fmt.Println("command found: ", i.Command)
	}
}
