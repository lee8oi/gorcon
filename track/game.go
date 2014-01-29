/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

*/

//
package track

import (
	"fmt"
	"strings"
)

type game struct {
	Name, Ranked, Balance, Map, Mode, Round, Players, Joining,
	Ntickets, Nsize, Rtickets, Rsize, Elapsed, Remaining string
}

func (g *game) update(data string) {
	if len(data) > 1 {
		splitLine := strings.Split(data, "\t")
		mode := strings.Split(splitLine[20], "_")[1]
		if len(splitLine) < 27 {
			return
		}
		*g = game{
			Name:      splitLine[7],
			Ranked:    splitLine[25],
			Balance:   splitLine[24],
			Map:       cleanMapName(splitLine[5]),
			Mode:      strings.ToUpper(mode),
			Round:     splitLine[31],
			Players:   splitLine[3],
			Joining:   splitLine[4],
			Ntickets:  splitLine[11],
			Nsize:     splitLine[26],
			Rtickets:  splitLine[16],
			Rsize:     splitLine[27],
			Elapsed:   splitLine[18],
			Remaining: splitLine[19],
		}
	}
	if err := writeJSON("game.json", g); err != nil {
		fmt.Println(err)
	}
}

func cleanMapName(name string) string {
	switch name {
	case "dependant_day":
		return "Inland Invasion"
	case "dependant_day_night":
		return "Inland Invasion Night"
	case "heat":
		return "Riverside Rush"
	case "heat_snow":
		return "Riverside Rush Snow"
	case "lake":
		return "Buccaneer Bay"
	case "lake_night":
		return "Buccaneer Bay Night"
	case "lake_snow":
		return "Buccaneer Bay Snow"
	case "lunar":
		return "Lunar Landing"
	case "mayhem":
		return "Sunset Showdown"
	case "river":
		return "Fortress Frenzy"
	case "royal_rumble":
		return "Perilous Port Night"
	case "royal_rumble_day":
		return "Perilous Port Day"
	case "royal_rumble_snow":
		return "Perilous Port Snow"
	case "ruin":
		return "Midnight Mayhem"
	case "ruin_day":
		return "Morning Mayhem"
	case "ruin_snow":
		return "Midnight Mayhem Snow"
	case "seaside_skirmish":
		return "Seaside Skirmish"
	case "seaside_skirmish_night":
		return "Seaside Skirmish Night"
	case "smack2":
		return "Coastal Clash"
	case "smack2_night":
		return "Coastal Clash Night"
	case "smack2_snow":
		return "Coastal Clash Snow"
	case "village":
		return "Victory Village"
	case "village_snow":
		return "Victory Village Snow"
	case "wicked_wake":
		return "Wicked Wake"
	case "woodlands":
		return "Alpine Assault"
	case "woodlands_snow":
		return "Alpine Assault Snow"
	default:
		return name
	}
}
