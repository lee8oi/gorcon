/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

command methods are used to process in-game commands.
*/

//
package track

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//interpret monitors com channel for messages sent from parseChat(). Used to interpret
//commands in messages.
func (t *Tracker) interpret(com chan *message) {
	for m := range com {
		id, _ := strconv.Atoi(m.Pid)
		split := strings.Split(m.Text[1:], " ")
		Log(fmt.Sprintf("%s[%s]: %s\n", m.Origin, m.Time, m.Text))
		if m.IsCommand {
			line := ""
			permitted := false
			public := false
			//if split[0] == "" {
			//	return
			//}
			if _, ok := t.aliases[split[0]]; !ok {
				return
			}
			if t.admins[t.players[id].Nucleus].Power >= t.aliases[split[0]].Power {
				permitted = true
			}
			if t.aliases[split[0]].Power == 0 {
				//public alias (no power check)
				public = true
			}
			if !public && !permitted {
				//fmt.Printf("%s - not enough power\n", t.players[id].Name)
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
						t.Rcon.Enqueue(l)
					} else if len(r) > 1 {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("multiple players found ('%s')", split[1]))
						t.Rcon.Enqueue(l)
					} else {
						fmt.Printf("No results found.")
					}
				}
			case split[0] == "promote" || split[0] == "demote":
				val := 1
				if split[0] == "demote" {
					val = 0
				}
				if len(split) > 1 {
					r := t.players.find(split[1])
					if len(r) == 1 {
						name := t.players[r[0]].Name
						nucleus := t.players[r[0]].Nucleus
						l := fmt.Sprintf(`exec game.setPersonaVipStatus %s %s %s`, name, nucleus, val)
						t.Rcon.Enqueue(l)
					} else if len(r) > 1 {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("multiple players found ('%s')", split[1]))
						t.Rcon.Enqueue(l)
					} else {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("player not found ('%s')", split[1]))
						t.Rcon.Enqueue(l)
					}
				}
			case split[0] == "info":
				if len(split) > 1 {
					r := t.players.find(split[1])
					if len(r) == 1 {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, t.aliases[split[0]].Message+" "+line)
						l = t.parseTags(r[0], l)
						t.Rcon.Enqueue(l)
					} else if len(r) > 1 {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("multiple players found ('%s')", split[1]))
						t.Rcon.Enqueue(l)
					} else {
						l := fmt.Sprintf(`exec game.sayToPlayerWithId %d "%s"`, id, fmt.Sprintf("player not found ('%s')", split[1]))
						t.Rcon.Enqueue(l)
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
				t.Rcon.Enqueue(full)
			}
		}
	}
}

//parseTags scans the message text for special tags used to represent certain data
//like player name, etc.
func (t *Tracker) parseTags(pid int, m string) string {
	tags, err := regexp.Compile(`\$+[A-Z]+\$`)
	if err != nil {
		fmt.Println(err)
	}
	result := tags.ReplaceAllFunc([]byte(m), func(b []byte) (r []byte) {
		//fmt.Println(fmt.Sprintf("%s", b))
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
		case "$VIP$":
			if t.players[pid].Vip == "1" {
				r = []byte("VIP")
			} else {
				r = []byte("")
			}
		case "$PING$":
			r = []byte(t.players[pid].Ping)
		}
		return
	})
	return fmt.Sprintf("%s", result)
}
