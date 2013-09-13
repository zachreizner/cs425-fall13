package main

import (
    "fmt"
    "io"
    "bytes"
    "encoding/binary"
    "testing"
)

func TestNextLog(t *testing.T) {
    query := "*."
    var buf bytes.Buffer
    request, err := NewRequest(&buf, query)
    if err != nil {
        t.Error("failed to make new request", err)
    }

    var reqSize uint32
    binary.Read(&buf, binary.BigEndian, &reqSize)
    if reqSize != uint32(len(query)){
        t.Error("Req set wrong lenght of query")
    }
    reqStringBuf := make([]byte, reqSize)
    buf.Read(reqStringBuf)
    if string(reqStringBuf) != query {
        t.Error("Recieved wrong string")
    }

    // clear the buffer, request thinks it can write to it
    buf.Reset()

    // log 0
    binary.Write(&buf, binary.BigEndian, int8(1))
    binary.Write(&buf, binary.BigEndian, int64(42))
    binary.Write(&buf, binary.BigEndian, uint32(5))
    fmt.Fprintf(&buf, "hello")

    // log 1
    binary.Write(&buf, binary.BigEndian, int8(1))
    binary.Write(&buf, binary.BigEndian, int64(45))
    binary.Write(&buf, binary.BigEndian, uint32(14))
    fmt.Fprintf(&buf, "whales are fun")

    binary.Write(&buf, binary.BigEndian, int8(0))


    log, err := request.NextLog()
    if err != nil {
        t.Error("failed to get log", err)
    }

    if log.TimeStamp.UnixNano() != 42 {
        t.Error("Wrong time on log")
    }

    if log.Message != "hello" {
        t.Error("Wrong message for returned log")
    }

    log, err = request.NextLog()
    if err != nil {
        t.Error("failed to get log", err)
    }

    if log.TimeStamp.UnixNano() != 45 {
        t.Error("Wrong time on log")
    }

    if log.Message != "whales are fun" {
        t.Error("Wrong message for returned log")
    }

    _, err = request.NextLog()
    if err != io.EOF {
        t.Error("Did not sent EOF at end of results")
    }
}
