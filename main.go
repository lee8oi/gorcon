//The core Rcon package for gorcon2
package gorcon

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Rcon struct {
	Admin   string
	pass    string
	seed    string
	service string
	socket  net.Conn
}

//Establish connection & grab seed.
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

//Run command and return all data.
func (r *Rcon) GetData(cmd string) (string, error) {
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

//Encrypt seed & pass, perform authentication, and set admin name.
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
			result, err := r.SetAdmin(r.Admin)
			if err != nil {
				fmt.Println("Error setting admin name:"+result, err)
			}
			fmt.Println("Authentication successful.", result)
		}
	}
	return
}

//Read all incoming data up to (but not including) EOT.
func (r *Rcon) ReadAll() (string, error) {
	result, err := bufio.NewReader(r.socket).ReadString('\u0004')
	return strings.Trim(result, "\u0004"), err
}

//Attempt to reconnect to existing Rcon socket.
func (r *Rcon) Reconnect() (string, error) {
	if err := r.Connect(r.service); err != nil {
		return "", err
	}
	if err := r.Login(r.pass); err != nil {
		return "", err
	}
	return "Successful", nil
}

//Set admin name on rcon server.
func (r *Rcon) SetAdmin(name string) (string, error) {
	data, err := r.GetData("bf2cc setadminname " + name)
	if err != nil {
		return "", err
	}
	return data, err
}

//Prefix line to enable EOT & write to connection.
func (r *Rcon) Write(line string) (int, error) {
	if r == nil {
		return 1, errors.New("no connection available")
	}
	return r.socket.Write([]byte("\u0002" + strings.TrimSpace(line) + "\n"))
}
