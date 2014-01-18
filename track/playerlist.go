/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

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
	Name, Profileid, Team, Level, Kit, Score,
	Kills, Deaths, Alive, Connected, Vip, Nucleus,
	Ping, Suicides, Connection string
	Status    []string
	Pid, Idle int
	Joined    time.Time
}

func (p *player) playtime() string {
	return strings.Split(time.Since(p.Joined).String(), ".")[0] + "s"
}

//playerList contains a maximum of 16 player 'slots' as per game server limits.
type playerList [16]player

//empty returns true if all slots are 'empty'
func (pl *playerList) empty() bool {
	for i := 0; i < 16; i++ {
		if pl[i].Name != "" {
			return false
		}
	}
	return true
}

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
			id, _ := strconv.Atoi(splitLine[0])
			idle, _ := strconv.Atoi(splitLine[41])
			p = player{
				Pid:       id,
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
				Idle:      idle,
				Suicides:  strings.TrimSpace(splitLine[30]),
			}
			plist[id] = p
		}
		return
	}
	return
}

/*
parse parses 'bf2cc pl' data string and uses it to track connection states and status's for each
player. Player data is updated and a player pointer is sent to the monitor channel for handling.
*/
func (pl *playerList) parse(str string, mon chan *player) {
	defer close(mon)
	list := pl.new(str)
	//if list.empty() {
	//	return
	//}
	for i := 0; i < 16; i++ {
		if len(pl[i].Name) == 0 && len(list[i].Name) == 0 {
		}
		pl.status(i, &list[i])
		pl.state(i, &list[i])
		if pl[i].Connection == "disconnected" {
			mon <- &pl[i]
		}
		pl.update(i, &list[i])
		mon <- &pl[i]
	}
}

/*
state sets player connection state.

Connection states are:
	"initial" - Initial player connection.
	"connecting" - Currently loading/connecting to game server.
	"connected" - Player successfully connected to the game server.
	"established" - Connection is connected & active.
	"disconnected" - Player has disconnected from the game server.
*/
func (pl *playerList) state(key int, p *player) {
	if pl[key].Name == p.Name && len(pl[key].Name) > 0 {
		if pl[key].Connected == "0" && p.Connected == "1" { //connected
			if pl[key].Joined == *new(time.Time) {
				pl[key].Joined = time.Now()
				pl[key].Connection = "connected"
			}
		}
		if pl[key].Connected == "1" && p.Connected == "1" { //established
			if pl[key].Joined == *new(time.Time) {
				fmt.Printf("%s: tracker reset\n", pl[key].Name)
				pl[key].Joined = time.Now()
			} else {
				pl[key].Connection = "established"
			}
		}
		if pl[key].Connected == "0" && p.Connected == "0" {
			pl[key].Connection = "connecting"
		}
	} else {
		if len(pl[key].Name) > 0 && len(p.Name) == 0 { //disconnected
			pl[key].Connection = "disconnected"
		}
		if len(pl[key].Name) == 0 && len(p.Name) > 0 { //connecting
			p.Connection = "initial"
		}
	}
}

/*
status sets the current player status(s).

Player status's are:
	"stopped" - is now idle.
	"resumed" - is no longer idle.
	"killed"  - has killed someone.
	"died"    - has died.
	"suicided"- has commit suicide.
	"promoted"- is now a vip.
	"demoted" - is no longer a vip.
	"leveled" - has leveled up.
*/
func (pl *playerList) status(key int, p *player) {
	if len(p.Name) == 0 || len(pl[key].Name) == 0 {
		return
	}
	if pl[key].Vip != p.Vip {
		if p.Vip == "1" {
			p.Status = append(p.Status, "promoted")
		} else {
			p.Status = append(p.Status, "demoted")
		}
	}
	if p.Level > pl[key].Level && pl[key].Level != "-1" {
		p.Status = append(p.Status, "leveled")
	}
	if pl[key].Idle == 0 && p.Idle > 0 {
		p.Status = append(p.Status, "stopped")
	}
	if pl[key].Idle > 0 && p.Idle == 0 {
		fmt.Println(pl[key].Idle)
		p.Status = append(p.Status, "resumed")
	}
	if p.Kills > pl[key].Kills {
		p.Status = append(p.Status, "killed")
	}
	if pl[key].Alive == "1" && p.Alive == "0" {
		p.Status = append(p.Status, "died")
	}
	if p.Suicides > pl[key].Suicides {
		p.Status = append(p.Status, "suicided")
	}
}

/*
update the player slot at the index specifed by key.
After initial update: only elements that can potentially change during playtime
are updated.
*/
func (pl *playerList) update(key int, p *player) {
	if len(p.Name) > 0 && pl[key].Name == p.Name {
		pl[key].Connected = p.Connected
		pl[key].Status = p.Status
		pl[key].Alive = p.Alive
		pl[key].Vip = p.Vip
		pl[key].Deaths = p.Deaths
		pl[key].Kills = p.Kills
		pl[key].Level = p.Level
		pl[key].Ping = p.Ping
		pl[key].Score = p.Score
		pl[key].Suicides = p.Suicides
		pl[key].Idle = p.Idle
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
