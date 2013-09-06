package main

import "errors"
import "fmt"
import "strconv"
import "strings"
import "time"

type Log struct {
    TimeStamp time.Time
    Message string
}

// Reads and returns just one log in the reader.
func ReadLog(s string) (*Log, error) {
    logParts := strings.SplitN(s, "|", 2)
    if len(logParts) != 2 {
        return nil, errors.New("invalid log: no key")
    }
    timeStampNano, err := strconv.ParseInt(logParts[0], 10, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid key time stamp: %v", err)
    }
    return &Log{time.Unix(0, timeStampNano), logParts[1]}, nil
}
