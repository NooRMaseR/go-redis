package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/NooRMaseR/go-redis/stores"
)

func SendOk(conn net.Conn) {
	conn.Write([]byte("OK\n"))
}

func FormatListToString(list []string) string {
	return "[ " + strings.Join(list, ", ") + " ]" + "\n"
}

func ParseCommand(command string) []string {
	command = strings.TrimSuffix(command, "\r\n")
	command = strings.TrimSuffix(command, "\n")

	return strings.Fields(command)
}

func RunGetCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 2 {
		conn.Write([]byte("unknown Get command: GET <keys>\n"))
		return
	}

	if length == 2 {
		res, err := store.Get(args[1])
		if err != nil {
			conn.Write([]byte(err.Error()))
			return
		}
		conn.Write([]byte(res + "\n"))
	} else {
		results := make([]string, length-1)

		for i, command := range args[1:] {
			res, err := store.Get(command)

			if err != nil {
				results[i] = "nil"
			} else {
				results[i] = res
			}
		}

		conn.Write([]byte(FormatListToString(results)))
	}
}

func RunSetCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 3 {
		conn.Write([]byte("expected 2 or more arguments: SET <key> <value> <expiration?>\n"))
		return
	}

	if length == 4 {
		expiration, err := strconv.Atoi(args[3])
		if err != nil {
			conn.Write([]byte("expiration must be an integer\n"))
			return
		}
		store.Set(args[1], stores.TTL{Value: args[2], Expiration: time.Now().Add(time.Millisecond * time.Duration(expiration))})
		SendOk(conn)
		return
	}
	store.Set(args[1], stores.TTL{Value: args[2]})
	SendOk(conn)
}

func RunDeleteCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 2 {
		conn.Write([]byte("unknown DELETE command: DELETE <keys>\n"))
		return
	}

	for _, key := range args[1:] {
		store.Delete(key)
	}

	SendOk(conn)
}

func RunRenameCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 3 {
		conn.Write([]byte("unknown RENAME command: RENAME <oldkey> <newkey>\n"))
		return
	}

	err := store.Rename(args[1], args[2])
	if err != nil {
		conn.Write([]byte("could not rename: " + err.Error()))
		return
	}
	SendOk(conn)
}

func RunRenameNxCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 3 {
		conn.Write([]byte("unknown RENAME command: RENAME <oldkey> <newkey>\n"))
		return
	}

	err := store.RenameNX(args[1], args[2])
	if err != nil {
		conn.Write([]byte("could not rename: " + err.Error()))
		return
	}
	SendOk(conn)
}

func RunExistCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 2 {
		conn.Write([]byte("unknown EXIST command: EXIST <keys>\n"))
		return
	}

	if length > 2 {
		results := make([]string, length-1)

		for i, key := range args[1:] {
			ok := store.Exists(key)
			if ok {
				results[i] = "true"
			} else {
				results[i] = "false"
			}
		}

		conn.Write([]byte(FormatListToString(results)))
	} else {
		ok := store.Exists(args[1])
		if ok {
			conn.Write([]byte("true\n"))
		} else {
			conn.Write([]byte("false\n"))
		}
	}
}

func RunPopCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 2 {
		conn.Write([]byte("unknown POP command: POP <keys>\n"))
		return
	}

	if length > 2 {
		results := make([]string, length-1)

		for i, key := range args[1:] {
			val, err := store.Pop(key)
			if err != nil {
				results[i] = "nil"
			} else {
				results[i] = val
			}
		}

		conn.Write([]byte(FormatListToString(results)))
	} else {
		val, err := store.Pop(args[1])
		if err != nil {
			conn.Write([]byte("could not pop: " + err.Error()))
			return
		}
		conn.Write([]byte(val + "\n"))
	}
}
