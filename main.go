/* gorcon version 13.12.3 (lee8oi)

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon package contains the essential functions needed for, connecting to &
running commands on, BF2CC based Rcon servers.

*/
package gorcon

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Config struct {
	Admin, Address, RconPort, Pass string
}

type Rcon struct {
	pass, seed, service string
	socket              net.Conn
}

//Connect to rcon server and grab seed.
func (r *Rcon) Connect(addr string) (err error) {
	r.service = addr
	r.socket, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(r.socket)
	for scanner.Scan() {
		if text := scanner.Text(); strings.Contains(text, "### Digest seed:") {
			splitLine := strings.Split(text, ":")
			r.seed = strings.TrimSpace(splitLine[1])
			break
		}
	}
	if err = scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return
}

//Login encrypts seed & pass, performs authentication.
func (r *Rcon) Login(pass string) (err error) {
	r.pass = pass
	hash := md5.New()
	hash.Write([]byte(r.seed + pass))
	loginhash := fmt.Sprintf("%x", hash.Sum(nil))
	if _, err = r.Write("login " + loginhash); err != nil {
		return
	}
	if text, err := r.ReadAll(); err == nil {
		if strings.Contains(text, "Authentication successful, rcon ready.") {
			fmt.Println("Authentication successful.")
		}
	}
	return
}

//ReadAll data up to the EOT (and trim it off).
func (r *Rcon) ReadAll() (string, error) {
	result, err := bufio.NewReader(r.socket).ReadString('\u0004')
	return strings.Trim(result, "\u0004"), err
}

//Reconnect the current Rcon. If connection fails - wait duration & try again.
//Returns on Login errors or upon successful reconnection.
func (r *Rcon) Reconnect(duration time.Duration) error {
	for {
		fmt.Println("Attempting reconnection.")
		if err := r.Connect(r.service); err != nil {
			fmt.Println("Reconnection attempt failed. Waiting.")
			time.Sleep(duration)
			continue
		}
		if err := r.Login(r.pass); err != nil {
			return err
		}
		fmt.Println("Reconnection successful.")
		break
	}
	return nil
}

//Send an rcon command and return response.
func (r *Rcon) Send(cmd string) (string, error) {
	_, err := r.Write(cmd)
	if err != nil {
		return "", err
	}
	data, err := r.ReadAll()
	if err != nil {
		return "", err
	}
	return strings.Trim(data, "\n"), nil
}

//SetAdmin sets the admin name for the rcon connection.
func (r *Rcon) SetAdmin(name string) (string, error) {
	data, err := r.Send("bf2cc setadminname " + name)
	if err != nil {
		return "", err
	}
	return data, err
}

//Start creates a connection, performing login, using Config variable values.
func (r *Rcon) Start(c *Config) {
	if err := r.Connect(c.Address + ":" + c.RconPort); err != nil {
		fmt.Println(err)
		return
	}
	if err := r.Login(c.Pass); err != nil {
		fmt.Println(err)
		return
	}
}

//Write prefixes line to enable EOT & writes command to the rcon connection.
func (r *Rcon) Write(line string) (int, error) {
	if r == nil {
		return 1, errors.New("no connection available")
	}
	return r.socket.Write([]byte("\u0002" + strings.TrimSpace(line) + "\n"))
}
