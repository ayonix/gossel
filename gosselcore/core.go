package main

import (
	"errors"
	"fmt"
)

type Gosselcore struct {
	users map[string]User
}

func NewGosselcore() *Gosselcore {
	return &Gosselcore{
		users: make(map[string]User),
	}
}

func (g *Gosselcore) AddUser(name, password string) error {
	if _, ok := g.users[name]; ok {
		return errors.New("User already added")
	}
	u := NewUser(name, password)
	g.users[name] = u
	return nil
}

func (g *Gosselcore) FindUser(name string) (*User, error) {
	if u, ok := g.users[name]; ok {
		return &u, nil
	} else {
		return nil, fmt.Errorf("User %s not found", name)
	}
}
