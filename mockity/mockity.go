package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wilig/mockity"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var configFile = flag.String("conf", "mockity.conf",
	"Location of the JSON configuration file")

var port = flag.Int("port", 8989, "Port on which Mockity will listen")

func findLine(data []byte, offset int64) int {
	upToErr := data[:offset]
	lines := bytes.Count(upToErr, []byte("\n"))
	return lines + 1
}

// Try to read and unmarshal the configuration file.
// If anything goes wrong just report it to the user
// and exit.
func readConfig(filePath string) (routes []mockity.Route) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Could not read " + filePath + ", exiting.")
	}
	jsonable := preProcess(data)
	err = json.Unmarshal(jsonable, &routes)
	if err != nil {
		if synerr, ok := err.(*json.SyntaxError); ok {
			line := findLine(data, synerr.Offset)
			fmt.Fprintf(os.Stderr, "Syntax error in %s\nLine: %d - %s\n",
				filePath, line, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Error parsing %s\n%s\n", filePath, err.Error())
		}
		os.Exit(-1)
	}
	return
}

func main() {
	flag.Parse()
	routes := readConfig(*configFile)
	http.HandleFunc("/", mockity.MakeMockery(routes))
	fmt.Printf("Mockity is awaiting you on port %d\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// preProcess removes comments and collapses multiline string values 
// to prepare the data for JSON unmarshalling.  It's important to
// preserve the character count as errors are reported via offset.  
// Thus we just replace all preprocessed characters with whitespace.
// 
// Not sure this the best approach perhaps writing a custom
// unMarshaler with a custom scanner would be more flexible.  If
// this gets any more complex will have to look at alternatives.
func preProcess(data []byte) []byte {
	var inStr bool
	var buffer bytes.Buffer
	reader := bytes.NewReader(data)
	c, err := reader.ReadByte()
	for err == nil {
		switch {
		case c == '\n':
			if inStr {
				c = ' '
			}
			buffer.WriteByte(c)
		case c == '\t':
			if inStr {
				c = ' '
			}
			buffer.WriteByte(c)
		case c == '"':
			if inStr {
				inStr = false
			} else {
				inStr = true
			}
			buffer.WriteByte(c)
		case c == '\\':
			buffer.WriteByte(c)
			escChar, _ := reader.ReadByte()
			buffer.WriteByte(escChar)
		case c == '/': // Start of comment?
			nc, err := reader.ReadByte()
			if err == nil {
				if !inStr && nc == '/' && err == nil {
					i := 2
					for c != '\n' && err == nil {
						i += 1
						c, err = reader.ReadByte()
					}
					buffer.WriteString(strings.Repeat(" ", i+1))
				} else {
					buffer.WriteByte(c)
					buffer.WriteByte(nc)
				}
			}
		default:
			buffer.WriteByte(c)
		}
		c, err = reader.ReadByte()
	}
	return buffer.Bytes()
}
