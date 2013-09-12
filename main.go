package main

import (
    "bufio"
    "flag"
    "fmt"
    "net"
    "os"
    "strings"
)

var listenAddress = flag.String("bind", ":7777", "the address for listening for log queries")
var hostsList = flag.String("machines", "127.0.0.1:7777", "comma seperated list of addresses of other hosts with logs")
var batch = flag.Bool("batch", false, "set to true to disable the prompt (but still listen for queries")
var logFile = flag.String("logs", "machine.log", "set to true to disable the prompt but still listen for queries")

func runListener(quit chan int) {
    if len(*listenAddress) != 0 {
        listener, err := net.Listen("tcp", *listenAddress)
        if err != nil {
            fmt.Println("failed to listen: ", err)
        } else {
            fmt.Println("starting listener!")
            ListenForQueries(listener, *logFile)
        }
    }
    quit <- 1
}

func runPrompt(quit chan int) {
    promptReader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("> ")
        line, err := promptReader.ReadString('\n')
        if err != nil {
            fmt.Println("error reading prompt: ", err)
            quit <- 1
            return
        }

        query := strings.TrimSuffix(line, "\n")
        if len(query) == 0 {
            continue
        }

        conn, err := net.Dial("tcp", *hostsList)
        req, err := NewRequest(conn, query)
        log, err := req.NextLog()
        for err == nil {
            fmt.Println(log.Message)
            log, err = req.NextLog()
        }

    }
}

func main() {
    flag.Parse()

    // The most important part of the program...
    fmt.Println("LogProUltraPrime 0.5 (RC1)")
    fmt.Println("Copyright 2013 SickNasty Productions LLC.\n")

    quit := make(chan int)
    go runListener(quit)
    go runPrompt(quit)
    <-quit
}