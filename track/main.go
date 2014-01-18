/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

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
	"strconv"
	"strings"
	"time"
)

type Tracker struct {
	players playerList
	aliases map[string]alias
	admins  map[string]admin
	game    game
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
	loadJSON("players.json", &t.players)
	loadJSON("game.json", &t.game)
	if err := loadJSON("admins.json", &t.admins); err != nil {
		t.admins = make(map[string]admin)
		t.admins["2318009192"] = admin{Power: 100, Name: "Vegabruda"}
		if err := writeJSON("admins.json", &t.admins); err != nil {
			fmt.Println(err)
		}
	}
	if err := loadJSON("aliases.json", &t.aliases); err != nil {
		t.aliases = make(map[string]alias)
		t.aliases["say"] = alias{Power: 100, Command: "bf2cc sendserverchat"}
		t.aliases["test"] = alias{Power: 100, Command: "bf2cc sendserverchat testing successful"}
		t.aliases["fart"] = alias{Power: 0, Command: "bf2cc sendserverchat Someone lets a smelly one go."}
		t.aliases["request"] = alias{Power: 0, Command: "bf2cc sendserverchat Nobody can hear you scream here."}
		if err := writeJSON("aliases.json", &t.aliases); err != nil {
			fmt.Println(err)
		}
	}
	for {
		str, err := t.Rcon.Send("bf2cc si")
		if err != nil {
			fmt.Println(err)
			break
		}
		t.game.update(str)

		str, err = t.Rcon.Send("bf2cc clientchatbuffer")
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
		mon := make(chan *player)
		go t.players.parse(str, mon)
		t.monitor(mon)

		time.Sleep(dur)
	}
}

//monitor channel data from t.players.track(). Used to monitor player connection states.
func (t *Tracker) monitor(mon chan *player) {
	for i := range mon {
		switch i.Connection {
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

		for _, value := range i.Status {
			switch value {
			case "assisted":
				//fmt.Printf("%s has assisted a kill.\n", i.Name)
			case "killed":
				//fmt.Printf("%s has killed someone!\n", i.Name)
			case "died":
				//fmt.Printf("%s has died!\n", i.Name)
			case "stopped":
				//fmt.Printf("%s has stopped moving.\n", i.Name)
			case "suicided":
				//fmt.Printf("%s has commit suicide!\n", i.Name)
			case "leveled":
				fmt.Printf("%s has leveled up!\n", i.Name)
			case "resumed":
				//fmt.Printf("%s is moving again.\n", i.Name)
			case "promoted":
				fmt.Printf("%s has been promoted to vip.", i.Name)
			case "demoted":
				fmt.Printf("%s has been demoted from vip.", i.Name)
			case "captured":
				fmt.Printf("%s has captured a control point!\n", i.Name)
			case "defended":
				fmt.Printf("%s has defended a control point!\n", i.Name)
			case "neutralized":
				fmt.Printf("%s has neutralized a control point!\n", i.Name)
			}
		}
	}
	if err := writeJSON("players.json", t.players); err != nil {
		fmt.Println(err)
	}
	t.players.analyze()
}

//command monitors com channel for messages sent from chat.parse(). Used to handle
//in-game commands typed by players.
func (t *Tracker) command(com chan *message) {
	for m := range com {
		id, _ := strconv.Atoi(m.Pid)
		split := strings.Split(m.Text[1:], " ")
		if len(t.aliases[split[0]].Command) > 0 { //existing alias
			full := t.aliases[split[0]].Command + " " + strings.Join(split[1:], " ") + "\n"
			if t.aliases[split[0]].Power == 0 { //public alias
				fmt.Printf("Public alias: %s", full)
			} else { //admin alias
				if t.admins[t.players[id].Nucleus].Power >= t.aliases[split[0]].Power { //enough power
					fmt.Printf("As Admin: %s", full)
				} else { //not enough power
					fmt.Printf("%s - not enough power\n", t.players[id].Name)
				}
			}
		}
	}
}

func writeJSON(path string, m interface{}) (e error) {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		e = err
		return
	}
	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		e = err
	}
	return
}

func loadJSON(path string, m interface{}) (e error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		e = err
	}
	err = json.Unmarshal(b, m)
	if err != nil {
		e = err
	}
	return
}
