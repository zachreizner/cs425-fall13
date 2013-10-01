package main

import (
    "bytes"
    "encoding/binary"
    "strings"
    "testing"
)

func TestHandleQuery(t *testing.T) {
    var buf bytes.Buffer
    binary.Write(&buf, binary.BigEndian, uint32(5))
    buf.WriteString("hello")
    logFile := strings.NewReader("123:helloooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo")

    HandleQuery(&buf, logFile)

    var status uint8
    binary.Read(&buf, binary.BigEndian, &status)
    if status != 1 {
        t.Errorf("query returned no results")
        return
    }

    var keySize uint32
    var logSize uint32
    binary.Read(&buf, binary.BigEndian, &keySize)
    logKey := make([]byte, keySize)
    buf.Read(logKey)
    binary.Read(&buf, binary.BigEndian, &logSize)
    logMessage := make([]byte, logSize)
    buf.Read(logMessage)

    if string(logKey) != "123" {
        t.Errorf("query returned incorrect key")
    }

    if string(logMessage) != "hello" {
        t.Errorf("query returned incorrect message")
    }

    binary.Read(&buf, binary.BigEndian, &status)
    if status != 0 {
        t.Errorf("query did not termiante results")
        return
    }
}

func TestProtocol(t *testing.T) {
    var buf bytes.Buffer
    logFile := strings.NewReader("123:hello")

    req, err := NewRequest(&buf, "hello")
    if err != nil {
        t.Errorf("requester returned error: %v", err)
        return
    }

    HandleQuery(&buf, logFile)

    log, err := req.NextLog()
    if err != nil {
        t.Errorf("requester returned error instead of log: %v", err)
        return
    }

    if log.Key != "123" {
        t.Errorf("returned log key expected %v; was actually %v", 123, log.Key)
        return
    }

    _, err = req.NextLog()
    if err == nil {
        t.Errorf("responder returned too many logs")
        return
    }
}
