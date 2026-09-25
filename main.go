package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/NooRMaseR/go-redis/stores"
)

func main() {
	store := stores.NewTTLStore(time.Second)
	defer store.Close()

	lister, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Panic("Error listening:")
	}
	defer lister.Close()

	fmt.Println("Go-Redis is running on port 6379")

	for {
		conn, err := lister.Accept()
		if err != nil {
			log.Println("Error accepting connection")
		}
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store stores.IStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		conn.Write([]byte(conn.LocalAddr().String() + "> "))
		data, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading command %w", err)
			return
		}

		if len(data) == 0 {
			continue
		}

		args := ParseCommand(data)

		if len(args) == 0 {
			continue
		}
		
		cmd := strings.ToUpper(args[0])

		switch cmd {
			case "KEYS":
				conn.Write([]byte(FormatListToString(store.Keys())))
			case "GET":
				RunGetCommand(conn, store, args)
			case "SET":
				RunSetCommand(conn, store, args)
			case "DELETE":
				RunDeleteCommand(conn, store, args)
			case "RENAME":
				RunRenameCommand(conn, store, args)
			case "RENAMENX":
				RunRenameNxCommand(conn, store, args)
			case "EXIST":
				RunExistCommand(conn, store, args)
			case "POP":
				RunPopCommand(conn, store, args)
			case "LEN":
				conn.Write([]byte(strconv.Itoa(store.Len()) + "\n"))
			case "CLEAR":
				store.Clear()
				SendOk(conn)
		}
	}
}
