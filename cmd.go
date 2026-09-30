package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/NooRMaseR/go-redis/stores"
	"github.com/NooRMaseR/go-redis/utils"
)

func SendOk(conn net.Conn) {
	conn.Write([]byte("OK\n"))
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

		conn.Write([]byte(utils.FormatListToString(results)))
	}
}

func RunSetCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 3 {
		conn.Write([]byte("expected 2 or more arguments: SET <key> <value> <expiration?>\n"))
		return
	}

	if length == 4 {
		expiration, err := time.ParseDuration(args[3])
		if err != nil {
			conn.Write([]byte(stores.ErrInvalidDuration.Error()))
			return
		}
		store.Set(args[1], args[2], expiration)
		SendOk(conn)
		return
	}
	store.Set(args[1], args[2], 0)
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

		conn.Write([]byte(utils.FormatListToString(results)))
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

		conn.Write([]byte(utils.FormatListToString(results)))
	} else {
		val, err := store.Pop(args[1])
		if err != nil {
			conn.Write([]byte("could not pop: " + err.Error()))
			return
		}
		conn.Write([]byte(val + "\n"))
	}
}

func RunRPushCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length < 3 {
		conn.Write([]byte("unknown RPUSH command: RPUSH <key> <values...>\n"))
		return
	}

	err := store.RPush(args[1], args[2:])
	if err != nil {
		conn.Write([]byte("could not push: " + err.Error()))
		return
	}
	SendOk(conn)
}

func RunLRangeCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 4 {
		conn.Write([]byte("unknown LRANGE command: LRANGE <key> <start> <stop>\n"))
		return
	}

	start, err := strconv.Atoi(args[2])
	if err != nil {
		conn.Write([]byte("start number error: please enter a valid number"))
		return
	}

	stop, err := strconv.Atoi(args[3])
	if err != nil {
		conn.Write([]byte("stop number error: please enter a valid number"))
		return
	}

	values, err := store.LRange(args[1], start, stop)
	if err != nil {
		conn.Write([]byte("could not get range: " + err.Error()))
		return
	}
	conn.Write([]byte(utils.FormatListToString(values)))
}

func RunHGetCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 2 {
		conn.Write([]byte("unknown HGET command, expected HGET <key>\n"))
		return
	}
	val, err := store.HGet(args[1])
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	conn.Write([]byte(utils.FormatMapToString(val)))
}

func RunHSetCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 4 {
		conn.Write([]byte("unknown HGET command, expected HSET <key> <keyName> <value>\n"))
		return
	}
	err := store.HSet(args[1], args[2], args[3])
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	SendOk(conn)
}

func RunExpireCommand(conn net.Conn, store stores.IStore, args []string) {
	var (
		length int = len(args)
		dur    time.Duration
		err    error
	)

	if length != 3 {
		conn.Write([]byte("unknown EXPIRE command: expected EXPIRE <key> <ttl>\n"))
		return
	}

	dur, err = time.ParseDuration(args[2])
	if err != nil {
		conn.Write([]byte("unknown ttl number\n"))
		return
	}

	err = store.Expire(args[1], dur)
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}

	SendOk(conn)
}

func RunTTLCommand(conn net.Conn, store stores.IStore, args []string) {
	length := len(args)
	if length != 2 {
		conn.Write([]byte("unknown TTL command: expected TTL <key>\n"))
		return
	}
	
	obj, err := store.OGet(args[1])
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	
	remaining := time.Until(obj.Expiration)
	conn.Write([]byte(strconv.Itoa(int(remaining.Milliseconds())) + "\n"))
}
