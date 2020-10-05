package mccat

import (
	"fmt"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"
)

func setPrompt(url string, cmdHistory []string) string {
	return prompt.Input(fmt.Sprintf("%s> ", url), completerFunc,
		prompt.OptionTitle(fmt.Sprintf("mccat on %s", url)),
		prompt.OptionHistory(cmdHistory),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
}
func completerFunc(d prompt.Document) []prompt.Suggest {
	var cmd string
	current := d.GetWordBeforeCursorWithSpace()
	currentLine := d.Lines()[0]

	cmdFields := strings.Fields(currentLine)
	if len(cmdFields) > 0 {
		cmd = cmdFields[0]
	} else {
		cmd = current
	}

	var s []prompt.Suggest

	if strings.HasPrefix(currentLine, "getall ") {
		s = []prompt.Suggest{
			{Text: "getall --name", Description: "grep by namespace"},
			{Text: "getall --vname", Description: "grep except namespace"},
			{Text: "getall --grep", Description: "grep word in whole key name (ex: test -> hoge_test_moge)"},
			{Text: "getall --vgrep", Description: "grep word in whole except key name (ex: test -> hoge_moge"},
			{Text: "getall --verbose", Description: "diaplay result with value like key : value"},
		}
	} else if strings.HasPrefix(currentLine, "keycounts ") {
		s = []prompt.Suggest{
			{Text: "keycounts --name", Description: "grep by namespace"},
			{Text: "keycounts --vname", Description: "grep except namespace"},
			{Text: "keycounts --grep", Description: "grep word in whole key name (ex: test -> hoge_test_moge)"},
			{Text: "keycounts --vgrep", Description: "grep word in whole except key name (ex: test -> hoge_moge"},
			{Text: "keycounts --verbose", Description: "diaplay result with value like key : value"},
		}
	} else if strings.HasPrefix(currentLine, "set ") {
		s = []prompt.Suggest{
			{Text: "set [key] [ttl]", Description: "type key name and ttl(sec)"},
		}
	} else if strings.HasPrefix(currentLine, "add ") {
		s = []prompt.Suggest{
			{Text: "add [key] [ttl]", Description: "type key name and ttl(sec)"},
		}
	} else if strings.HasPrefix(currentLine, "append ") {
		s = []prompt.Suggest{
			{Text: "append [key] [ttl]", Description: "type key name and ttl(sec)"},
		}
	} else if strings.HasPrefix(currentLine, "prepend ") {
		s = []prompt.Suggest{
			{Text: "prepend [key] [ttl]", Description: "type key name and ttl(sec)"},
		}
	} else if strings.HasPrefix(currentLine, "replace ") {
		s = []prompt.Suggest{
			{Text: "replace [key] [ttl]", Description: "type key name and ttl(sec)"},
		}
	} else if strings.HasPrefix(currentLine, "incr ") {
		s = []prompt.Suggest{
			{Text: "incr [key] [numeric]", Description: "type key name and numeric value"},
		}
	} else if strings.HasPrefix(currentLine, "decr ") {
		s = []prompt.Suggest{
			{Text: "decr [key] [numeric]", Description: "type key name and numeric value"},
		}
	} else if strings.HasPrefix(currentLine, "del ") {
		s = []prompt.Suggest{
			{Text: "del [key]", Description: "type key name for delete"},
		}
	} else if strings.HasPrefix(currentLine, "delete ") {
		s = []prompt.Suggest{
			{Text: "delete [key]", Description: "type key name for delete"},
		}
	} else if strings.HasPrefix(currentLine, "rm ") {
		s = []prompt.Suggest{
			{Text: "rm [key]", Description: "type key name for delete"},
		}
	} else if strings.HasPrefix(currentLine, "remove ") {
		s = []prompt.Suggest{
			{Text: "remove [key]", Description: "type key name for delete"},
		}
	} else if strings.HasPrefix(currentLine, "get ") {
		s = []prompt.Suggest{
			{Text: "get [key]", Description: "type key name for get value"},
		}
	} else {
		s = []prompt.Suggest{
			{Text: "get", Description: "Get data from server"},
			{Text: "set", Description: "Set data (overwrite when exist)"},
			{Text: "add", Description: "Add new data (error when key exist)"},
			{Text: "append", Description: "Append data from exist data"},
			{Text: "prepend", Description: "Prepend data from exist data"},
			{Text: "replace", Description: "Replace data from exist data"},
			{Text: "incr", Description: "Increase numeric value"},
			{Text: "decr", Description: "Decrease numeric value"},
			{Text: "del", Description: "Remove key item from server"},
			{Text: "delete", Description: "Remove key item from server"},
			{Text: "rm", Description: "Remove key item from server"},
			{Text: "remove", Description: "Remove key item from server"},
			{Text: "keycounts", Description: "Get key counts (can grep by namespace or key words)"},
			{Text: "getall", Description: "Get all items from server (can grep by namespace or key words)"},
			{Text: "help", Description: "Show usage"},
			{Text: "exit", Description: "Terminate the mccat"},
		}
	}

	return prompt.FilterHasPrefix(s, cmd, true)
}
