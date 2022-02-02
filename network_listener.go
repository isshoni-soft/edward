package network

import (
	"container/list"
	"fmt"
	"net"
)

// Network Listener
// TODO: allow this to listen on an ip address and port #
// TODO: open a new connection listener thread on ip addr & port
// TODO: on new connection open a new packet channel
// TODO: keep a reference to that packet channel for communication

type Listener struct {
	Address string
	Port    string

	running     bool
	connections *list.List
	listener    net.Listener
	shutdown    chan bool
}

func (l *Listener) Start() {
	if l.running {
		return
	}

	if l.connections == nil {
		l.connections = list.New()
	}

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

			channel := NewPacketChannel(conn)
			channel.Start()

			l.connections.PushBack(channel)
		}
	}()
}
