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
    "os/signal"
    "regexp"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "syscall"
    "time"

    "mykv"
    "membertable"
    "movie"
)

var listenAddress = flag.String("bind", ":7777", "the address for listening to services")
var seedAddress = flag.String("seed", "", "the address of some machine to grab the inital membertable from")
var machineName = flag.String("name", "", "the name of this machine")
var logFile = flag.String("logs", "machine.log", "the file name to store the log in")
var command = flag.String("run", "", "command to run")
var interactive = flag.Bool("interactive", false, "set to true to run interactively; cancels running a node and run command")
var movieInteractive = flag.Bool("movie", false, "set to true to run interactively in movie search mode")
var loadMovie = flag.Bool("load", false, "set to true to cause the machines to load the movie index")

type HTTPRPCConnector struct {

}

func (c HTTPRPCConnector) Connect(addr string) (*rpc.Client, error) {
    return rpc.DialHTTP("tcp", addr)
}

type Controller struct {
    g *mykv.KVGraph
}

func (c *Controller) LoadMovies(args *int, reply *bool) error {
    go func() {
        entriesFile, err := os.Open("entries.bin")
        if err != nil {
            log.Printf("error opening entries file: %v", err)
            return
        }
        defer entriesFile.Close()
        if err := movie.LoadEntries(c.g, entriesFile); err != nil {
            log.Printf("error loading entries: %v", err)
            return
        }
    }()
    return nil
}

func loadMovies() {
    client, err := rpc.DialHTTP("tcp", *seedAddress)
    if err != nil {
        log.Printf("error connecting to seed: %v", err)
        return
    }

    var dummy int
    var members []membertable.Member
    client.Call("Table.RPCGetActiveMembers", dummy, &members)
    client.Close()


    memberDoneChan := make(chan int, 10)
    memberDone := func() {
        memberDoneChan <- 1
    }
    membersRunning := 0
    for _, member := range members {
        membersRunning += 1
        go func(addr string) {
            defer memberDone()
            remoteNode, err := rpc.DialHTTP("tcp", addr)
            if err != nil {
                log.Printf("error connecting to node: %v", err)
                return
            }
            var reply bool
            err = remoteNode.Call("Controller.LoadMovies", &dummy, &reply)
            remoteNode.Close()
            if err != nil {
                log.Printf("error calling node: %v", err)
                return
            }
        }(member.ID.Address)
    }
    for membersRunning > 0 {
        <- memberDoneChan
        membersRunning -= 1
    }
}

type commandDispatch struct {
    RE *regexp.Regexp
    Handler func([]string, *mykv.KVGraph) bool
}

func parseConstLevel(s string) mykv.ConstLvl {
    if strings.EqualFold(s, "one") {
        return mykv.One
    }
    if strings.EqualFold(s, "quorum") {
        return mykv.Quorum
    }
    if strings.EqualFold(s, "All") {
        return mykv.All
    }
    return mykv.Invalid

}
func handleInsert(params []string, g *mykv.KVGraph) bool {
    if len(params) != 4 {
        log.Println("not enough params")
        return true
    }
    fmt.Println("const level", params[2])
    keyUint, err := strconv.ParseUint(params[1], 10, 32)
    if err != nil {
        log.Println("invalid integer key")
        return true
    }
    c := parseConstLevel(params[2])
    if c == mykv.Invalid {
        log.Println("invalid consistancy level")
        return true
    }
    kv := mykv.KeyValue{ mykv.Key(keyUint), mykv.StampNow(), params[3] }
    if err := g.Insert(kv, c); err != nil {
        log.Println("insert error: ", err)
        return true
    }
    return true
}

func handleUpdate(params []string, g *mykv.KVGraph) bool {
    if len(params) != 4 {
        log.Println("not enough params")
        return true
    }
    fmt.Println("const level", params[2])
    keyUint, err := strconv.ParseUint(params[1], 10, 32)
    if err != nil {
        log.Println("invalid integer key")
        return true
    }
    c := parseConstLevel(params[2])
    if c == mykv.Invalid {
        log.Println("invalid consistancy level")
        return true
    }
    kv := mykv.KeyValue{ mykv.Key(keyUint), mykv.StampNow(), params[3] }
    if err := g.Update(kv, c); err != nil {
        log.Println("update error: ", err)
        return true
    }
    return true
}

func handleLookup(params []string, g *mykv.KVGraph) bool {
    if len(params) != 3 {
        log.Println("not enough params")
        return true
    }
    keyUint, err := strconv.ParseUint(params[1], 10, 32)
    if err != nil {
        log.Println("invalid integer key")
        return true
    }
    c := parseConstLevel(params[2])
    if c == mykv.Invalid {
        log.Println("invalid consistancy level")
        return true
    }
    kv, err := g.Lookup(mykv.Key(keyUint), c)
    if err != nil {
        log.Println("lookup error: ", err)
        return true
    }
    if kv != nil {
        fmt.Println(kv.Value)
    } else {
        fmt.Println("key does not exist")
    }
    return true
}

func handleDelete(params []string, g *mykv.KVGraph) bool {
    if len(params) != 3 {
        log.Println("not enough params")
        return true
    }
    keyUint, err := strconv.ParseUint(params[1], 10, 32)
    if err != nil {
        log.Println("invalid integer key")
        return true
    }
    c := parseConstLevel(params[2])
    if c == mykv.Invalid {
        log.Println("invalid consistancy level")
        return true
    }
    if err = g.Delete(mykv.Key(keyUint), c); err != nil {
        log.Println("delete error: ", err)
        return true
    }
    return true
}

var commandRE = []commandDispatch{
    { regexp.MustCompile(`insert\s+(\S+)\s+(\S+)\s+(.*)`), handleInsert },
    { regexp.MustCompile(`update\s+(\S+)\s(\S+)\s+(.*)`), handleUpdate },
    { regexp.MustCompile(`lookup\s+(\S+)\s(\S+)`), handleLookup },
    { regexp.MustCompile(`delete\s+(\S+)\s(\S+)`), handleDelete },
}

func runCommand(cmd string, g *mykv.KVGraph) bool {
    for _, command := range commandRE {
        matches := command.RE.FindStringSubmatch(cmd)
        if matches != nil {
            g.Seed(*seedAddress)
            return command.Handler(matches, g)
        }
    }
    log.Println("invalid command")
    return true
}

func runInteractive(g *mykv.KVGraph) {
    if *seedAddress == "" {
        log.Println("must have machine to connect to in interactive mode")
        return
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
func getColor(id uint32) string {
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

func runServer(g *mykv.KVGraph) {
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

    localNode := mykv.NewNode(mykv.HashedKey(myID.Hashed()))

    var t membertable.Table
    t.Init(myID)
    t.Changed = func(t *membertable.Table, changedMembers []membertable.ID, dropped bool) {
        log.Println("membertable changed")
        g.SetByMembertable(t.ActiveMembers())
        myVertex := g.FindNode(mykv.HashedKey(myID.Hashed()))
        myVertex.LocalNode = localNode
        go g.HandleStaleKeys(changedMembers, dropped)
    }

    addr := bindAddress + ":" + bindPort

    // Configure the log file to be something nice
    log.SetPrefix("[\x1B[" + getColor(myID.Hashed()) + "m" + name + "\x1B[0m]:")

    logfd, err := os.Create(*logFile)

    if err != nil {
        log.Println(err)
    }

    log.SetOutput(io.MultiWriter(logfd, os.Stdout))
    log.Println("Hostname :", hostname)
    log.Println("Name     :", name)
    log.Println("IP       :", bindAddress)
    log.Println("Address  :", addr)

    go t.SendHeartbeatProcess(nil)

    s := Controller{g}

    rpc.Register(&s)
    rpc.Register(localNode)
    rpc.Register(&t)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", *listenAddress)
    if err != nil {
        log.Println(err)
        return
    }

    if *seedAddress != "" {
        log.Printf("sending heartbeat to seed member")
        if err = t.SendHeartbeatToAddress(*seedAddress); err != nil {
             log.Println(err)
            return
        }
    }

    // Setup a signal for showing the last 10 reads/writes
    usrSigChan := make(chan os.Signal, 0)
    signal.Notify(usrSigChan, syscall.SIGUSR1)
    go func() {
        for _ = range usrSigChan {
            localNode.Show()
        }
    }()

    // Setup some so we can exit clean on SIGTSTP
    var exitMutex sync.Mutex // Used to prevent exit while handling signal
    sigChan := make(chan os.Signal, 0)
    signal.Notify(sigChan, syscall.SIGTSTP)

    go func() {
        sig := <-sigChan
        exitMutex.Lock()
        log.Printf("got signal %v", sig)
        t.Changed = nil
        l.Close()
        g.RemoveLocalNodes()
        exitMutex.Unlock()
    }()

    http.Serve(l, nil)
    exitMutex.Lock()
}

func main() {
    runtime.GOMAXPROCS(2)
    log.SetFlags(0)
    flag.Parse()

    var g mykv.KVGraph
    g.Connector = HTTPRPCConnector{}

    if *loadMovie {
        if *seedAddress == "" {
            log.Println("must have machine to connect to load movies")
            return
        }
        loadMovies()
        return
    }

    if *movieInteractive {
        if *seedAddress == "" {
            log.Println("must have machine to connect to in movie search mode")
            return
        }
        movie.RunInteractive(&g, *seedAddress)
        return
    }

    if *interactive {
        runInteractive(&g)
        return
    }

    if *command != "" {
        runCommand(*command, &g)
        return
    }

    runServer(&g)
}
