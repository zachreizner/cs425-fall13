package main

import "errors"
import "fmt"
import "strconv"
import "strings"
import "time"
import "io"
import "bufio"

type Log struct {
    TimeStamp time.Time
    Message string
}

// Reads and returns just one log in the reader.
func ReadLog(r io.Reader) (*Log, error) {
    logReader := bufio.NewReader(r)
    logLine, err := logReader.ReadString('\n')
    // If we did not reach the end of the fle
    if err != io.EOF {
        logLine = logLine[:len(logLine) - 1] // Strip the newline
    }
    logParts := strings.SplitN(logLine, "|", 2)
    if len(logParts) != 2 {
        return nil, errors.New("invalid log: no key")
    }
    timeStampNano, err := strconv.ParseInt(logParts[0], 10, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid key time stamp: %v", err)
    }
    return &Log{time.Unix(0, timeStampNano), logParts[1]}, nil
}
