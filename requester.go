package main

import (
    "encoding/binary"
    "io"
    "time"
)

type Request struct {
    c io.ReadWriter
    status uint8
}

func NewRequest(req io.ReadWriter, query string) (*Request, error) {
    if err := binary.Write(req, binary.BigEndian, uint32(len(query))); err != nil {
        return nil, err
    }

    if _, err := req.Write([]byte(query)); err != nil {
        return nil, err
    }

    r := &Request{req, 1}
    return r, nil
}

// Pull the next log from the request
func (r *Request) NextLog() (*Log, error) {
    req := r.c
    if r.status == 1 {
        if err := binary.Read(req, binary.BigEndian, &r.status); err != nil {
           return nil, err
        }

        if r.status == 0 {
            return nil, io.EOF
        }

        var timeStamp int64
        var logSize uint32

        if err := binary.Read(req, binary.BigEndian, &timeStamp); err != nil {
            return nil, err
        }

        if err := binary.Read(req, binary.BigEndian, &logSize); err != nil {
            return nil, err
        }

        logMessage := make([]byte, logSize)

        if _, err := req.Read(logMessage); err != nil {
            return nil, err
        }

        return &Log{time.Unix(0, timeStamp), string(logMessage)}, nil
    }

    return nil, io.EOF
}
