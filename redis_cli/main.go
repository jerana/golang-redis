package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
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
	fmt.Printf("Connecting to redis Server at: %s:%s\n", *host, *port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", *host, *port))
	if err != nil {
		fmt.Errorf("Failed to connect redis server:%s", err)
		return
	}
	runClient(conn)
}

// RunClient runs a session that takes user input and makes socket connection
// to server
func runClient(conn net.Conn) {
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("redis-cli> ")
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, "\n")
		// Redis server accepts RESPArray(RESPBulkString)
		parts := strings.Split(text, " ")
		commandArray := make([]string, len(parts))
		for i := 0; i < len(parts); i++ {
			part := parts[i]
			// Get length of part
			commandArray[i] = fmt.Sprintf("$%d\r\n%s\r\n", len(part), part)
		}
		cmd := fmt.Sprintf("*%d\r\n", len(commandArray)) + strings.Join(commandArray, "")
		// send to socket
		fmt.Fprintf(conn, cmd)
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print(message)
	}
}
