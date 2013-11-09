package main

import (
    "flag"
    "io"
    "log"
    "net"
    "net/http"
    "net/rpc"
    "os"
    "time"

    "mykv"
)

// Replace this simple implentation with whatever you have
// Note that we may get rid of the concept of a leader all together
type Leader struct {
    lastID int32
}

func (l *Leader) GetID(args *int32, id *int32) error {
    l.lastID += 1
    *id = l.lastID
    return nil
}

var listenAddress = flag.String("bind", ":7777", "the address for listening to services")
var seedAddress = flag.String("seed", "", "the address of some machine to grab the inital membertable from")
var machineName = flag.String("name", "", "the name of this machine")
var logFile = flag.String("logs", "machine.log", "the file name to store the log in")


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
func main() {
    flag.Parse()

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

    addr := bindAddress + ":" + bindPort

    // Configure the log file to be something nice
    log.SetPrefix("[\x1B[" + getColor(3) + "m" + name + "\x1B[0m]:")
    log.SetFlags(0)

    logfd, err := os.Create(*logFile)

    if err != nil {
        log.Println(err)
    }

    log.SetOutput(io.MultiWriter(logfd, os.Stdout))
    log.Println("Hostname :", hostname)
    log.Println("Name     :", name)
    log.Println("IP       :", bindAddress)
    log.Println("Address  :", addr)

    kv := new(mykv.KVNode)
    leader := new(Leader)
    rpc.Register(kv)
    rpc.Register(leader)
    rpc.HandleHTTP()
    l, _ := net.Listen("tcp", ":7777")
    go http.Serve(l, nil)

    client, err := rpc.DialHTTP("tcp", addr)
    if err != nil {
        log.Fatal(err)
    }
    var gotId int32
    client.Call("Leader.GetID", &gotId, &gotId)
    log.Println(gotId)

    for {
        time.Sleep(1)
    }
}