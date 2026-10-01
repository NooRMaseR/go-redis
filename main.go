package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/NooRMaseR/go-redis/stores"
	"github.com/NooRMaseR/go-redis/utils"
)

func main() {
	noTTL := flag.Bool("no-ttl", false, "Disable TTL expiration and run as a persistent store")
	duration := flag.Duration("duration", time.Second, "Janitor sweep interval (e.g. 500ms, 2s, 1m)")
	port := flag.String("port", "6379", "Port to listen on")

	flag.Usage = func() {
		fmt.Println("Go-Redis is a simple tool for in-memory store Redis like, built by NooRMaseR")
		fmt.Println("Usage of Go-Redis:")
		flag.PrintDefaults()
	}

	flag.Parse()

	var (
		store stores.IStore
		adr   string = ":" + *port
	)

	if *noTTL {
		store = stores.NewStore()
	} else {
		store = stores.NewTTLStore(*duration)
	}

	defer store.Close()

	lister, err := net.Listen("tcp", adr)
	if err != nil {
		log.Fatalf("Error listening: %v\n", err)
	}
	defer lister.Close()

	log.Printf("Go-Redis is running on port %v\n", adr)

	for {
		conn, err := lister.Accept()
		if err != nil {
			log.Fatal("Error accepting connection")
			continue
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
			log.Printf("Error reading command %v\n", err)
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
			conn.Write([]byte(utils.FormatListToString(store.Keys())))
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
		case "RPUSH":
			RunRPushCommand(conn, store, args)
		case "LRANGE":
			RunLRangeCommand(conn, store, args)
		case "HGET":
			RunHGetCommand(conn, store, args)
		case "HSET":
			RunHSetCommand(conn, store, args)
		case "EXPIRE":
			RunExpireCommand(conn, store, args)
		case "TTL":
			RunTTLCommand(conn, store, args)
		case "LEN":
			conn.Write([]byte(strconv.Itoa(store.Len()) + "\n"))
		case "CLEAR":
			store.Clear()
			SendOk(conn)
		case "HELP":
			helpText := "Available Commands:\n" +
				"  KEYS                               - Get all active keys\n" +
				"  GET <keys...>                      - Get key value(s)\n" +
				"  SET <key> <val> [ttl]              - Set key with optional TTL\n" +
				"  DELETE <keys...>                   - Delete keys\n" +
				"  RENAME <oldKey> <newKey>           - Rename a key with a new key name (Danger: no checking if the new Key exist and can override data)\n" +
				"  RENAMENX <oldKey> <newKey>         - Rename a key with a new key name with checking if the new Key exists first or not\n" +
				"  EXIST <keys...>                    - Check if the keys are exists or not\n" +
				"  EXPIRE <key> <ttl>                 - Change the expiration of a key\n" +
				"  POP <keys...>                      - Get and delete a key's value at the same time\n" +
				"  LEN                                - Count of active keys\n" +
				"  RPUSH                              - Adds an element to the right of an array, if it doesn't exists it creates a new array\n" +
				"  LRANGE <start> <stop>              - Gets elements starting from the left of an array (index 0)\n" +
				"  HGet <key>                         - Gets the HashMap formated as a string\n" +
				"  HSET <key> <keyName> <value>       - Sets keyName value to a HashMap\n" +
				"  CLEAR                              - Flush all keys\n" +
				"  EXIST                              - Exit the server\n"
			conn.Write([]byte(helpText))
		case "EXIT":
			conn.Write([]byte("Bye!\n"))
			conn.Close()
			return
		default:
			conn.Write([]byte("-ERR unknown command '" + cmd + "', try HELP command for help\n"))
		}
	}
}
