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
	Kills, Deaths, Alive, Idle, Connected, Vip, Nucleus,
	Ping, Suicides, Connection string
	Status []string
	Pid    int
	Joined time.Time

	//temporary values
	DamageAssists, PassAssists, CpAssists, CpCaptures, CpDefends, Neutralizes, NeutralizesAssists string
}

func (p *player) playtime() string {
	return strings.Split(time.Since(p.Joined).String(), ".")[0] + "s"
}

func (p *player) team() (team string) {
	if p.Team == "2" {
		team = "Royal"
	} else if p.Team == "1" {
		team = "National"
	} else {
		team = p.Team
	}
	return
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
			kit := ""
			ksplit := strings.Split(splitLine[34], "_")
			if len(ksplit) > 1 {
				kit = ksplit[1]
			} else {
				kit = splitLine[34]
			}
			id, _ := strconv.Atoi(splitLine[0])
			//idle, _ := strconv.Atoi(splitLine[41])
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
				Idle:          splitLine[41],
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
parse parses 'bf2cc pl' data string and uses it to update player data and track
connection states & status's
*/
func (pl *playerList) parse(str string) {
	list := pl.new(str)
	for i := 0; i < 16; i++ {
		//pl.status(i, &list[i])
		pl.state(i, &list[i])
		pl.update(i, &list[i])
	}
}

/*
state sets player connection state.

Connection states are:
	"initial" - Initial player connection.
	"connecting" - Currently loading/connecting to game server.
	"connected" - Player successfully connected to the game server.
	"reconnecting" - Player was connected and is connecting again (map change?).
	"interrupted" - Player connection was interrupted before completion.
	"established" - Connection is connected & active.
	"disconnected" - Player has disconnected from the game server.
*/
func (pl *playerList) state(key int, p *player) {
	var (
		base time.Time
		s    string
	)
	switch {
	case pl[key].Connected == "" && p.Connected == "0":
		s = "initial"
		//fmt.Sprintf("CONNECTING: %s\n", p.Name)
		Log(fmt.Sprintf("CONNECTING: %s\n", p.Name))
	case pl[key].Connected == "0" && p.Connected == "0":
		s = "connecting"
	case pl[key].Connected == "0" && p.Connected == "":
		s = "interrupted"
		//fmt.Printf("INTERRUPTED: %s\n", pl[key].Name)
		Log(fmt.Sprintf("INTERRUPTED: %s\n", pl[key].Name))
	case pl[key].Connected == "0" && p.Connected == "1":
		s = "connected"
		//fmt.Printf("CONNECTED: %s\n", p.Name)
		Log(fmt.Sprintf("CONNECTED: %s\n", p.Name))
	case pl[key].Connected == "1" && p.Connected == "1":
		s = "established"
	case pl[key].Connected == "1" && p.Connected == "":
		s = "disconnected"
		if pl[key].Joined.Equal(base) {
			//fmt.Printf("DISCONNECTED: %s\n", pl[key].Name)
			Log(fmt.Sprintf("DISCONNECTED: %s\n", pl[key].Name))
		} else {
			//fmt.Printf("DISCONNECTED: %s (%s)\n", pl[key].Name, pl[key].playtime())
			Log(fmt.Sprintf("DISCONNECTED: %s (%s)\n", pl[key].Name, pl[key].playtime()))
		}
	case pl[key].Connected == "1" && p.Connected == "0":
		s = "reconnecting"
	}
	if p.Connected == "1" && pl[key].Joined.Equal(base) {
		pl[key].Joined = time.Now()
	}
	pl[key].Connection = s
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
	if pl[key].Vip != p.Vip {
		if p.Vip == "1" {
			p.Status = append(p.Status, "promoted")
			//fmt.Printf("%s has been promoted to vip\n", p.Name)
			Log(fmt.Sprintf("%s has been promoted to vip\n", p.Name))
		} else {
			p.Status = append(p.Status, "demoted")
			//fmt.Printf("%s has lost vip status\n", p.Name)
			Log(fmt.Sprintf("%s has lost vip status\n", p.Name))
		}
	}
	if p.Level > pl[key].Level && pl[key].Level != "-1" {
		p.Status = append(p.Status, "leveled")
		//fmt.Printf("%s has leveled up!\n", p.Name)
		Log(fmt.Sprintf("%s has leveled up!\n", p.Name))
	}
	idle, _ := strconv.Atoi(pl[key].Idle)
	if idle == 0 && idle > 0 {
		p.Status = append(p.Status, "stopped")
		//fmt.Printf("%s is idle!\n", p.Name)
	}
	if idle > 0 && idle == 0 {
		p.Status = append(p.Status, "resumed")
		//fmt.Printf("%s is no longer idle.\n", p.Name)
	}
	if p.Kills > pl[key].Kills {
		if p.DamageAssists > pl[key].DamageAssists {
			p.Status = append(p.Status, "assisted")
			//fmt.Printf("%s assisted a kill!\n", p.Name)
		} else {
			p.Status = append(p.Status, "killed")
			//fmt.Printf("%s scored a kill!\n", p.Name)
		}
	}
	if p.Deaths > pl[key].Deaths {
		p.Status = append(p.Status, "died")
		//fmt.Printf("%s has died.\n", p.Name)
	}
	if p.Suicides > pl[key].Suicides {
		p.Status = append(p.Status, "suicided")
		//fmt.Printf("%s has suicided.\n", p.Name)
	}
	if pl[key].Neutralizes != p.Neutralizes {
		p.Status = append(p.Status, "neutralized")
		//fmt.Printf("%s neturalized a control point!\n", p.Name)
		Log(fmt.Sprintf("%s neturalized a control point!\n", p.Name))
	}
	if p.CpCaptures > pl[key].CpCaptures {
		p.Status = append(p.Status, "captured")
		//fmt.Printf("%s captured a control point!\n", p.Name)
		Log(fmt.Sprintf("%s captured a control point!\n", p.Name))
	}
	if p.CpDefends > pl[key].CpDefends {
		p.Status = append(p.Status, "defended")
		//fmt.Printf("%s defended a control point!\n", p.Name)
		Log(fmt.Sprintf("%s defended a control point!\n", p.Name))
	}
	//temporary/testing
	if pl[key].PassAssists != p.PassAssists {
		//fmt.Printf("%s PassAssists change %s to %s\n", p.Name, pl[key].PassAssists, p.PassAssists)
		Log(fmt.Sprintf("%s PassAssists change %s to %s\n", p.Name, pl[key].PassAssists, p.PassAssists))
	}
	if pl[key].CpAssists != p.CpAssists {
		//fmt.Printf("%s CpAssists change %s to %s\n", p.Name, pl[key].CpAssists, p.CpAssists)
		Log(fmt.Sprintf("%s CpAssists change %s to %s\n", p.Name, pl[key].CpAssists, p.CpAssists))
	}
	if pl[key].NeutralizesAssists != p.NeutralizesAssists {
		//fmt.Printf("%s NeutralizesAssists change %s to %s\n", p.Name, pl[key].NeutralizesAssists, p.NeutralizesAssists)
		Log(fmt.Sprintf("%s NeutralizesAssists change %s to %s\n", p.Name, pl[key].NeutralizesAssists, p.NeutralizesAssists))
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
		pl[key].Kit = p.Kit
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

func (pl *playerList) find(term string) []int {
	var results []int
	for key := range pl {
		if strings.Contains(strings.ToLower(pl[key].Name), strings.ToLower(term)) {
			results = append(results, key)
		}
	}
	return results
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
			msg := fmt.Sprintf("- %s KILLED %s", c.killers[0].Name, c.victims[0].Name)
			if assistants > 0 {
				msg += fmt.Sprintf(" - Assisted by")
				for i := range c.assistants {
					msg += fmt.Sprintf(" %s", c.assistants[i].Name)
				}
			}
			msg += fmt.Sprintf(".\n")
			Log(msg)
		case killers == 1 && victims > 1:
			msg := fmt.Sprintf("- %s - SCORED %d KILLS - VICTIMS:", c.killers[0].Name, victims)
			for i := range c.victims {
				msg += fmt.Sprintf(" %s", c.victims[i].Name)
			}
			msg += fmt.Sprintf("\n")
			Log(msg)
		//case c.killers[0].Name == c.victims[0].Name && c.killers[1].Name == c.victims[1].Name:
		//	fmt.Printf("- %s and %s - KILLED EACH OTHER.\n", c.killers[0].Name, c.killers[1].Name)
		case killers > 1 && victims > 1:
			msg := fmt.Sprintf("- %d KILLERS %d VICTIMS!\n", killers, victims)
			msg += fmt.Sprintf("KILLERS")
			for i := range c.killers {
				msg += fmt.Sprintf(" %s", c.killers[i].Name)
			}
			msg += fmt.Sprintf(" VICTIMS")
			for i := range c.victims {
				msg += fmt.Sprintf(" %s", c.victims[i].Name)
			}
			msg += fmt.Sprintf(".")
			if assistants > 0 {
				msg += fmt.Sprintf(" ASSISTED BY ")
				for i := range c.assistants {
					msg += fmt.Sprintf("%s ", c.assistants[i].Name)
				}
			}
			msg += fmt.Sprintf("\n")
			Log(msg)
		case suicides == 1:
			Log(fmt.Sprintf("- %s - KILLED SELF", c.suicides[0]))
		case suicides > 1:
			msg := fmt.Sprintf("- %d HEROES KILLED SELVES:", suicides)
			for i := range c.suicides {
				msg += fmt.Sprintf(" %s", c.suicides[i].Name)
			}
			msg += fmt.Sprintf("\n")
			Log(msg)
		}
	}
}
