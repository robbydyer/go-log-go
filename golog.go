package main

import (
    "fmt"
    "time"
    "bufio"
    "regexp"
    "strings"
    "os"
    "flag"
)

// Globals
var default_max_parsers int = 2 // Default max concurrent reader routines
var line_buffer int = 1000 // Number of lines to buffer before analyzing
var wait_between int = 5 // Time to wait between re-init'ing the scanner, in seconds
var total_matches int = 0 // Total number of matched lines
var debug = flag.Bool("debug", false, "Debug mode")

type Golog struct {
    Filename    string
    Query       string
    Regex       *regexp.Regexp
    Max_parsers int
    Daemon      bool
}

func NewGolog (filename string, query string, max_parsers int, is_regex bool, daemon bool) {
    gl := new(Golog)
    gl.Filename = filename
    gl.Query = query
    gl.Daemon = daemon

    if (max_parsers < 1) {
        gl.Max_parsers = default_max_parsers
    } else {
        gl.Max_parsers = max_parsers
    }

    // Create regex
    if is_regex {
        gl.Regex, _ = regexp.Compile(query)
    }

    active_readers := 0
    var line_num int64
    line_num = 1
    data := make(map[int64]string)
    reader_chan := make(chan int) 

    file, err := os.Open(gl.Filename)

    defer file.Close()
    if err != nil {
        fmt.Println("Error opening file")
    }
    scanner := bufio.NewScanner(file)
    for {
        for scanner.Scan() {
            data[line_num] = scanner.Text()
            line_num++

            // Check for line_buffer
            if len(data) >= line_buffer {
                // Spin off reader routine
                gl.flush(data, &active_readers, reader_chan)

                // Empty the data
                data = make(map[int64]string)
            }
        }

        // Flush remaining lines that are under buffer
        if len(data) > 0 {
            gl.flush(data, &active_readers, reader_chan)

            // Empty the data
            data = make(map[int64]string)
        }

        if gl.Daemon {
            // Re-init Scanner
            time.Sleep(5 * time.Second)
            scanner = bufio.NewScanner(file)
        } else {
            break
        }
    }

    
    // Wait for all readers to finish
    for ; active_readers > 0 ; {
        // active_readers is decremented from parse() routine
        <- reader_chan 
    }

    fmt.Println("Total lines processed:", line_num)

}

func (gl *Golog) flush (data map[int64]string, active_readers *int, reader_chan chan int) {
    for {
        if *active_readers < gl.Max_parsers {
            *active_readers++
            go gl.parse(data, active_readers, reader_chan)
            return
        } else {
            // Wait for a thread slot to open up
            <- reader_chan
        }
    }
}

func (gl *Golog) parse (data map[int64]string, active_readers *int, reader_chan chan int) {
    var is_match bool
    for line_num, line_data := range data {
        is_match = false
        // Check if line matches query
        if gl.Regex != nil {
            is_match = gl.Regex.MatchString(line_data)
        } else if strings.Contains(line_data, gl.Query) {
            is_match = true
        }

        if is_match {
            if *debug {
                fmt.Println("MATCH Line:", line_num)
            }
            total_matches++
        } else {
            //fmt.Println("NOMATCH Line:", line_num)
        }
    }
    *active_readers--
    reader_chan <- 1
}

func main() {
    var filename = flag.String("file", "testlog.txt", "Name of file to parse")
    var query = flag.String("query", "Hello World", "String or Regular Expression to parse for")
    var max_threads = flag.Int("max_threads", default_max_parsers, "Max number of concurrent parse threads")
    var is_regex = flag.Bool("is_regex", false, "Mark as true if your parse string is a regular expression")
    var daemon = flag.Bool("daemon", false, "Mark as true to run in daemon mode")

    flag.Parse()

    NewGolog( *filename, *query, *max_threads, *is_regex, *daemon)
    fmt.Println("Total matches:", total_matches)
}
