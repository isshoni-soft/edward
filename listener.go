package network

import (
	"container/list"
	"fmt"
	error2 "github.com/isshoni-soft/edward/error"
	"github.com/isshoni-soft/edward/packet"
	"net"
)

type Listener struct {
	Address        string
	Port           string
	Encoder        packet.Encoder
	Protocol       Manager
	ChannelPreInit func(channel packet.Channel, protocol Manager)

	running     bool
	connections *list.List
	listener    net.Listener
	shutdown    chan bool
}

func (l *Listener) Start() error {
	if l.running {
		return error2.ListenerRunning{}
	}

	if l.Protocol == nil {
		return error2.NoProtocol{}
	}

	if l.connections == nil {
		l.connections = list.New()
	}

	if l.Encoder == nil {
		l.Encoder = packet.NewSimpleEncoder()
	}

	l.Protocol.RegisterPackets(l.Encoder.PacketRegistry())

	if li, err := net.Listen("tcp", l.Address+":"+l.Port); err == nil {
		l.listener = li
	} else {
		return err
	}

	l.shutdown = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-l.shutdown:
				return
			default:
			}

			conn, e := l.listener.Accept()

			if e != nil {
				fmt.Println("Incoming connection errored:", e)
				continue
			}

			fmt.Println("Accepted new connection")
			fmt.Println("Creating channel...")
			channel := packet.NewChannel(conn, l.Encoder)
			fmt.Println("Applying protocol...")
			l.Protocol.RegisterListeners(channel)
			l.ChannelPreInit(channel, l.Protocol)
			fmt.Println("Starting channel...")
			channel.Start()

			elem := l.connections.PushBack(channel)

			fmt.Println("Registering close callback")
			channel.SetCloseCallback(func(c packet.Channel) {
				fmt.Println("Removing", elem)
				l.connections.Remove(elem)
			})

			fmt.Println("Channel initialized!")
		}
	}()

	return nil
}

func (l *Listener) Close() {
	l.shutdown <- true
}

func (l *Listener) OnAllConnections(f func(channel packet.Channel)) {
	connections := l.Connections()

	for e := connections.Front(); e != nil; e = e.Next() {
		channel := e.Value.(packet.Channel)

		fmt.Println("Iterating on", channel.UUID())

		f(channel)
	}
}

func (l Listener) Connections() list.List {
	return *l.connections
}
