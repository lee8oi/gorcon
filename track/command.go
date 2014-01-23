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
	"regexp"
	"strconv"
	"strings"
)

type process struct {
	instruct, line string
	reply          chan string
}

func (t *Tracker) process(instruct, line string) string {
	reply := make(chan string)
	t.proc <- process{instruct, line, reply}
	return <-reply
}

//interpret monitors com channel for messages sent from parseChat(). Used to interpret
//command lines in messages and create processes to handle them.
func (t *Tracker) interpret(com chan *message) {
	for m := range com {
		id, _ := strconv.Atoi(m.Pid)
		split := strings.Split(m.Text[1:], " ")
		fmt.Printf("%s[%s]: %s\n", m.Origin, m.Time, m.Text)
		if m.IsCommand {
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
				var cmd string
				if len(split) > 1 {
					line = strings.Join(split[1:], " ")
				}
				switch t.aliases[split[0]].Visibility {
				case "public":
					cmd = fmt.Sprintf(`bf2cc sendserverchat %s`, t.aliases[split[0]].Message+" "+line)
				case "private":
					cmd = fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, t.aliases[split[0]].Message+" "+line)
				case "server":
					cmd = t.aliases[split[0]].Message + " " + line
				default:
					continue
				}
				full := t.parseTags(id, cmd)
				t.process("send", full)
			}
		}
	}
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
			fmt.Printf("%s\n", d.line)
			str, err := t.Rcon.Send(d.line)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(str)
			//d.reply <- ""
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
	tags, err := regexp.Compile(`\$+[A-Z]+\$`)
	if err != nil {
		fmt.Println(err)
	}
	result := tags.ReplaceAllFunc([]byte(m), func(b []byte) (r []byte) {
		fmt.Println(fmt.Sprintf("%s", b))
		//return []byte("value")
		switch fmt.Sprintf("%s", b) {
		case "$PN$":
			r = []byte(t.players[pid].Name)
		case "$PL$":
			r = []byte(t.players[pid].Level)
		case "$PT$":
			r = []byte(t.players[pid].Team)
		case "$PC$":
			r = []byte(t.players[pid].Kit)
		case "$ET$":
			if t.players[pid].Team == "2" {
				r = []byte("National")
			}
			r = []byte("Royal")
		case "$PTN$":
			if t.players[pid].Team == "1" {
				r = []byte(t.game.Nsize)
			}
			r = []byte(t.game.Rsize)
		}
		return
	})
	return fmt.Sprintf("%s", result)
}
