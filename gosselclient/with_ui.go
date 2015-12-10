package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/andlabs/ui"
	"github.com/ayonix/gossel/gosselproto"
	"github.com/sorcix/irc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	window     ui.Window
	serverAddr = flag.String("host", "localhost:4343", "Serveraddress with port where the core listens on")
	channel    = flag.String("channel", "#gossel_test", "")
	network    = flag.String("network", "freenode", "")
	user       = flag.String("user", "Test", "")
	password   = flag.String("password", "Test", "")
	stream     gosselproto.Gosselcore_ConnectClient
	backlog    ui.Grid
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	runUi()
	wg.Wait()
}

func runUi() {
	log.Println("runUi")
	go ui.Do(func() {
		input := ui.NewTextField()
		button := ui.NewButton("Send")
		backlog = ui.NewGrid()
		stack := ui.NewVerticalStack(
			backlog,
			input,
			button)
		window = ui.NewWindow("DurrClient", 400, 400, stack)
		button.OnClicked(func() {
			sendIrcMessage(input.Text())
			input.SetText("")
		})
		window.OnClosing(func() bool {
			ui.Stop()
			return true
		})
		log.Println("go connect")
		go connectCore()
		window.Show()
	})
	err := ui.Go()
	if err != nil {
		panic(err)
	}
}

func sendIrcMessage(msg string) {
	stream.Send(gosselproto.IrcMsg(*network, irc.PRIVMSG, nil, []string{*channel}, msg))
}

func connectCore() {
	log.Println("Connecting")
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gosselproto.NewGosselcoreClient(conn)
	stream, err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// First thing to do is auth
	err = stream.Send(gosselproto.AuthMsg(*user, *password))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Sent auth")
	err = stream.Send(gosselproto.SubscribeMsg(*network, *channel, false))
	if err != nil {
		log.Fatal(err)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Fatal("Got EOF")
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Got message: %v \n", in)
		switch x := in.MessageType.(type) {
		case *gosselproto.Msg_Irc:
			str := fmt.Sprintf("%s: %s", x.Irc.Prefix.Name, x.Irc.Trailing)
			backlog.Add(ui.NewLabel(str), nil, ui.South, true, ui.LeftTop, true, ui.LeftTop, 1, 1)
		default:
			backlog.Add(ui.NewLabel(in.String()), nil, ui.South, true, ui.LeftTop, true, ui.LeftTop, 1, 1)
		}
	}
}
