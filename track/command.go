/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

command methods are used to process in-game commands, system commands, etc.
*/

//
package track

import (
	"fmt"
	"strconv"
	"strings"
)

//interpret monitors com channel for messages sent from parseChat(). Used to interpret
//command lines in messages and create processes to handle them.
func (t *Tracker) interpret(com chan *message) {
	for m := range com {
		id, _ := strconv.Atoi(m.Pid)
		split := strings.Split(m.Text[1:], " ")
		fmt.Printf("%s[%s]: %s\n", m.Origin, m.Time, m.Text)
		if m.IsCommand && len(t.aliases[split[0]].Command) > 0 {
			line := ""
			permitted := false
			public := false
			if t.admins[t.players[id].Nucleus].Power >= t.aliases[split[0]].Power {
				permitted = true
			}
			if t.aliases[split[0]].Power == 0 {
				//public alias (no power check)
				public = true
			}
			if !public && !permitted {
				fmt.Printf("%s - not enough power\n", t.players[id].Name)
				continue
			}
			switch {
			case split[0] == "testkick" || split[0] == "testban":
				if len(split) > 1 {
					r := t.players.find(split[1])
					if len(r) == 1 {
						split[1] = t.players[r[0]].Name
						//t.process("send", t.aliases[split[0]].Command+" "+strings.Join(split[1:], " "))
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("Pretending to %s %s", split[0], split[1]))
						t.process("send", l)
					} else if len(r) > 1 {
						//fmt.Sprintf("multiple players found ('%s')", split[1])
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("multiple players found ('%s')", split[1]))
						t.process("send", l)
					} else {
						fmt.Printf("No results found.")
					}
				}
			default:
				if len(split) > 1 {
					line = t.parseTags(id, strings.Join(split[1:], " "))
					t.process("send", t.aliases[split[0]].Command+" "+line)
				}
			}
		}
	}
}

type process struct {
	instruct, line string
	reply          chan string
}

func (t *Tracker) process(instruct, line string) string {
	reply := make(chan string)
	t.proc <- process{instruct, line, reply}
	return <-reply
}

func (t *Tracker) processor() {
	for d := range t.proc {
		switch d.instruct {
		case "reload":
			if err := loadJSON(d.line, &t.aliases); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Aliases reloaded.")
			}
		case "save":
			if err := writeJSON(d.line, &t.aliases); err != nil {
				fmt.Println(err)
			}
		case "send":
			fmt.Printf("%s", d.line)
			str, err := t.Rcon.Send(d.line)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(str)
			d.reply <- ""
		case "reply":
			str, err := t.Rcon.Send(d.line)
			if err != nil {
				fmt.Println(err)
			}
			d.reply <- str
		}
		close(d.reply)
	}
}

func (t *Tracker) parseTags(pid int, m string) string {
	if strings.Contains(m, "$PN$") {
		//replace with player name.
		m = strings.Replace(m, "$PN$", t.players[pid].Name, -1)
	}
	if strings.Contains(m, "$PL$") {
		//replace with player level:
		m = strings.Replace(m, "$PL$", t.players[pid].Level, -1)
	}
	if strings.Contains(m, "$PT$") {
		//replace with player team
		var team string
		if t.players[pid].Team == "1" {
			team = "National"
		} else if t.players[pid].Team == "2" {
			team = "Royal"
		} else {
			team = t.players[pid].Team
		}
		m = strings.Replace(m, "$PT$", team, -1)
	}
	return m
}
