package mccat

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func convertTOHumanDigitNumber(num uint64) string {
	strNum := strconv.FormatUint(num, 10)
	numOfDigits := len(strNum)

	numOfCommas := (numOfDigits - 1) / 3

	outPut := make([]byte, len(strNum)+numOfCommas)

	for i, j, k := len(strNum)-1, len(outPut)-1, 0; ; i, j = i-1, j-1 {
		outPut[j] = strNum[i]
		if i == 0 {
			return string(outPut)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			outPut[j] = ','
		}
	}
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
		return nil, fmt.Errorf("got error! (cache missed)")
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
func (c *Client) GetAll(ops options) error {
	var allKeys, keys []string
	var SlabIDs []int
	var keyCounts uint64
	var err error

	SlabIDs, keyCounts, err = c.getSlabDataAndKeyCount()
	if err != nil {
		return fmt.Errorf("cannot get slab data from memcached server: %s", err.Error())
	}

	fmt.Printf("Key counts: %s\n", convertTOHumanDigitNumber(keyCounts))

	if !ops.countOnly {
		allKeys, err = c.getKeyListFromCachedump(SlabIDs)
		if err != nil {
			return err
		}

		keys = checkKeyMatch(allKeys, ops)

		for _, key := range keys {
			if ops.keyOnly {
				fmt.Printf("  - %s\n", key)
			} else {
				item, err := c.Get(key)
				if err != nil {
					fmt.Printf("  - %s : %s\n", key, err.Error())
				} else {
					fmt.Printf("  - %s : %s\n", item.Key, item.Value)
				}
			}
		}
	}

	return nil
}

// FlushAll delete all exist keys
func (c *Client) FlushAll() error {
	err := c.Write("flush_all")
	if err != nil {
		return err
	}

	buff, err := c.Read()
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
	}

	if strings.Contains(buff, "ERROR") {
		return fmt.Errorf("got error on flush all keys from memcached server")
	}

	return nil
}

func checkKeyMatch(keys []string, ops options) []string {
	var matchKey []string
	var matchName bool
	var matchGrep bool

	for _, key := range keys {
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
			matchKey = append(matchKey, key)
		}
	}

	return matchKey
}

func (c *Client) getSlabDataAndKeyCount() ([]int, uint64, error) {
	var slabIDs []int
	keyCounts := uint64(0)

	err := c.Write("stats items")
	if err != nil {
		return nil, keyCounts, err

	}

	for {
		buff, err := c.Read()
		if err != nil && err != io.EOF {
			return nil, keyCounts, fmt.Errorf("failed on reading response from memcached server: %s", err.Error())
		}

		if strings.HasPrefix(buff, "END") {
			break
		}
		if strings.Contains(buff, "ERROR") {
			return nil, keyCounts, fmt.Errorf("got error on reading response from memcached server")
		}
		if strings.HasPrefix(buff, "STAT") {
			s := strings.Split(buff, ":")
			if slabID, err := strconv.Atoi(s[1]); err != nil {
				os.Stderr.WriteString(fmt.Sprintf("got error on parse slave ID: %s", err.Error()))
				os.Exit(1)
			} else {
				numberString := fmt.Sprintf("STAT items:%d:number ", slabID)
				if strings.HasPrefix(buff, numberString) {
					slabIDs = append(slabIDs, slabID)

					f := strings.Fields(buff)
					count, err := strconv.ParseUint(f[2], 10, 64)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("got error on get slab %d's object count %s: %s", slabID, f[2], err.Error()))
						continue
					} else {
						keyCounts += count
					}
				}
			}
		}
	}

	return slabIDs, keyCounts, nil
}

func (c *Client) getKeyListFromCachedump(SlabIDs []int) ([]string, error) {
	var keys []string

	for _, slab := range SlabIDs {
		err := c.Write(fmt.Sprintf("stats cachedump %d 0", slab))
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
				return nil, fmt.Errorf("got error on reading response from memcached server")
			}
			if strings.HasPrefix(buff, "ITEM") {
				key := strings.Split(buff, " ")[1]
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}
