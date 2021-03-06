package main

import (
    "bufio"
    "flag"
    "fmt"
    "net"
    "os"
    "strings"
    "time"
)

var listenAddress = flag.String("bind", ":7777", "the address for listening for log queries")
var hostsList = flag.String("machines", "127.0.0.1:7777", "comma seperated list of addresses of other hosts with logs")
var batch = flag.Bool("batch", false, "set to true to disable the prompt (but still listen for queries")
var logFile = flag.String("logs", "machine.log", "set to true to disable the prompt but still listen for queries")

func runListener(quit chan int) {
    defer func() { quit <- 1 }() // Signal this listener has finished
    if len(*listenAddress) != 0 {
        listener, err := net.Listen("tcp", *listenAddress)
        if err != nil {
            fmt.Println("failed to listen: ", err)
        } else {
            fmt.Println("starting listener!")
            ListenForQueries(listener, *logFile)
        }
    }
}

// Just a log with origin information
type HostLog struct {
    host string
    log *Log
}

func runRequest(host string, query string, output chan *HostLog) {
    defer func() { output <- nil }() // Signal this request has finished

    conn, err := net.Dial("tcp", host)
    if err != nil {
        fmt.Printf("failed to dial %v: %v\n", host, err)
        return
    }

    req, err := NewRequest(conn, query)
    if err != nil {
        fmt.Printf("failed to start request for %v: %v\n", host, err)
        return
    }

    log, err := req.NextLog()
    for err == nil {
        output <- &HostLog{host, log}
        log, err = req.NextLog()
    }
}

func runPrompt(quit chan int) {
    defer func() { quit <- 1 }() // Signal this prompt has finished

    if *batch {
        return
    }

    hosts := strings.Split(*hostsList, ",")
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

        queryStartTime := time.Now()

        requestOutput := make(chan *HostLog)
        aliveRequests := 0
        for _,host := range hosts {
            aliveRequests++
            go runRequest(strings.TrimSpace(host), query, requestOutput)
        }

        for aliveRequests > 0 {
            log := <-requestOutput
            if log == nil {
                aliveRequests--
            } else {
                fmt.Printf("%v: %v\n", log.host, log.log.Message)
            }
        }

        fmt.Println("query finished; took", time.Since(queryStartTime))
    }
}

func main() {
    // The most important part of the program...
    fmt.Println("LogProUltraPrime 824633720831")
    fmt.Println("Copyright 2013 SickNasty Productions LLC\n")

    flag.Parse()

    quit := make(chan int)

    // Run the listener and prompt concurrently
    go runListener(quit)
    go runPrompt(quit)

    // Wait for both to exit
    <-quit
    <-quit
}