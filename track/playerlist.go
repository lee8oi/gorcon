/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.16 (lee8oi)

playerList and its methods are used to track player stats & connection
changes. Includes a snapshot system used to store the current playerList in a file
as JSON (./snapshot.json).
*/

//
package track

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type player struct {
	Pid, Name, Profileid, Team, Level, Kit, Score,
	Kills, Deaths, Alive, Connected, Vip, Nucleus,
	Ping, Suicides, State, Command string

	Joined time.Time
}

func (p *player) playtime() string {
	return strings.Split(time.Since(p.Joined).String(), ".")[0] + "s"
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

/*
track compares current playerList to new list to track player connection states.
Players are immedidately sent to mon(itor) channel for handling.

Connection states are:
	"initial" - Initial player connection.
	"connecting" - Currently loading/connecting to game server.
	"connected" - Player successfully connected to the game server.
	"established" - Connection is connected & active.
	"disconnected" - Player has disconnected from the game server.
*/
func (pl *playerList) track(str string, mon chan player) {
	list := pl.new(str)
	for i := 0; i < 16; i++ {
		if pl[i].Name == list[i].Name && len(pl[i].Name) > 0 {
			if pl[i].Connected == "0" && list[i].Connected == "1" { //connected
				if pl[i].Joined == *new(time.Time) {
					pl[i].Joined = time.Now()
					pl[i].State = "connected"
				}
			}
			if pl[i].Connected == "1" && list[i].Connected == "1" { //established
				if pl[i].Joined == *new(time.Time) {
					fmt.Printf("%s: tracker reset\n", pl[i].Name)
					pl[i].Joined = time.Now()
				}
				pl[i].State = "established"
			}
			if pl[i].Connected == "0" && list[i].Connected == "0" {
				pl[i].State = "connecting"
			}
		} else {
			if len(pl[i].Name) > 0 && len(list[i].Name) == 0 { //disconnected
				pl[i].State = "disconnected"
				mon <- pl[i]
			}
			if len(pl[i].Name) == 0 && len(list[i].Name) > 0 { //connecting
				list[i].State = "initial"
			}
		}
		pl.update(i, &list[i])
		mon <- pl[i]
	}
	close(mon)
	writeJSON("snapshot.json", pl)
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

func (pl *playerList) search(terms string) (p *player) {
	for key, value := range pl {
		if strings.Contains(value.Name, terms) {
			p = &pl[key]
			break
		}
	}
	return
}
