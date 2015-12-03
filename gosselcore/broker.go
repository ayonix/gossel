package main

import (
	"fmt"

	pb "github.com/ayonix/gossel/gosselproto"
	"github.com/sorcix/irc"
)

type Broker struct {
	ToClient      chan *irc.Message                // channel for the to broadcasted messages
	subscriptions map[chan *pb.Msg]map[string]bool // map client to map[network/#channel]bool
}

func NewBroker() *Broker {
	return &Broker{
		ToClient:      make(chan *irc.Message),
		subscriptions: make(map[chan *pb.Msg]map[string]bool),
	}
}

func (b *Broker) Register(ch chan *pb.Msg) {
	b.subscriptions[ch] = make(map[string]bool)
}

func (b *Broker) Unregister(ch chan *pb.Msg) {
	delete(b.subscriptions, ch)
}

func (b *Broker) Subscribe(ch chan *pb.Msg, net, channel string) {
	sub := fmt.Sprintf("%s/%s", net, channel)
	b.subscriptions[ch][sub] = true
}

func (b *Broker) Unsubscribe(ch chan *pb.Msg, net, channel string) {
	sub := fmt.Sprintf("%s/%s", net, channel)
	delete(b.subscriptions[ch], sub)
}

// check which channels to send the message to
func (b *Broker) Broadcast(network_name string, msg *irc.Message) {
	// TODO Handle subscriptions
	for ch := range b.subscriptions {
		_ = "breakpoint"
		ch <- toProto(network_name, msg)
	}
}

func toProto(network string, msg *irc.Message) *pb.Msg {
	return &pb.Msg{
		MessageType: &pb.Msg_Irc{
			Irc: &pb.Irc{
				Command:  msg.Command,
				Prefix:   toProtoPrefix(msg.Prefix),
				Params:   msg.Params,
				Trailing: msg.Trailing,
				Network:  network,
			},
		},
	}
}

func toProtoPrefix(prefix *irc.Prefix) *pb.Irc_Prefix {
	if prefix != nil {
		return &pb.Irc_Prefix{
			Name: prefix.Name,
			User: prefix.User,
			Host: prefix.Host,
		}
	}
	return nil
}
