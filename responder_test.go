package main

import (
    "testing"
    "encoding/binary"
    "bytes"
    "strings"
)

type BufferCloser struct {
    buf *bytes.Buffer
    closed bool
}


func (b BufferCloser) Read(p []byte) (n int, err error) {
    return b.buf.Read(p)
}

func (b BufferCloser) Write(p []byte) (n int, err error) {
    return b.buf.Write(p)
}

func (b BufferCloser) Close() error {
    b.closed = true
    return nil
}

func TestHandleQuery(t *testing.T) {
    var buf bytes.Buffer
    binary.Write(&buf, binary.BigEndian, uint32(5))
    buf.WriteString("hello")
    logFile := strings.NewReader("123|hello")
    HandleQuery(BufferCloser{buf: &buf}, logFile)

    var status uint8
    binary.Read(&buf, binary.BigEndian, &status)
    if status != 1 {
        t.Errorf("query returned no results")
        return
    }

    var timeStamp int32
    var logSize uint32
    binary.Read(&buf, binary.BigEndian, &timeStamp)
    binary.Read(&buf, binary.BigEndian, &logSize)
    logMessage := make([]byte, logSize)
    buf.Read(logMessage)

    if timeStamp != 123 {
        t.Errorf("query returned incorrect time stamp")
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
