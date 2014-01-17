/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.16 (lee8oi)

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
	"encoding/json"
	"fmt"
	"github.com/lee8oi/gorcon"
	"io/ioutil"
	"strings"
	"time"
)

type Tracker struct {
	players playerList
	aliases map[string]alias
	admins  map[string]admin
	Rcon    gorcon.Rcon
}

type admin struct {
	Power int
	Name  string
}

type alias struct {
	Power   int
	Command string
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
	loadJSON("snapshot.json", &t.players)
	if err := loadJSON("admins.json", &t.admins); err != nil {
		t.admins = make(map[string]admin)
		t.admins["2318009192"] = admin{Power: 100, Name: "Vegabruda"}
		writeJSON("admins.json", &t.admins)
	}
	if err := loadJSON("aliases.json", &t.aliases); err != nil {
		t.aliases = make(map[string]alias)
		t.aliases["say"] = alias{Power: 100, Command: "bf2cc sendserverchat"}
		t.aliases["test"] = alias{Power: 100, Command: "bf2cc sendserverchat testing successful"}
		writeJSON("aliases.json", &t.aliases)
	}
	for {
		str, err := t.Rcon.Send("bf2cc clientchatbuffer")
		if err != nil {
			fmt.Println(err)
			break
		}
		com := make(chan *message)
		go parseChat(str, com)
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
func (t *Tracker) command(com chan *message) {
	for m := range com {
		p := t.players.search(m.Origin)
		if p == nil {
			continue
		}
		split := strings.Split(m.Text[1:], " ")
		if len(t.aliases[split[0]].Command) > 0 { //existing alias
			full := t.aliases[split[0]].Command + " " + strings.Join(split[1:], " ") + "\n"
			if t.aliases[split[0]].Power == 0 { //public alias
				fmt.Printf("Public alias: %s", full)
			} else { //admin alias
				if t.admins[p.Nucleus].Power >= t.aliases[split[0]].Power { //enough power
					fmt.Printf("As Admin: %s", full)
				} else { //not enough power
					fmt.Printf("%s - not enough power\n", p.Name)
				}
			}
		}
	}
}

func writeJSON(path string, m interface{}) {
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func loadJSON(path string, m interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = json.Unmarshal(b, m)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
