package main

import (
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
	upToErr := string(data[:offset])
	lines := strings.Split(upToErr, "\n")
	return len(lines)
}

// Try to read and unmarshal the configuration file.
// If anything goes wrong just report it to the user
// and exit.
func readConfig(filePath string) (routes []mockity.Route) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Could not read " + filePath + ", exiting.")
	}
	err = json.Unmarshal(data, &routes)
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
