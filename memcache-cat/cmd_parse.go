package mccat

import (
	"fmt"
	"strings"
)

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
			countOnly:  false,
		},
	}

	args := strings.Split(cmd, " ")
	maxArgs := len(args)

	cmd = strings.ToLower(args[0])

	switch cmd {
	case "get":
		c.maxArgCount = 0
		break
	case "getall", "get_all":
		c.maxArgCount = 10
		c.getall = true
		cmd = "getall"
		break
	case "keycounts", "key_counts":
		c.maxArgCount = 1
		c.ops.countOnly = true
		cmd = "keycounts"
		break
	case "flushall", "flush_all", "flush":
		c.maxArgCount = 1
		cmd = "flushall"
	case "set", "add", "replace", "append", "prepend":
		c.maxArgCount = 3
		break
	case "del", "delete", "rm", "remove":
		c.maxArgCount = 0
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

	c.argv = append(c.argv, cmd)

	if c.maxArgCount > 0 {
		if maxArgs > c.maxArgCount {
			usage()
			return nil, fmt.Errorf("wrong command %s", cmd)
		}
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
