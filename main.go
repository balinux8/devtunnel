package main

import (
	"fmt"
	"github.com/hashicorp/yamux"
	"github.com/spf13/cobra"
	"io"
	"net"
)

// Config contains configuration for launching a daemon
type Config struct {
	Services []ServiceDef // list service definitions
}

// ServiceDef is a service definition
type ServiceDef struct {
	Address string // TCP address Eg: localhost:9090
	Name    string // Name of service Eg: redis/kafka
}

type server yamux.Stream
func Decode(stream *yamux.Stream) error {
	stream.Read()
	return nil
}

func runAgent() {
}

// bridge integrate a stream to underlying service
func bridge(stream *yamux.Stream, svc ServiceDef) error {
	// dial to underlying service
	conn, err := net.Dial("tcp", svc.Address)
	if err != nil {
		return err
	}

	// pairs stream and underlying service together
	go io.Copy(conn, stream)
	go io.Copy(stream, conn)

	return nil
}

func runDaemon(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	for {
		// wait for agents
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		go func() {
			fmt.Println("accept new conn from", conn.RemoteAddr().String())
			session, err := yamux.Server(conn, nil)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			go func() {
				for {
					stream, err := session.Accept()
					if err != nil {
						fmt.Println(err.Error())
						return
					}

					// we can exchange data from here
					_ = stream
					// each stream must be paired to underlying service
					// to do so, client must
					// specify which service the client want to connect to.
					// and server must wait for some header to know it
				}
			}()
		}()
	}
}

func main() {
	rootCmd := cobra.Command{
		Use:   "devtunnel",
		Short: "devtunnel",
		Long:  "devtunnel",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	daemon := &cobra.Command{
		Use:   "daemon",
		Short: "daemon",
		Long:  "daemon",
		Run: func(cmd *cobra.Command, args []string) {
			runDaemon("localhost:9999")

			fmt.Println("daemon")
			listener, err := net.Listen("tcp", "localhost:9999")
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			fmt.Println("accepted a TCP connection")
			session, err := yamux.Server(conn, nil)
			if err != nil {
				panic(err)
			}
			stream, err := session.Accept()
			if err != nil {
				panic(err)
			}
			fmt.Println("accepted a yamux connection")
			stream.Write([]byte("hello, i'm tunneld"))
		},
	}

	agent := &cobra.Command{
		Use:   "agent",
		Short: "agent",
		Long:  "agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("agent")
			conn, err := net.Dial("tcp", "localhost:9999")
			if err != nil {
				panic(err)
			}
			fmt.Println("connected to server")
			session, err := yamux.Client(conn, nil)
			if err != nil {
				panic(err)
			}
			stream, err := session.Open()
			fmt.Println("openned a stream")
			buf := make([]byte, 512)
			stream.Read(buf)
			fmt.Println("from server:", string(buf))
		},
	}

	rootCmd.AddCommand(daemon, agent)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
