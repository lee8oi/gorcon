/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.15 (lee8oi)

Groups and its methods are used to track user groups & permission power for players.
Player profile IDs are used to identify the player.
*/

//
package track

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type user struct {
	Power int
	Name  string
	Id    string
}

type group struct {
	Power   int
	Members map[string]user
}

func (g *group) add(pid, name string) {
	g.Members[pid] = user{Power: g.Power, Name: name, Id: pid}
}

func (g *group) remove(id string) {
	delete(g.Members, id)
}

func (g *group) member(id string) bool {
	if len(g.Members[id].Id) > 0 {
		return true
	}
	return false
}

//save a snapshot of the current group to file as JSON.
func (g *group) save(path string) {

	b, err := json.Marshal(g)
	fmt.Println(fmt.Sprintf("String: %s", b))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

//load reads a snapshot file & updates current group
func (g *group) load(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = json.Unmarshal(b, g)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
