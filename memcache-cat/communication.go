package mccat

import (
	"fmt"
	"io"
	"strings"
)

// Write command to memcached server
func (c *Client) Write(cmd string) error {
	res := error(nil)

	// set CRLF end of cmd line (memcached recommanded)
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
