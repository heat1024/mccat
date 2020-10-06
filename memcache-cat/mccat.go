package mccat

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTTL = 3600
)

type cmds struct {
	argv        []string
	ops         options
	maxArgCount int
	getall      bool
}

type options struct {
	namespace  string
	vnamespace string
	grep       string
	vgrep      string
	keyOnly    bool
	countOnly  bool
}

// Client is a memcache client.
type Client struct {
	Conn        net.Conn
	buff        *bufio.ReadWriter
	historyFile *os.File
	historyRW   *bufio.ReadWriter
	url         string
	cmdHistory  []string
}

// Item is struct of stored data
type Item struct {
	Key   string
	Value string
}

func usage() {
	fmt.Println("Command list")
	fmt.Println("> get key [key2] [key3] ...                                             : Get data from server")
	fmt.Println("> set key ttl                                                           : Set data (overwrite when exist)")
	fmt.Println("> add key ttl                                                           : Add new data (error when key exist)")
	fmt.Println("> append key ttl                                                        : Append data from exist data")
	fmt.Println("> prepend key ttl                                                       : Prepend data from exist data")
	fmt.Println("> replace key ttl                                                       : Replace data from exist data")
	fmt.Println("> incr[increase] key number                                             : Increase numeric value")
	fmt.Println("> decr[decrease] key number                                             : Decrease numeric value")
	fmt.Println("> del[delete|rm|remove] key [key2] [key3] ...                           : Remove key item from server")
	fmt.Println("> key_counts                                                            : Get key counts")
	fmt.Println("> get_all [--name namespace] [--grep grep_words] --verbose              : Get almost all items from server (can grep by namespace or key words)")
	fmt.Println("> flush_all                                                             : Get key counts")
	fmt.Println("> help                                                                  : Show usage")
}

func getServerAddr(url string) string {
	var port int
	var err error

	addr := strings.SplitN(url, ":", 2)
	if len(addr) != 2 {
		fmt.Println("connect to default port(11211)")

		port = 11211
	} else {
		port, err = strconv.Atoi(addr[1])
		if err != nil {
			fmt.Println("connect to default port(11211)")

			port = 11211
		}
	}

	return fmt.Sprintf("%s:%d", addr[0], port)
}

// New make connection to provided address:port
// and returns a memcache client.
func New(url string, cmdHistoryFilePath string) (*Client, error) {
	url = getServerAddr(url)

	nc, err := net.Dial("tcp", fmt.Sprintf("%s", url))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to memcached server: %s", err.Error())
	}

	err = nc.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		return nil, err
	}

	err = nc.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		return nil, err
	}

	historyFile, err := os.OpenFile(cmdHistoryFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("cannot open mccat history file [%s]: %s. mccat will not store history to file\n", cmdHistoryFilePath, err.Error()))
		historyFile = nil
	}

	c := &Client{
		Conn:        nc,
		url:         url,
		historyFile: historyFile,
		historyRW:   nil,
		cmdHistory:  nil,
		buff:        bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}

	if c.historyFile != nil {
		c.historyRW = bufio.NewReadWriter(bufio.NewReader(c.historyFile), bufio.NewWriter(c.historyFile))

		for {
			buff, err := c.historyRW.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					os.Stderr.WriteString(fmt.Sprintf("got error while read cmd history: %s\n", err.Error()))

					// close fd and set nil when got error history file load
					c.historyFile.Close()
					c.historyFile = nil
					c.historyRW = nil
				}
				break
			}

			c.cmdHistory = append(c.cmdHistory, strings.TrimRight(buff, "\r\n"))
		}
	}

	return c, nil
}

// Start function is start mccat console
func (c *Client) Start() error {
	for {
		cmd := setPrompt(c.url, c.cmdHistory)

		// exit program
		if strings.HasPrefix(strings.ToLower(cmd), "exit") || strings.HasPrefix(strings.ToLower(cmd), "quit") {
			break
		}

		// parse and execute command
		cmds, err := parseCmd(cmd)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if cmds != nil {
				if err := c.Run(cmds); err != nil {
					fmt.Printf("%s\n", err.Error())
				}
			}
		}

		// append to command history when history is empty or current command not duplicate with latest
		if len(c.cmdHistory) == 0 || c.cmdHistory[len(c.cmdHistory)-1] != cmd {
			c.cmdHistory = append(c.cmdHistory, cmd)

			if c.historyFile != nil && c.historyRW != nil {
				if _, err := c.historyRW.WriteString(cmd + "\n"); err != nil {
					os.Stderr.WriteString(fmt.Sprintf("cannot write to cmd history file: %s", err.Error()))
				} else {
					if err := c.historyRW.Flush(); err != nil {
						os.Stderr.WriteString(fmt.Sprintf("cannot write to cmd history file: %s", err.Error()))
					}
				}
			}
		}
	}

	return nil
}

func calcTTL(ttl string) int {
	if ttl == "" {
		return defaultTTL
	}

	t, err := strconv.Atoi(ttl)
	if err != nil {
		fmt.Printf("ttl is wrong. use default ttl (%d)\n", defaultTTL)
		return defaultTTL
	}

	return t
}

func readValueInput() (string, error) {
	r := bufio.NewReader(os.Stdin)

	buff, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(buff, "\r\n"), nil
}

// Run execute command line
func (c *Client) Run(cmds *cmds) error {
	switch cmds.argv[0] {
	case "keycounts":
		err := c.GetAll(cmds.ops)
		if err != nil {
			return err
		}

		break
	case "get":
		if len(cmds.argv) < 2 {
			return fmt.Errorf("key must needed")
		}

		for i := 1; i < len(cmds.argv); i++ {
			item, err := c.Get(cmds.argv[i])
			if err != nil {
				fmt.Printf("%s : %s\n", cmds.argv[i], err.Error())
			} else {
				fmt.Printf("%s : %s\n", item.Key, item.Value)
			}
		}

		break
	case "getall":
		err := c.GetAll(cmds.ops)
		if err != nil {
			return err
		}

		break
	case "add", "set", "append", "prepend", "replace":
		var ttl int

		if len(cmds.argv) < 2 {
			return fmt.Errorf("key must needed")
		}

		if len(cmds.argv) < 3 {
			ttl = defaultTTL
		} else {
			ttl = calcTTL(cmds.argv[2])
		}

		value, err := readValueInput()
		if err != nil {
			return err
		}

		if err := c.Store(cmds, ttl, value); err != nil {
			return err
		}

		fmt.Printf("key %s %s complate\n", cmds.argv[1], cmds.argv[0])

		break
	case "del", "delete", "rm", "remove":
		if len(cmds.argv) < 2 {
			return fmt.Errorf("key must needed")
		}

		for i := 1; i < len(cmds.argv); i++ {
			if err := c.Del(cmds.argv[i]); err != nil {
				return err
			}

			fmt.Printf("key %s deleted\n", cmds.argv[i])
		}

		break
	case "flushall":
		if err := c.FlushAll(); err != nil {
			return err
		}

		fmt.Println("All keys deleted")

		break
	case "incr", "decr":
		if len(cmds.argv) < 2 {
			return fmt.Errorf("key must needed")
		}

		if len(cmds.argv) < 3 {
			return fmt.Errorf("numeric must needed")
		}

		res, err := c.IncrDecr(cmds)
		if err != nil {
			return err
		}

		fmt.Printf("%s: %s\n", cmds.argv[1], res)

		break
	default:
		return fmt.Errorf("wrone command: %s", cmds.argv[0])
	}

	return nil
}

// Close is close Client connection.
func (c *Client) Close() error {
	fmt.Println("exit mccat terminal")

	if c.historyFile != nil {
		if err := c.historyFile.Close(); err != nil {
			return err
		}
	}

	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
