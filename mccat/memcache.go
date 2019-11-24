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

	"github.com/pkg/term"
)

const (
	defaultTTL = 3600
	maxBuff    = 3
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
}

// Client is a memcache client.
type Client struct {
	Conn       net.Conn
	buff       bufio.ReadWriter
	url        string
	cmdHisroty []string
}

// Item is struct of stored data
type Item struct {
	Key   string
	Value string
}

func parseCmd(cmd string) (*cmds, error) {

	c := &cmds{
		argv:        nil,
		maxArgCount: 1,
		getall:      false,
		ops: options{
			namespace:  "",
			vnamespace: "",
			grep:       "",
			vgrep:      "",
			keyOnly:    true,
		},
	}

	args := strings.Split(cmd, " ")
	maxArgs := len(args)

	c.argv = append(c.argv, strings.ToLower(args[0]))
	cmd = c.argv[0]

	switch cmd {
	case "get":
		c.maxArgCount = 2
		break
	case "allitems", "getall":
		c.maxArgCount = 10
		c.getall = true
		break
	case "set", "add", "replace", "append", "prepend":
		c.maxArgCount = 3
		break
	case "del", "delete", "rm", "remove":
		c.maxArgCount = 2
		break
	case "incr", "increase", "decr", "decrease":
		c.maxArgCount = 3
		break
	case "help":
		// show usage
		usage()
		return nil, nil
	default:
		return nil, fmt.Errorf("wrong command %s", cmd)
	}

	if maxArgs > c.maxArgCount {
		usage()
		return nil, fmt.Errorf("wrong command %s", cmd)
	}

	// parse ope command options (only one option allowed. default is usage)
	for i := 1; i < maxArgs; i++ {
		argv := args[i]

		switch argv {
		case "--name", "-n":
			if i+1 < maxArgs && c.getall {
				c.ops.namespace = args[i+1]
			} else {
				usage()
				return nil, fmt.Errorf("failed on parse command")
			}
			i++
			break
		case "--vname", "-vn":
			if i+1 < maxArgs && c.getall {
				c.ops.vnamespace = args[i+1]
			} else {
				usage()
				return nil, fmt.Errorf("failed on parse command")
			}
			i++
			break
		case "--grep", "-g":
			if i+1 < maxArgs && c.getall {
				c.ops.grep = args[i+1]
			} else {
				usage()
				return nil, fmt.Errorf("failed on parse command")
			}
			i++
			break
		case "--vgrep", "-vg":
			if i+1 < maxArgs && c.getall {
				c.ops.vgrep = args[i+1]
			} else {
				usage()
				return nil, fmt.Errorf("failed on parse command")
			}
			i++
			break
		case "--verbose", "-v":
			if c.getall {
				c.ops.keyOnly = false
			} else {
				usage()
				return nil, fmt.Errorf("failed on parse command")
			}
			break
		case "help", "h":
			// show usage
			usage()
			return nil, nil
		default:
			c.argv = append(c.argv, argv)
		}
	}

	return c, nil
}

func usage() {
	fmt.Println("Command list")
	// fmt.Println("  --name[-n] {key_name}     : get all data in memcached server only starts with key_name.")
	// fmt.Println("  --vname[-vn] {key_name}   : get all data in memcached server only starts without key_name.")
	// fmt.Println("  --grep[-g] {grep_words}   : get all data in memcached server only contain grep_words.")
	// fmt.Println("  --vgrep[-vg] {grep_words} : get all data in memcached server only not contain grep_words.")
	// fmt.Println("  --count[-c]               : only display the number of keys that meet the conditions.")
	fmt.Println("> get key [key2] [key3] ...                                             : get data from server")
	fmt.Println("> set key ttl                                                           : set data (overwrite when exist)")
	fmt.Println("> add key ttl                                                           : add new data (error when key exist)")
	fmt.Println("> append key ttl                                                        : append data from exist data")
	fmt.Println("> prepend key ttl                                                       : prepend data from exist data")
	fmt.Println("> incr[increase] key number                                             : increase numeric value")
	fmt.Println("> decr[decrease] key number                                             : decrease numeric value")
	fmt.Println("> del[delete|rm|remove] key                                             : remove key item from server")
	fmt.Println("> getall[allitems] [--name namespace] [--grep grep_words] --verbose     : get all items from server (can grep by namespace or key words)")
	fmt.Println("> help                                                                  : show usage")
}

// New make connection to provided address:port
// and returns a memcache client.
// socket read/write timeout (default: 1 hours)
func New(url string) (*Client, error) {
	nc, err := net.Dial("tcp", fmt.Sprintf("%s", url))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to memcached server: %s", err.Error())
	}

	nc.SetDeadline(time.Now().Add(time.Duration(defaultTTL) * time.Second))
	mc := &Client{
		Conn:       nc,
		url:        url,
		cmdHisroty: nil,
		buff: bufio.ReadWriter{
			Reader: bufio.NewReader(nc),
			Writer: bufio.NewWriter(nc),
		},
	}

	return mc, nil
}

// Start function is start mccat console
func Start(url string) error {
	// connect to memcached server
	fmt.Printf("connect to %s\r\n", url)

	nc, err := New(url)
	if err != nil {
		return fmt.Errorf("cannot connect to server %s", url)
	}
	defer nc.Close()

	for {
		cmd, err := nc.readCmdInput()
		if err != nil {
			return fmt.Errorf("error on input commands: %s", err.Error())
		}
		if len(nc.cmdHisroty) == 0 || nc.cmdHisroty[len(nc.cmdHisroty)-1] != cmd {
			nc.cmdHisroty = append(nc.cmdHisroty, cmd)
		}

		if strings.HasPrefix(strings.ToLower(cmd), "exit") || strings.HasPrefix(strings.ToLower(cmd), "quit") {
			break
		}

		cmds, err := parseCmd(cmd)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if cmds != nil {
				if err := nc.Run(cmds); err != nil {
					fmt.Printf("%s\r\n", err.Error())
				}
			}
		}

	}

	fmt.Printf("terminate mccat")

	return nil
}

func calcTTL(ttl string) int {
	if ttl == "" {
		return defaultTTL
	}

	t, err := strconv.Atoi(ttl)
	if err != nil {
		fmt.Printf("ttl is wrong. use default ttl (%d)\r\n", defaultTTL)
		return defaultTTL
	}

	return t
}

func refreshInputLine(url string, cmd string) {
	var blank string
	curCmdLength := len(cmd) + len(url) + 3
	for i := 0; i < curCmdLength; i++ {
		blank = blank + " "
	}
	fmt.Printf("\r%s", blank)
}

func getChar() (int, []byte, error) {
	var err error

	var numRead int
	bytes := make([]byte, maxBuff)

	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	defer func() {
		t.Restore()
		t.Close()
	}()

	numRead, err = t.Read(bytes)
	if err != nil {
		return 0, nil, fmt.Errorf("error on read input")
	}

	return numRead, bytes, err
}

func (c *Client) getCmdLine(line string) (string, error) {

	historyIndex := len(c.cmdHisroty)
	maxHistoryIndex := historyIndex - 1

inputLoop:
	for {
		refreshInputLine(c.url, line)
		fmt.Printf("\r%s> %s", c.url, line)

		rc, buff, err := getChar()
		if err != nil {
			return "", err
		}

		if rc == 3 && buff[0] == 27 && buff[1] == 91 {
			if historyIndex >= 0 {
				if len(c.cmdHisroty) > 0 {
					if buff[2] == 65 {
						// Up
						refreshInputLine(c.url, line)

						if historyIndex > 0 {
							historyIndex--
						}

						line = c.cmdHisroty[historyIndex]

						fmt.Printf("\r%s> %s", c.url, line)
					} else if buff[2] == 66 {
						// Down
						refreshInputLine(c.url, line)

						if historyIndex < maxHistoryIndex {
							historyIndex++
							line = c.cmdHisroty[historyIndex]
						} else {
							line = ""
						}

						fmt.Printf("\r%s> %s", c.url, line)
					}
				}
			}

			// if buff[2] == 68 {
			// 	// Left
			// 	fmt.Printf("left pressed\r\n")
			// } else if buff[2] == 67 {
			// 	// Right
			// 	fmt.Printf("right pressed\r\n")
			// }
		} else if rc == 1 {
			switch buff[0] {
			case 127:
				if len(line) > 0 {
					line = line[:len(line)-1]
				}
			case 13:
				fmt.Println()
				break inputLoop
			case 3:
				fmt.Println()
				line = "exit"
				break inputLoop
			default:
				line = line + string(buff[0])
			}
		}
	}

	return strings.TrimRight(line, "\r\n"), nil
}

func (c *Client) readCmdInput() (string, error) {
	var cmd string = ""

	line, err := c.getCmdLine(cmd)
	if err != nil {
		return "", err
	}

	return line, nil
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
	case "get":
		if len(cmds.argv) < 2 {
			return fmt.Errorf("key must needed")
		}

		item, err := c.Get(cmds.argv[1])
		if err != nil {
			return err
		}

		fmt.Printf("%s: %s\r\n", item.Key, item.Value)

		break
	case "allitems", "getall":
		items, err := c.GetAll(cmds.ops)
		if err != nil {
			return err
		}

		fmt.Printf("Key counts: %d\r\n", len(items))

		for _, i := range items {
			if cmds.ops.keyOnly {
				fmt.Printf("  - %s\r\n", i.Key)
			} else {
				fmt.Printf("  - %s : %s\r\n", i.Key, i.Value)
			}
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

		fmt.Printf("key %s %s complate\r\n", cmds.argv[1], cmds.argv[0])

		break
	case "del", "delete", "rm", "remove":
		if err := c.Del(cmds.argv[1]); err != nil {
			return err
		}

		fmt.Printf("key %s deleted\r\n", cmds.argv[1])

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

		fmt.Printf("%s: %s\r\n", cmds.argv[1], res)

		break
	default:
		return fmt.Errorf("wrone command: %s", cmds.argv[0])
	}

	return nil
}

// Close is close Client connection.
func (c *Client) Close() error {
	return c.Conn.Close()
}

// Write command to memcached server
func (c *Client) Write(cmd string) error {
	res := error(nil)

	// set LF end of cmd line
	cmd = strings.TrimRight(cmd, "\r\n") + "\r\n"

	_, err := c.buff.Writer.WriteString(cmd)
	if err != nil {
		res = fmt.Errorf("failed on sending command to memcached server: %s", err.Error())

	} else {
		if c.buff.Writer.Flush() != nil {
			res = fmt.Errorf("failed on sending command to memcached server: %s", err.Error())
		}
	}

	return res
}

// Read response and trim out CRLF
func (c *Client) Read() (string, error) {
	buff, err := c.buff.Reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
	}

	return strings.TrimRight(buff, "\r\n"), nil
}

func (c *Client) getSlabData() ([]int, error) {
	var slabIDs []int

	err := c.Write("stats items")
	if err != nil {
		return nil, err

	}

	slab := make(map[int]bool)

	for {
		buff, err := c.Read()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
		}

		if strings.HasPrefix(buff, "END") {
			break
		}
		if strings.Contains(buff, "ERROR") {
			return nil, fmt.Errorf("got error on reading response from memcached server: %s", err.Error())
		}
		if strings.HasPrefix(buff, "STAT") {
			s := strings.Split(buff, ":")
			if slabID, err := strconv.Atoi(s[1]); err != nil {
				os.Stderr.WriteString(fmt.Sprintf("got error on parse slave ID: %s", err.Error()))
				os.Exit(1)
			} else {
				if _, exists := slab[slabID]; !exists {
					slab[slabID] = true
					slabIDs = append(slabIDs, slabID)
				}
			}
		}
	}

	return slabIDs, nil
}

// Get search data by key and return by Item struct
func (c *Client) Get(key string) (*Item, error) {
	var value string

	err := c.Write(fmt.Sprintf("get %s", key))
	if err != nil {
		return nil, err
	}

	for {
		buff, err := c.Read()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
		}

		if buff == "END" {
			break
		}
		if strings.Contains(buff, "ERROR") {
			return nil, fmt.Errorf("got error on get data of key [%s] from memcached server", key)
		}
		if !strings.HasPrefix(buff, "VALUE") {
			value = buff
		}
	}

	if len(value) == 0 {
		return nil, fmt.Errorf("no values")
	}

	return &Item{Key: key, Value: value}, nil
}

// Store function stores key / value to memcached server by each commands
func (c *Client) Store(cmds *cmds, ttl int, value string) error {
	cmd := cmds.argv[0]
	key := cmds.argv[1]
	size := len(value)

	err := c.Write(fmt.Sprintf("%s %s 0 %d %d\r\n%s", cmd, key, ttl, size, value))
	if err != nil {
		return err
	}

	buff, err := c.Read()
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
	}

	if strings.HasPrefix(buff, "NOT_STORED") {
		if cmd == "add" {
			return fmt.Errorf("failed to %s: key exist", cmd)
		}

		return fmt.Errorf("failed to %s: key does not exist", cmd)

	}

	if strings.Contains(buff, "ERROR") {
		return fmt.Errorf("got error on %s value to memcached server", cmd)
	}

	return nil
}

// Del function delete data by key from memcached server
func (c *Client) Del(key string) error {
	err := c.Write(fmt.Sprintf("delete %s", key))
	if err != nil {
		return err
	}

	buff, err := c.Read()
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
	}

	if strings.HasPrefix(buff, "NOT_FOUND") {
		return fmt.Errorf("key %s not found", key)
	}

	if strings.Contains(buff, "ERROR") {
		return fmt.Errorf("got error on delete key %s from memcached server", key)
	}

	return nil
}

// IncrDecr function increment or decrement numeric data
func (c *Client) IncrDecr(cmds *cmds) (string, error) {
	cmd := cmds.argv[0]
	key := cmds.argv[1]
	value := cmds.argv[2]

	err := c.Write(fmt.Sprintf("%s %s %s", cmd, key, value))
	if err != nil {
		return "", err
	}

	buff, err := c.Read()
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
	}

	if strings.HasPrefix(buff, "NOT_FOUND") {
		return "", fmt.Errorf("key %s not found", key)
	}
	if strings.Contains(buff, "ERROR") {
		return "", fmt.Errorf("cannot increment or decrement non-numeric value")
	}

	return buff, nil
}

// GetAll return all key/value data in memcached server
func (c *Client) GetAll(ops options) ([]*Item, error) {
	var keys []string
	var matchName bool
	var matchGrep bool

	SlabIDs, err := c.getSlabData()
	if err != nil {
		return nil, fmt.Errorf("cannot get slab data from memcached server: %s", err.Error())
	}

	for _, slabID := range SlabIDs {
		err := c.Write(fmt.Sprintf("stats cachedump %d 0", slabID))
		if err != nil {
			return nil, err
		}

		for {
			buff, err := c.Read()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
			}

			if strings.HasPrefix(buff, "END") {
				break
			}
			if strings.Contains(buff, "ERROR") {
				return nil, fmt.Errorf("got error on reading response from memcached server: %s", err.Error())
			}
			if strings.HasPrefix(buff, "ITEM") {
				key := strings.Split(buff, " ")[1]
				ns := strings.SplitN(key, ":", 2)[0]

				matchName = true
				matchGrep = true

				// if namespace defined, compare with namespace
				if len(ops.namespace) > 0 && ns != ops.namespace {
					matchName = false
				}
				// if vnamespace defined, compare with namespace
				if len(ops.vnamespace) > 0 && ns == ops.vnamespace {
					matchName = false
				}
				// if grep word defined, check about key contains words
				if len(ops.grep) > 0 && !strings.Contains(key, ops.grep) {
					matchGrep = false
				}
				// if vgrep word defined, check about key contains words
				if len(ops.vgrep) > 0 && strings.Contains(key, ops.vgrep) {
					matchGrep = false
				}
				if matchName && matchGrep {
					keys = append(keys, key)
				}
			}
		}
	}

	var result []*Item

	for _, key := range keys {
		if ops.keyOnly {
			result = append(result, &Item{Key: key})
		} else {
			item, err := c.Get(key)
			if err == nil {
				result = append(result, item)
			}
		}
	}

	return result, nil
}
