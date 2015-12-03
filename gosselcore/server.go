package main

import (
	"fmt"
	"io"
	"log"

	pb "github.com/ayonix/gossel/gosselproto"
)

type gosselcoreServer struct {
	c *Gosselcore
}

func (g *gosselcoreServer) Connect(stream pb.Gosselcore_ConnectServer) error {
	// first: check for authentication
	var user *User

	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	switch x := in.MessageType.(type) {
	case *pb.Msg_Auth:
		user, err = g.c.FindUser(x.Auth.Username)
		if err != nil {
			return err
		}
		if err = user.Authenticate(x.Auth.Password); err != nil {
			return err
		}
	default:
		fmt.Errorf("Expected auth message, got %x")
	}

	outgoing := make(chan *pb.Msg, 10)
	user.Register(outgoing)
	incoming := make(chan *pb.Msg, 10)
	go handleIncoming(incoming, stream)

	// now loop and read messages
	for {
		select {
		case in := <-incoming:
			log.Printf("Got core<-client %v \n", in)
			switch x := in.MessageType.(type) {
			case *pb.Msg_Network:
				if x.Network.Add {
					user.AddNetwork(x.Network)
				} else {
					user.RemoveNetwork(x.Network.Name)
				}

			case *pb.Msg_Control:
			case *pb.Msg_Subscribe:
				if !x.Subscribe.Unsubscribe {
					user.Subscribe(outgoing, x.Subscribe.Network, x.Subscribe.Channel)
				} else {
					user.Unsubscribe(outgoing, x.Subscribe.Network, x.Subscribe.Channel)
				}
			case *pb.Msg_Irc:
				net := user.Networks[x.Irc.Network]
				net.SendProto(x.Irc)
			case nil:
				log.Println("No field was set")
			default:
				// TODO: just ignore it?
				return fmt.Errorf("Msg.Msg has unexpected type %T", x)
			}
		case out := <-outgoing:
			switch out.MessageType.(type) {
			case *pb.Msg_Irc:
				stream.Send(out)
			default:
				log.Println("Out ignored")
			}
		}
	}
}

func handleIncoming(ch chan *pb.Msg, stream pb.Gosselcore_ConnectServer) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("EOF")
			return
		}
		if err != nil {
			log.Printf("Error while handling incoming: %s", err)
			return
		}
		ch <- in
	}
}
