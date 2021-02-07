package main

// https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go

import (
	"bufio"
	"flag"
	"fmt"
	"golang-redis/commands"
	"golang-redis/resp"
	"net"
	"os"
)

// Redis server constants
const (
	RedisHost = "localhost"
	RedisPort = "3333"
	connType  = "tcp"
)

func main() {
	host := flag.String("host", RedisHost, "Remote redis server HostIPs")
	port := flag.String("port", RedisPort, "Remote Redis server Listen port")
	flag.Parse()
	// Listen for incoming connections.
	l, err := net.Listen(connType, *host+":"+*port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + RedisHost + ":" + RedisPort)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// takeFullInput is a custom splitFunc of type SplitFunc that
// takes in the full CRLF feed for processing.
func takeFullInput(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if atEOF == true {
		return 0, []byte{}, nil
	}
	return len(data), data, nil
}

// Handles incoming TCP client session request.
func handleRequest(conn net.Conn) {
	defer conn.Close()
	// Create a new reader
	scanner := bufio.NewScanner(conn)
	//Use custom splitFunc to read complete input
	scanner.Split(takeFullInput)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		// The basic premise is as follows. The incoming message is parsed by an appropriate
		// parser. If any of the parsers panic, we recover and return RedisError serialized
		// to the client. Otherwise, we execute the command using CommandExecutor
		ras, _, f := resp.ParseRedisClientRequest(bytes)
		if f == resp.EmptyRedisError {
			for _, ra := range ras {
				dataType, err := commands.ExecuteStringCommand(ra)
				if err != resp.EmptyRedisError {
					conn.Write([]byte(err.ToString() + "\n"))
				} else {
					if dataType == nil {
						conn.Write([]byte("(nil)" + "\n"))
					} else {
						conn.Write([]byte(dataType.ToString() + "\n"))
					}
				}
			}
		} else {
			conn.Write([]byte(f.ToString() + "\n"))
		}
	}
}
