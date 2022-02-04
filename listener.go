package network

import (
	"container/list"
	"fmt"
	"github.com/isshoni-soft/edward/packet"
	"github.com/isshoni-soft/edward/protocol"
	"net"
)

// Network Listener
// TODO: allow this to listen on an ip address and port #
// TODO: open a new connection listener thread on ip addr & port
// TODO: on new connection open a new packet channel
// TODO: keep a reference to that packet channel for communication

type Listener struct {
	Address        string
	Port           string
	Encoder        packet.Encoder
	Protocol       protocol.Manager
	ChannelPreInit func(channel packet.Channel, protocol protocol.Manager)

	running     bool
	connections *list.List
	listener    net.Listener
	shutdown    chan bool
}

func (l *Listener) Start() {
	if l.running {
		return
	}

	if l.Protocol == nil {
		panic(fmt.Errorf("network listener requires a protocol"))
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
		panic(err)
	}

	l.shutdown = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-l.shutdown:
				return
			default:
			}

			conn, _ := l.listener.Accept()
			fmt.Println("Accepted new connection")

			fmt.Println("Creating channel...")
			channel := packet.NewChannel(conn, l.Encoder)
			fmt.Println("Applying protocol...")
			l.Protocol.RegisterListeners(channel)
			l.ChannelPreInit(channel, l.Protocol)
			fmt.Println("Starting channel...")
			channel.Start()

			l.connections.PushBack(channel)
			fmt.Println("Channel initialized!")
		}
	}()
}

func (l *Listener) OnAllConnections(f func(channel packet.Channel)) {
	connections := l.Connections()

	for e := connections.Front(); e != nil; e = e.Next() {
		channel := e.Value.(packet.Channel)

		f(channel)
	}
}

func (l Listener) Connections() list.List {
	return *l.connections
}
