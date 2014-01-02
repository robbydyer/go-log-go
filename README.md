# Go Log Go
Go-log-go is a File watcher/parser. It is essentially just a dumbed-down grep, with a daemon mode. You can have it scan a file for either a string, or a regular expression.

    Usage of golog:
      -daemon=false: Mark as true to run in daemon mode
      -debug=false: Debug mode
      -file="testlog.txt": Name of file to parse
      -is_regex=false: Mark as true if your parse string is a regular expression
      -max_threads=2: Max number of concurrent parse threads
      -query="Hello World": String or Regular Expression to parse for

## Example
    golog -file="testlog.txt" -query="Hello World"
        Total lines processed: 1000
        Total matches: 4
