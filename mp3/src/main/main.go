package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "net/rpc"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"

    "mykv"
    "membertable"
)

var listenAddress = flag.String("bind", ":7777", "the address for listening to services")
var seedAddress = flag.String("seed", "", "the address of some machine to grab the inital membertable from")
var machineName = flag.String("name", "", "the name of this machine")
var logFile = flag.String("logs", "machine.log", "the file name to store the log in")
var command = flag.String("run", "", "command to run")
var interactive = flag.Bool("interactive", false, "set to true to run interactively; cancels running a node and run command")

func makeGraph() *mykv.KVGraph {
    return nil
}

type commandDispatch struct {
    RE *regexp.Regexp
    Handler func([]string, *mykv.KVGraph) bool
}

func handleInsert(params []string, g *mykv.KVGraph) bool {
    if len(params) != 3 {
        log.Println("not enough params")
        return true
    }
    keyUint, err := strconv.ParseUint(params[1], 10, 32)
    if err != nil {
        log.Println("invalid integer key")
        return true
    }
    kv := mykv.KeyValue{ mykv.Key(keyUint), params[2] }
    if err := g.Insert(kv); err != nil {
        log.Println("insert error: ", err)
        return true
    }
    return true
}

var commandRE = []commandDispatch{
    { regexp.MustCompile(`insert\s+(\S+)\s+(.*)`), handleInsert },
    // { regexp.MustCompile(`insert\s+(\S+)\s+(.*)`), handleUpdate },
    // { regexp.MustCompile(`insert\s+(\S+)\s+(.*)`), handleLookup },
    // { regexp.MustCompile(`insert\s+(\S+)\s+(.*)`), handleDelete },
}

func runCommand(cmd string, g *mykv.KVGraph) bool {
    for _, command := range commandRE {
        matches := command.RE.FindStringSubmatch(cmd)
        if matches != nil {
            return command.Handler(matches, g)
        }
    }
    log.Println("invalid command")
    return true
}

func runInteractive(g *mykv.KVGraph) {
    if *seedAddress == "" {
        log.Println("must have machine to connect to in interactive mode")
        //return
    }

    promptReader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("> ")
        line, err := promptReader.ReadString('\n')
        if err != nil {
            log.Println("error reading prompt: ", err)
            break
        }

        cmd := strings.TrimSuffix(line, "\n")
        if len(cmd) == 0 {
            continue
        }

        queryStartTime := time.Now()

        if !runCommand(cmd, g) {
            break
        }

        fmt.Println("cmd finished; took", time.Since(queryStartTime))
    }
}

func getIP(hostname string) string {
    machineIP, err := net.InterfaceAddrs()
    if err != nil || len(machineIP) == 0 {
        return ""
    }

    var preferredIP net.IP
    for _, ipStr := range machineIP {
        var ip net.IP

        // We need to read the ip address which can be nested within different data structures.
        switch ipSource := ipStr.(type) {
        case *net.IPNet:
            ip = ipSource.IP
        case *net.IPAddr:
            ip = ipSource.IP
        default:
            continue
        }

        // Prefer IPv4 addresses that come sooner in the list and are not local of LookupHost
        if preferredIP == nil || (preferredIP.To4() == nil && ip.To4() != nil) || (preferredIP[12] == 127) {
            preferredIP = ip
        }
    }

    return preferredIP.String()
}

// Choose a color for a given ID
func getColor(id int32) string {
    switch id % 6 {
        case 0: return "1;31";
        case 1: return "1;32";
        case 2: return "1;34";
        case 3: return "1;33";
        case 4: return "1;35";
        case 5: return "1;36";
    }
    return "0";
}

func runServer() {
    // Get the machines name
    hostname, _ := os.Hostname()

    name := *machineName
    if name == "" {
        name = hostname
    }

    // Get the address that this machine can be contacted from if none was given
    bindAddress, bindPort, err := net.SplitHostPort(*listenAddress)
    if bindAddress == "" {
        bindAddress = getIP(hostname)
    }

    id, idErr := membertable.IncrementIDFile(bindAddress + "_" + bindPort + ".bin")
    if idErr != nil {
        log.Println("Error retriving id number")
        id = 0
    }

    // Add ourselves to the table
    myID := membertable.ID{
        Num: id,
        Name: *machineName,
        Address: bindAddress + ":" + bindPort,
    }

    // If no name was given, default to the host name
    if myID.Name == "" {
        myID.Name = hostname
    }

    var t membertable.Table
    t.Init(myID)

    addr := bindAddress + ":" + bindPort

    // Configure the log file to be something nice
    log.SetPrefix("[\x1B[" + getColor(3) + "m" + name + "\x1B[0m]:")

    logfd, err := os.Create(*logFile)

    if err != nil {
        log.Println(err)
    }

    log.SetOutput(io.MultiWriter(logfd, os.Stdout))
    log.Println("Hostname :", hostname)
    log.Println("Name     :", name)
    log.Println("IP       :", bindAddress)
    log.Println("Address  :", addr)

    var kv mykv.KVNode
    rpc.Register(&kv)
    rpc.Register(&t)
    rpc.HandleHTTP()
    l, _ := net.Listen("tcp", *listenAddress)
    http.Serve(l, nil)
}

func main() {
    log.SetFlags(0)
    flag.Parse()

    if *interactive {
        runInteractive(nil)
        return
    }

    if *command != "" {
        runCommand(*command, nil)
        return
    }

    runServer()
}