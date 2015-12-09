package main

import (
	"flag"
	"io"
	"log"
	"sync"

	"github.com/ayonix/gossel/gosselproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var serverAddr = flag.String("host", "localhost:4343", "Serveraddress with port where the core listens on")

func main() {
	var wg sync.WaitGroup

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gosselproto.NewGosselcoreClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to server")

	// First thing to do is auth
	err = stream.Send(authMsg("Test", "Test"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Sent auth")

	wg.Add(1)
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				log.Fatal("Got EOF")
			}
			if err != nil {
				log.Fatal(err)
			}
			switch x := in.MessageType.(type) {
			case *gosselproto.Msg_Irc:
				log.Printf("<< %v \n", x.Irc)
			default:
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func authMsg(user, pw string) *gosselproto.Msg {
	return &gosselproto.Msg{
		MessageType: &gosselproto.Msg_Auth{
			Auth: &gosselproto.Auth{
				Username: user,
				Password: pw,
			},
		},
	}
}

func subscribeMsg(network, channel string, unsubscribe bool) *gosselproto.Msg {
	return &gosselproto.Msg{
		MessageType: &gosselproto.Msg_Subscribe{
			Subscribe: &gosselproto.Subscribe{
				Network:     network,
				Channel:     channel,
				Unsubscribe: unsubscribe,
			},
		},
	}
}

func identityMsg(user, real, away string) *gosselproto.Identity {
	return &gosselproto.Identity{
		Username: user,
		Realname: real,
		Awaymsg:  away,
	}
}

func networkMsg(name, host, password string, tls bool, identity *gosselproto.Identity, add bool) *gosselproto.Msg {
	return &gosselproto.Msg{
		MessageType: &gosselproto.Msg_Network{
			Network: &gosselproto.Network{
				Name:     name,
				Network:  host,
				Password: password,
				Tls:      tls,
				Identity: identity,
				Add:      add,
			},
		},
	}
}

func prefix(name, user, host string) *gosselproto.Irc_Prefix {
	return &gosselproto.Irc_Prefix{
		Name: name,
		User: user,
		Host: host,
	}
}

func ircMsg(network, command string, prefix *gosselproto.Irc_Prefix, params []string, trailing string) *gosselproto.Msg {
	return &gosselproto.Msg{
		MessageType: &gosselproto.Msg_Irc{
			Irc: &gosselproto.Irc{
				Command:  command,
				Prefix:   prefix,
				Params:   params,
				Trailing: trailing,
				Network:  network,
			},
		},
	}
}
