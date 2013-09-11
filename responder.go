package main

import (
    "fmt"
    "os"
    "io"
    "net"
    "encoding/binary"
)

func HandleQuery(connection io.ReadWriteCloser, logfile io.Reader) {

    var size uint32
    binary.Read(connection, binary.BigEndian, &size)
    str_buf := make([]byte, size)
    read_bytes, err := connection.Read(str_buf)
    if uint32(read_bytes) != size {
        fmt.Println("HandleQuery: improper size read")
        os.Exit(1)
    }

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    query := string(str_buf)

    for {
        log, err := ReadLog(logfile)
        if err != nil {
            terminateSignal := uint8(0)
            writeErr := binary.Write(connection, binary.BigEndian, terminateSignal)
            if writeErr != nil {
                fmt.Println(writeErr)
                os.Exit(1)
            }
            return
        }


        if QueryLog(*log, query) {
            contSignal := uint8(1)
            binary.Write(connection, binary.BigEndian, contSignal)
            binary.Write(connection, binary.BigEndian, log.TimeStamp.UnixNano())
            binary.Write(connection, binary.BigEndian, uint32(len(log.Message)))
            fmt.Fprintf(connection, log.Message)
        }
    }
}

func ListenForQueries(listener net.Listener, logfile string) {
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        file, file_err := os.Open(logfile)
        if file_err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        defer file.Close()

        go HandleQuery(conn, file)
    }
}
