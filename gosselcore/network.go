package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/ayonix/gossel/gosselproto"
	"github.com/sorcix/irc"
)

type Identity struct {
	User     string
	Realname string
	AwayMsg  string
}

func NewIdentity(user, real, away string) Identity {
	return Identity{
		User:     user,
		Realname: real,
		AwayMsg:  away,
	}
}

func NewIdentityProto(id *pb.Identity) Identity {
	return Identity{
		User:     id.Username,
		Realname: id.Realname,
		AwayMsg:  id.Awaymsg,
	}
}

type Subscriptions map[string][](chan *irc.Message)

type Network struct {
	Name      string
	Server    string
	Identity  Identity
	Password  string
	TLSConfig *tls.Config
	Data      chan *irc.Message
	SendQueue chan *irc.Message

	user   User
	reader *irc.Decoder
	writer *irc.Encoder
	conn   net.Conn
	tries  uint
}

func NewNetwork(user User, name, server, password string, identity Identity) *Network {
	return &Network{
		Name:      name,
		Server:    server,
		Identity:  identity,
		Password:  password,
		TLSConfig: nil,
		Data:      make(chan *irc.Message, 10),
		SendQueue: make(chan *irc.Message, 10),
		user:      user,
		tries:     0,
	}
}

func NewNetworkTls(user User, name, server, password string, identity Identity, tls *tls.Config) *Network {
	return &Network{
		Name:      name,
		Server:    server,
		Identity:  identity,
		Password:  password,
		TLSConfig: tls,
		Data:      make(chan *irc.Message, 10),
		SendQueue: make(chan *irc.Message, 10),
		user:      user,
		tries:     0,
	}
}

func NewNetworkProto(user User, net *pb.Network) *Network {
	id := NewIdentityProto(net.Identity)
	if net.Tls {
		return NewNetworkTls(user, net.Name, net.Network, net.Password, id, &tls.Config{})
	} else {
		return NewNetwork(user, net.Name, net.Network, net.Password, id)
	}
}

func (n *Network) Start() {
	n.Connect()
	go n.HandleLoop()
}

func (n *Network) Connect() (err error) {
	n.tries = 0
	if n.TLSConfig == nil {
		n.conn, err = net.Dial("tcp", n.Server)
	} else {
		n.conn, err = tls.Dial("tcp", n.Server, n.TLSConfig)
	}
	n.reader = irc.NewDecoder(n.conn)
	n.writer = irc.NewEncoder(n.conn)

	for _, msg := range n.connectMessages() {
		n.Send(msg)
	}

	go n.readLoop()

	return err
}

func (n *Network) Disconnect(msg string) {
	n.SendQueue <- &irc.Message{
		Command: irc.QUIT,
	}
}

func (n *Network) connectMessages() []*irc.Message {
	messages := []*irc.Message{}
	if n.Password != "" {
		messages = append(messages, &irc.Message{
			Command: irc.PASS,
			Params:  []string{n.Password},
		})
	}

	messages = append(messages, &irc.Message{
		Command: irc.NICK,
		Params:  []string{n.Identity.User},
	})

	messages = append(messages, &irc.Message{
		Command:  irc.USER,
		Params:   []string{n.Identity.User, "0", "*"},
		Trailing: n.Identity.Realname,
	})
	return messages
}

func (n *Network) Reconnect() error {
	log.Println("Reconnecting")
	n.conn.Close()
	return n.Connect()
}

func (n *Network) Send(m *irc.Message) error {
	if n.writer != nil {
		log.Printf(">> %s \n", m.String())
		return n.writer.Encode(m)
	} else {
		return fmt.Errorf("Network %s is not connected", n.Name)
	}
}

func (n *Network) SendProto(m *pb.Irc) {
	n.SendQueue <- &irc.Message{
		Command: m.Command,
		Prefix: &irc.Prefix{
			Name: m.Prefix.Name,
			User: m.Prefix.User,
			Host: m.Prefix.Host,
		},
		Params:   m.Params,
		Trailing: m.Trailing,
	}
}

func (n *Network) Join(channel string) {
	n.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{channel},
	})
}

func (n *Network) HandleLoop() {
	for {
		select {
		case msg, ok := <-n.Data:
			if !ok {
				log.Println("Error in HandleLoop")
				return
			}
			switch msg.Command {
			case irc.PING:
				n.Send(&irc.Message{
					Command:  irc.PONG,
					Params:   msg.Params,
					Trailing: msg.Trailing,
				})
			default:
				if msg != nil {
					n.user.Broadcast(n.Name, msg)
				}
			}
		case msg := <-n.SendQueue:
			n.Send(msg)
		}
	}
}

func (n *Network) readLoop() error {
	for {
		n.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := n.reader.Decode()
		if err != nil {
			//			return n.Reconnect()
			return err
		}
		log.Printf("<< %s \n", msg)
		n.Data <- msg
	}
}
