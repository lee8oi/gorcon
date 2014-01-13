/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.13 (lee8oi)

track package contains the PlayerList types and the Tracker method
needed to track player connections & game stats. Includes a snapshot
system used to store the current PlayerList in a file as JSON (./snapshot.json).

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

type Player struct {
	Pid, Name, Profileid, Team, Level, Kit, Score,
	Kills, Deaths, Alive, Connected, Vip, Nucleus,
	Ping, Suicides string
	Joined time.Time
}

//PlayerList contains a maximum of 16 player 'slots' as per game server limits.
type PlayerList [16]Player

//new takes a 'bf2cc pl' result string and returns a new PlayerList.
func (pl *PlayerList) new(data string) (plist PlayerList) {
	if len(data) > 1 {
		split := strings.Split(data, "\r")
		for _, value := range split {
			var p Player
			splitLine := strings.Split(strings.TrimSpace(value), "\t")
			if len(splitLine) < 48 {
				continue
			}
			kit := "none"
			if splitLine[34] != "none" {
				kit = strings.Split(splitLine[34], "_")[1]
			}
			p = Player{
				Pid:       splitLine[0],
				Name:      splitLine[1],
				Profileid: splitLine[10],
				Team:      splitLine[2],
				Level:     splitLine[39],
				Kit:       kit,
				Score:     splitLine[37],
				Kills:     splitLine[31],
				Deaths:    splitLine[36],
				Alive:     splitLine[8],
				Connected: splitLine[4],
				Vip:       splitLine[46],
				Nucleus:   splitLine[47],
				Ping:      splitLine[3],
				Suicides:  strings.TrimSpace(splitLine[30]),
			}
			key, _ := strconv.Atoi(p.Pid)
			plist[key] = p
		}
		return
	}
	return
}

//snapshot saves a copy of the current PlayerList to file as JSON.
func (pl *PlayerList) snapshot(path string) {
	b, err := json.Marshal(*pl)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

//load reads a snapshot file & updates current PlayerList
func (pl *PlayerList) load(path string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(b, pl)
	if err != nil {
		fmt.Println(err)
	}
}

//Tracker uses an Rcon connection to monitor player connection changes and keep
//the current PlayerList updated. Uses 'bf2cc pl' rcon command to request player data.
func (pl *PlayerList) Tracker(r *gorcon.Rcon) {
	pl.load("snapshot.json")
	for {
		str, err := r.Send("bf2cc pl")
		if err != nil {
			fmt.Println("main 36 error: ", err)
			break
		}
		list := pl.new(str)
		pl.track(&list)
		pl.updateAll(&list)
		pl.snapshot("snapshot.json")
		time.Sleep(1 * time.Second)
	}
}

//track compares current PlayerList to new list, slot by slot, to track player connection changes.
func (pl *PlayerList) track(list *PlayerList) {
	var base time.Time
	for i := 0; i < 16; i++ {
		switch {
		case pl[i].Name == list[i].Name: //connecting existing
			if pl[i].Connected == "0" && list[i].Connected == "1" {
				if pl[i].Joined == base {
					fmt.Printf("%s: connected\n", list[i].Name)
					t := time.Now()
					pl[i].Joined = t
				}
			}
		case len(pl[i].Name) == 0 && len(list[i].Name) > 0: //connecting
			pl.update(i, &list[i])
			if pl[i].Connected == "1" && pl[i].Joined == base {
				pl[i].Joined = time.Now()
				fmt.Printf("%s - connection exists\n", list[i].Name)
			}
			if list[i].Connected == "0" {
				fmt.Printf("%s - connecting\n", list[i].Name)
			}
		case len(pl[i].Name) > 0 && len(list[i].Name) == 0: //disconnecting
			if pl[i].Joined != base {
				dur := time.Since(pl[i].Joined)
				fmt.Printf("%s - disconnected (playtime: %s)\n", pl[i].Name, dur.String())
			} else {
				fmt.Printf("%s - disconnected (interrupted)\n", pl[i].Name)
			}

		}
	}
}

//update the player slot at the index specifed by key.
//After initial update: only elements that can potentially change during playtime
//are updated.
func (pl *PlayerList) update(key int, p *Player) {
	if len(p.Name) > 0 && pl[key].Name == p.Name {
		pl[key].Alive = p.Alive
		pl[key].Connected = p.Connected
		pl[key].Deaths = p.Deaths
		pl[key].Kills = p.Kills
		pl[key].Level = p.Level
		pl[key].Ping = p.Ping
		pl[key].Score = p.Score
		pl[key].Suicides = p.Suicides
		pl[key].Vip = p.Vip
		return
	}
	pl[key] = *p
}

//updateall parses a given PlayerList and updates all existing player slots.
func (pl *PlayerList) updateAll(l *PlayerList) {
	var base Player
	for i := 0; i < 16; i++ {
		if pl[i] == base && l[i] == base { //skip if current & new are empty
			continue
		}
		pl.update(i, &l[i])
	}
}
