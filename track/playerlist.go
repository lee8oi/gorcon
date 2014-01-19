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

	//temporary values
	DamageAssists, PassAssists, CpAssists, CpCaptures, CpDefends, Neutralizes, NeutralizesAssists string
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
				Pid:           id,
				Name:          splitLine[1],
				Profileid:     splitLine[10],
				Team:          splitLine[2],
				Level:         splitLine[39],
				Kit:           kit,
				Score:         splitLine[37],
				Kills:         splitLine[31],
				Deaths:        splitLine[36],
				Alive:         splitLine[8],
				Connected:     splitLine[4],
				Vip:           splitLine[46],
				Nucleus:       splitLine[47],
				Ping:          splitLine[3],
				Idle:          idle,
				Suicides:      strings.TrimSpace(splitLine[30]),
				CpCaptures:    splitLine[25],
				CpDefends:     splitLine[26],
				DamageAssists: splitLine[19],
				Neutralizes:   splitLine[28],

				//temporary variables
				PassAssists:        splitLine[20],
				CpAssists:          splitLine[27],
				NeutralizesAssists: splitLine[29],
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
	for i := 0; i < 16; i++ {
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
	"interrupted" - Player connection was interrupted before completion.
	"established" - Connection is connected & active.
	"disconnected" - Player has disconnected from the game server.
*/
func (pl *playerList) state(key int, p *player) {
	switch {
	case pl[key].Connected == "" && p.Connected == "0":
		p.Connection = "initial"
	case pl[key].Connected == "0" && p.Connected == "0":
		pl[key].Connection = "connecting"
	case pl[key].Connected == "0" && p.Connected == "":
		pl[key].Connection = "interrupted"
	case pl[key].Connected == "1" && p.Connected == "":
		pl[key].Connection = "disconnected"
	case pl[key].Connected == "0" && p.Connected == "1":
		if pl[key].Joined == *new(time.Time) {
			pl[key].Joined = time.Now()
		}
		pl[key].Connection = "connected"
	case pl[key].Connected == "1" && p.Connected == "1":
		if pl[key].Joined == *new(time.Time) {
			pl[key].Joined = time.Now()
		}
		pl[key].Connection = "established"
	}
}

/*
status sets the current player status(s) based on stat changes.

Player status's are:
	"assisted"- assisted a kill
	"stopped" - is now idle.
	"resumed" - is no longer idle.
	"killed"  - has killed someone.
	"died"    - has died.
	"suicided"- has commit suicide.
	"promoted"- is now a vip.
	"demoted" - is no longer a vip.
	"leveled" - has leveled up.
	"neutralized" - neutralized a control point
	"defended"    - defended a control point
	"captured"	  - captured a control point
*/
func (pl *playerList) status(key int, p *player) {
	if len(p.Name) == 0 || len(pl[key].Name) == 0 {
		return
	}
	switch {
	case pl[key].Vip != p.Vip:
		if p.Vip == "1" {
			p.Status = append(p.Status, "promoted")
		} else {
			p.Status = append(p.Status, "demoted")
		}
	case p.Level > pl[key].Level && pl[key].Level != "-1":
		p.Status = append(p.Status, "leveled")
	case pl[key].Idle == 0 && p.Idle > 0:
		p.Status = append(p.Status, "stopped")
	case pl[key].Idle > 0 && p.Idle == 0:
		p.Status = append(p.Status, "resumed")
	case p.Kills > pl[key].Kills:
		if p.DamageAssists > pl[key].DamageAssists {
			p.Status = append(p.Status, "assisted")
		} else {
			p.Status = append(p.Status, "killed")
		}
	case p.Deaths > pl[key].Deaths:
		p.Status = append(p.Status, "died")
	case p.Suicides > pl[key].Suicides:
		p.Status = append(p.Status, "suicided")
	case pl[key].Neutralizes != p.Neutralizes:
		p.Status = append(p.Status, "neutralized")
	case p.CpCaptures > pl[key].CpCaptures:
		p.Status = append(p.Status, "captured")
	case p.CpDefends > pl[key].CpDefends:
		p.Status = append(p.Status, "defended")

	}
	if pl[key].PassAssists != p.PassAssists {
		fmt.Printf("%s PassAssists change %s to %s\n", p.Name, pl[key].PassAssists, p.PassAssists)
	}
	if pl[key].CpAssists != p.CpAssists {
		fmt.Printf("%s CpAssists change %s to %s\n", p.Name, pl[key].CpAssists, p.CpAssists)
	}
	if pl[key].NeutralizesAssists != p.NeutralizesAssists {
		fmt.Printf("%s NeutralizesAssists change %s to %s\n", p.Name, pl[key].NeutralizesAssists, p.NeutralizesAssists)
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
		pl[key].DamageAssists = p.DamageAssists
		pl[key].CpDefends = p.CpDefends
		pl[key].CpCaptures = p.CpCaptures
		pl[key].Neutralizes = p.Neutralizes
		//temporary

		pl[key].PassAssists = p.PassAssists
		pl[key].CpAssists = p.CpAssists
		pl[key].NeutralizesAssists = p.NeutralizesAssists
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

type crime struct {
	killers, assistants, victims, suicides []*player
}

func (pl *playerList) investigate() {
	var c crime
	for key := range pl {
		for _, value := range pl[key].Status {
			switch value {
			case "assisted":
				c.assistants = append(c.assistants, &pl[key])
			case "killed":
				c.killers = append(c.killers, &pl[key])
			case "died":
				c.victims = append(c.victims, &pl[key])
			case "suicided":
				c.suicides = append(c.suicides, &pl[key])
			}
		}
	}
	victims := len(c.victims)
	killers := len(c.killers)
	assistants := len(c.assistants)
	suicides := len(c.suicides)
	if killers > 0 && victims > 0 {
		switch {
		case killers == 1 && victims == 1:
			fmt.Printf("%s has killed %s! ", c.killers[0].Name, c.victims[0].Name)
			if assistants > 0 {
				fmt.Printf("Assisted by ")
				for i := range c.assistants {
					fmt.Printf("%s ", c.assistants[i].Name)
				}
			}
			fmt.Printf("\n")
		case killers == 1 && victims > 1:
			fmt.Printf("%s scored %d kills: ", c.killers[0].Name, victims)
			for i := range c.victims {
				fmt.Printf("%s ", c.victims[i].Name)
			}
			fmt.Printf("\n")
		case c.killers[0].Name == c.victims[0].Name && c.killers[1].Name == c.victims[1].Name:
			fmt.Printf("%s and %s killed each other.\n", c.killers[0].Name, c.killers[1].Name)
		case killers > 1 && victims > 1:
			fmt.Printf("%d killers with %d victims!\n", killers, victims)
			fmt.Printf("killers")
			for i := range c.killers {
				fmt.Printf(" %s", c.killers[i].Name)
			}
			fmt.Printf(" killed")
			for i := range c.victims {
				fmt.Printf(" %s", c.victims[i].Name)
			}
			fmt.Printf(".")
			if assistants > 0 {
				fmt.Printf("Assisted by ")
				for i := range c.assistants {
					fmt.Printf("%s ", c.assistants[i].Name)
				}
			}
			fmt.Printf("\n")
		case suicides == 1:
			fmt.Printf("%s killed himself", c.suicides[0])
		case suicides > 1:
			fmt.Printf("%d fools killed themselves: ", suicides)
			for i := range c.suicides {
				fmt.Printf("%s", c.suicides[i].Name)
			}
			fmt.Printf("\n")
		}
	}
}
