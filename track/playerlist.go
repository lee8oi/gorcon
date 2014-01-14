/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.13 (lee8oi)

playerList and its methods are used to track player stats & connection
changes. Includes a snapshot system used to store the current playerList in a file
as JSON (./snapshot.json).
*/

//
package track

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type player struct {
	Pid, Name, Profileid, Team, Level, Kit, Score,
	Kills, Deaths, Alive, Connected, Vip, Nucleus,
	Ping, Suicides string
	Joined time.Time
}

//playerList contains a maximum of 16 player 'slots' as per game server limits.
type playerList [16]player

//new takes a 'bf2cc pl' result string and returns a new playerList.
func (pl *playerList) new(data string) (plist playerList) {
	if len(data) > 1 {
		split := strings.Split(data, "\r")
		for _, value := range split {
			var p player
			splitLine := strings.Split(strings.TrimSpace(value), "\t")
			if len(splitLine) < 48 {
				continue
			}
			kit := "none"
			if splitLine[34] != "none" {
				kit = strings.Split(splitLine[34], "_")[1]
			}
			p = player{
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

//snapshot saves a copy of the current playerList to file as JSON.
func (pl *playerList) snapshot(path string) {
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

//load reads a snapshot file & updates current playerList
func (pl *playerList) load(path string) {
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

//track compares current playerList to new list, slot by slot, to track player connection changes.
func (pl *playerList) track(list *playerList) {
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
func (pl *playerList) update(key int, p *player) {
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

//updateall parses a given playerList and updates all existing player slots.
func (pl *playerList) updateAll(l *playerList) {
	var base player
	for i := 0; i < 16; i++ {
		if pl[i] == base && l[i] == base { //skip if current & new are empty
			continue
		}
		pl.update(i, &l[i])
	}
}
