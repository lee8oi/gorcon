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
	"github.com/lee8oi/gorcon/log"
	"io/ioutil"
	//"strconv"
	"strings"
	"time"
)

var Log func(string)

type Tracker struct {
	players playerList
	aliases map[string]alias
	admins  map[string]admin
	//proc    chan process
	game game
	Rcon gorcon.Rcon
}

type admin struct {
	Power int
	Name  string
}

type alias struct {
	Power      int
	Visibility string
	Message    string
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
	Log = log.H.Log
	go t.Rcon.Init()
	go t.Rcon.Handler(t.handle)
	go log.Start()
	t.Rcon.Enqueue("bf2cc monitor 1")
	t.Rcon.Enqueue("bf2cc setadminname Gorcon")

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
		t.aliases["say"] = alias{Power: 100, Visibility: "public", Message: ""}
		t.aliases["self"] = alias{Power: 0, Visibility: "private", Message: "$PN$ $PT$ $PL$ $PTN$ enemy: $ET$"}
		t.aliases["test"] = alias{Power: 100, Visibility: "private", Message: "testing successful"}
		t.aliases["toot"] = alias{Power: 0, Visibility: "public", Message: "$PN$ bites his lip and farts out the word *$PT$*"}
		t.aliases["kick"] = alias{Power: 100, Visibility: "server", Message: "kick"}
		t.aliases["ban"] = alias{Power: 100, Visibility: "server", Message: "ban"}
		t.aliases["tacos"] = alias{Power: 0, Visibility: "public", Message: "We only use the finest cuts of $ET$ found on the battlefield. These delicious tacos are for the $PT$ by the $PT$!"}
		t.aliases["pizza"] = alias{Power: 0, Visibility: "public", Message: "Only the freshest cuts of $ET$ meat go into our fine $PT$ deep dish pizzas!"}
		t.aliases["beer"] = alias{Power: 0, Visibility: "public", Message: "$PT$ have some tasty pale ale, but the $ET$'s are using them for target practice."}
		t.aliases["bacon"] = alias{Power: 0, Visibility: "public", Message: "Thinly sliced $ET$'s make the best bacon. Try it for yourself!"}
		t.aliases["rawr"] = alias{Power: 0, Visibility: "public", Message: "$PN$ howls out a thunderous battle cry."}
		t.aliases["joint"] = alias{Power: 0, Visibility: "public", Message: "Puff puff pass some of this wicked stuff from our private stash!"}
		t.aliases["cake"] = alias{Power: 0, Visibility: "public", Message: "The $PN$'s have ordered a cake for the $ET$'s! Filled with explosives."}
		t.aliases["panic"] = alias{Power: 0, Visibility: "public", Message: "$PN$ panics and starts screaming hysterically."}
		t.aliases["rage"] = alias{Power: 0, Visibility: "public", Message: "$PN$ gets mad an starts screaming like a maniac."}
		t.aliases["meow"] = alias{Power: 0, Visibility: "public", Message: "$PN$ lets out a scrappy alley cat meow."}
		t.aliases["fart"] = alias{Power: 0, Visibility: "public", Message: "$PN$ bites lip and lets out a horrendous grass-wilting fart."}
		t.aliases["complaint"] = alias{Power: 0, Visibility: "public", Message: "$PN says *heres the complaint department* and points to the exit."}
		t.aliases["rules"] = alias{Power: 0, Visibility: "public", Message: "Rules: Be respectful. Help your team. No cheating, no whining, no badmouthing, no idling, no t-bagging. no soliciting."}
		t.aliases["promote"] = alias{Power: 100, Visibility: "server", Message: ""}
		t.aliases["demote"] = alias{Power: 100, Visibility: "server", Message: ""}
		t.aliases["info"] = alias{Power: 100, Visibility: "server", Message: "$PN$ Class:$PC$ Lvl:$PL$ Ping:$PING$ $VIP$"}
		if err := writeJSON("aliases.json", &t.aliases); err != nil {
			fmt.Println(err)
		}
	}
	for {
		t.Rcon.Enqueue("bf2cc si")
		t.Rcon.Enqueue("bf2cc pl")
		t.Rcon.Enqueue("bf2cc clientchatbuffer")
		//t.Log("testing iteration")
		time.Sleep(dur)
	}
}

func (t *Tracker) handle(s string) {
	typ := identify(&s)
	switch typ {
	case "server":
		before := t.game.Players
		t.game.update(s)
		after := t.game.Players
		if before != "0" && after == "0" { //when last player leaves
			t.players.parse(" ")
		}
	case "chat":
		com := make(chan *message)
		go parseChat(s, com)
		t.interpret(com)
	case "player":
		t.players.parse(s)
		if err := writeJSON("players.json", t.players); err != nil {
			fmt.Println(err)
		}
		//t.players.investigate()
		//case "state", "other", "viplist", "maplist":
		//	fmt.Println(t)
		//	fmt.Println(s)
	}
}

func identify(s *string) (t string) {
	first := strings.Split(*s, "\r")[0]
	split := strings.Split(strings.TrimSpace(first), "\t")
	length := len(split)
	switch {
	case length == 48:
		t = "player"
	case length == 32:
		t = "server"
	case length == 5 || length == 6:
		t = "chat"
	case len(*s) == 1:
		t = "state"
	case length == 2:
		t = "viplist"
	case length == 1:
		split := strings.Split(first, "\n")
		nsplit := strings.Split(split[0], " ")
		if len(nsplit) == 3 {
			t = "maplist"
			return
		}
		fallthrough
	default:
		t = "other"
	}
	return
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
