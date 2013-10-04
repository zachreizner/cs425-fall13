package main

import (
    "bytes"
    "encoding/gob"
    "flag"
    "hash/crc32"
    "io"
    "leader"
    "log"
    "math/rand"
    "membertable"
    "net"
    "time"
)

var listenAddress = flag.String("bind", ":7777", "the address for listening")
var leaderAddress = flag.String("leader", "", "the address of the leader machine; leave unset to make this process leader")
var seedAddress = flag.String("seed", "", "the address of some machine to grab the inital membertable from")
var logFile = flag.String("logs", "machine.log", "the file name to store the log in")

type PacketHeader struct {
    Length int32
    Checksum uint32
}

type PacketWriter struct {
    b bytes.Buffer
}


func (p *PacketWriter) Write(b []byte) (int, error) {
    return p.b.Write(b)
}

func (p *PacketWriter) Send(w io.Writer) error {
    // Create the header for the packet, which involves hashing the buffer
    hasher := crc32.NewIEEE()
    hasher.Write(p.b.Bytes()) // Hash interface never has errors
    packetHeader := PacketHeader{
        Length: int32(p.b.Len()),
        Checksum: hasher.Sum32(),
    }

    // Send the header
    headerEncoder := gob.NewEncoder(w)
    err := headerEncoder.Encode(packetHeader)
    if err != nil {
        return err
    }

    // Send the payload
    _, err = p.b.WriteTo(w)
    return err
}

func sendHeartbeatToMember(m * membertable.Member, t * membertable.Table) error {
    // Connect to the given member
    conn, err := net.Dial("udp", m.Address)
    if err != nil {
        return err
    }

    defer conn.Close()

    // Write out the hearbeat to a packet
    var heartbeatPacket PacketWriter
    if err := t.WriteTo(&heartbeatPacket); err != nil {
        return err
    }

    // Send the packet down the wire
    return heartbeatPacket.Send(conn)
}

func sendHeartbeat(t * membertable.Table) error {
    // We are alone on this earth :(
    if len(t.Members) == 0 {
        return nil
    }

    // Form a list of each member from the ID amp
    memberList := make([]*membertable.Member, len(t.Members))
    memberIndex := 0
    for _, v := range t.Members {
        memberList[memberIndex] = &v
        memberIndex++
    }

    // Choose a member at random and send their heartbeat
    sendToMember := memberList[rand.Int() % memberIndex]
    return sendHeartbeatToMember(sendToMember, t)
}

func heartbeatProcess(t * membertable.Table, fatalChan chan bool) {
    for {
        log.Printf("Sending Hearbeat...\n")
        err := sendHeartbeat(t)
        if err != nil {
            log.Printf("    %v\n", err)
        }
        time.Sleep(50 * time.Millisecond)
    }
    fatalChan <- true
}

func leaderProcess(fatalChan chan bool) {
    if err := leader.Run(); err != nil {
        log.Fatal(err)
    }
    fatalChan <- true
}

func main() {
    flag.Parse()

    fatalChan := make(chan bool)

    var id membertable.ID
    var err error

    if *leaderAddress == "" {
        // We are the LEADER! Take an ID and take our role as Master of IDs.
        if id, err = leader.IncrementIDFile(); err != nil {
            log.Fatal(err)
            return
        }
        go leaderProcess(fatalChan)
    } else {
        if id, err = leader.RequestID(*leaderAddress); err != nil {
            log.Fatal(err)
            return
        }
    }

    log.Printf("My ID: %v", id)

    var t membertable.Table
    go heartbeatProcess(&t, fatalChan)

    <-fatalChan
}