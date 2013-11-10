package leader

import (
    "encoding/binary"
    "net"
    "os"
    "membertable"
)

const idFilename = "id.bin"

func handleRequest(c net.Conn, id membertable.IDNum) {
    defer c.Close()
    binary.Write(c, binary.BigEndian, id)
}

// Return the next ID (stored in the ID file) and write out the new next ID to the ID file.
func IncrementIDFile() (membertable.IDNum, error) {
    f, err := os.OpenFile(idFilename, os.O_RDWR | os.O_CREATE, os.ModePerm)
    if err != nil {
        return 0, err
    }

    defer f.Close()

    fi, err := f.Stat()
    if err != nil {
        return 0, err
    }

    // A zero file size means there was no ID file, so we should create one with the second ID
    // (which is 1) and return the starting ID (which is 1).
    if fi.Size() == 0 {
        err = binary.Write(f, binary.BigEndian, membertable.IDNum(1))
        return 0, err
    }

    // Read the id which we will return
    var id membertable.IDNum
    if err = binary.Read(f, binary.BigEndian, &id); err != nil {
        return 0, err
    }

    // Move to the beginning of the file and erase it
    if _, err = f.Seek(0, 0); err != nil {
        return id, err
    }

    if err = f.Truncate(0); err != nil {
        return id, err
    }

    // Write out what the next id should be, but do not return it
    err = binary.Write(f, binary.BigEndian, membertable.IDNum(id + 1))
    return id, err
}

// Continually run and pass out monotomically increasing IDs to users as they request then.
// Connecting to us implcitly requests a new ID, which we will never pass out again.
func Run() error {
    ln, err := net.Listen("tcp", ":38449")
    if err != nil {
        return err
    }


    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }

        // Get the next ID
        nextID, err := IncrementIDFile()
        if err != nil {
            return err
        }

        // Send it out
        go handleRequest(conn, nextID)
    }

    return nil
}

// Connect to the given leader and get a new ID for the group they lead.
func RequestID(leaderAddress string) (membertable.IDNum, error) {
    c, err := net.Dial("tcp", leaderAddress)
    if err != nil {
        return 0, err
    }

    defer c.Close()

    var id membertable.IDNum
    if err = binary.Read(c, binary.BigEndian, &id); err != nil {
        return 0, err
    }

    return id, nil
}
