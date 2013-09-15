package main

import (
    "io"
    "strings"
    "testing"
)

func TestReadLog(t *testing.T) {
    logFile := strings.NewReader("123:hello\n456:world\n")

    logReader := NewLogReader(logFile)

    log, err := logReader.ReadLog()

    if err != nil {
        t.Error("Errored while reading first log\n", err)
    }

    if log.Key != "123" {
        t.Error("Wrong time in first log")
    }

    if log.Message != "hello" {
        t.Error("Wrong message. Should be \"hello\", got: ", log.Message)
    }

    log, err = logReader.ReadLog()

    if err != nil {
        t.Error("Errored while reading second log\n", err)
    }

    if log.Key != "456" {
        t.Error("Wrong time in first log")
    }

    if log.Message != "world" {
        t.Error("Wrong message. Should be \"world\", got: ", log.Message)
    }

    _, err = logReader.ReadLog()

    if err != io.EOF {
        t.Error("Did not recieve EOF at last log.\n Error:", err)
    }
}
