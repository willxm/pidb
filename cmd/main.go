package main

import (
	"flag"
	"log"
	"strings"

	"github.com/tidwall/redcon"
	"github.com/willxm/pidb"
)

var (
	addr   string
	engine string
)

func init() {
	flag.StringVar(&addr, "addr", ":6380", "server address")
	flag.StringVar(&engine, "engine", "bptree", "engine type")
}

func main() {
	flag.Parse()

	var e pidb.Engine
	switch engine {
	case "mem":
		e = &pidb.Mem{}
	case "lsm_tree":
		//TODO: implement lsm_tree
	case "bptree":
		e = &pidb.BpTree{}
	default:
		log.Fatalf("unknown engine type: %s", engine)
	}

	log.Printf("pidb using engine: %s", engine)

	if err := e.New(); err != nil {
		log.Fatal(err)
	}

	var ps redcon.PubSub
	go log.Printf("started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				e.Set(string(cmd.Args[1]), cmd.Args[2])
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				val, err := e.Get(string(cmd.Args[1]))
				if err != nil {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				ok, _ := e.Exists(string(cmd.Args[1]))
				e.Del(string(cmd.Args[1]))
				if !ok {
					conn.WriteInt(0)
				} else {
					conn.WriteInt(1)
				}
			case "publish":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				conn.WriteInt(ps.Publish(string(cmd.Args[1]), string(cmd.Args[2])))
			case "subscribe", "psubscribe":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				command := strings.ToLower(string(cmd.Args[0]))
				for i := 1; i < len(cmd.Args); i++ {
					if command == "psubscribe" {
						ps.Psubscribe(conn, string(cmd.Args[i]))
					} else {
						ps.Subscribe(conn, string(cmd.Args[i]))
					}
				}
			}
		},
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
