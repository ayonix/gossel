package main

import (
	"fmt"
	"log"

	"code.google.com/p/go.crypto/bcrypt"
	pb "github.com/ayonix/gossel/gosselproto"
)

type User struct {
	*Broker
	Name       string
	Password   []byte
	Networks   map[string]*Network
	Identities []Identity
}

func NewUser(name, password string) User {
	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	b := NewBroker()
	u := User{
		Broker:     b,
		Name:       name,
		Password:   pw,
		Networks:   make(map[string]*Network),
		Identities: make([]Identity, 1),
	}
	return u
}

// Returns nil on success and an error otherwise
func (u User) Authenticate(password string) error {
	return bcrypt.CompareHashAndPassword(u.Password, []byte(password))
}

func (u *User) AddNetwork(net *pb.Network) (err error) {
	if _, ok := u.Networks[net.Name]; ok {
		fmt.Errorf("Network %s already added", net.Name)
	}
	n := NewNetworkProto(*u, net)
	if n != nil {
		u.Networks[net.Name] = n
		go n.Start()
		log.Printf("Network %s started \n", net.Name)
		return nil
	} else {
		return fmt.Errorf("Couldn't add network")
	}
}

func (u *User) RemoveNetwork(name string) {
	if net, ok := u.Networks[name]; ok {
		net.Disconnect("")
		delete(u.Networks, name)
	}
}
