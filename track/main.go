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
	//"strconv"
	//"strings"
	"time"
)

type Tracker struct {
	players playerList
	aliases map[string]alias
	admins  map[string]admin
	proc    chan process
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
	t.proc = make(chan process)
	go t.processor()
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
		t.aliases["testkick"] = alias{Power: 100, Command: "kick"}
		t.aliases["testban"] = alias{Power: 100, Command: "kick"}
		if err := writeJSON("aliases.json", &t.aliases); err != nil {
			fmt.Println(err)
		}
	}
	for {
		str := t.process("reply", "bf2cc si")
		t.game.update(str)

		str = t.process("reply", "bf2cc clientchatbuffer")
		com := make(chan *message)
		go parseChat(str, com)
		t.interpret(com)

		str = t.process("reply", "bf2cc pl")
		t.players.parse(str)
		if err := writeJSON("players.json", t.players); err != nil {
			fmt.Println(err)
		}
		t.players.investigate()

		time.Sleep(dur)
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
