package main

import (
    "hash/fnv"
    "encoding/json"
    "fmt"
    "encoding/binary"
    "io"
    "os"
)

func HashTitle(title string) uint32 {
    hasher := fnv.New32a()
    io.WriteString(hasher, title)
    return hasher.Sum32()
}

func main() {
    fmt.Println("Herro")
    r, _ := os.Open("entries.json")
    defer r.Close()
    w, _ := os.Create("entries.bin")
    defer w.Close()
    entryDecoder := json.NewDecoder(r)
    entries := make(map[string][][]string)
    fmt.Println("Decoding Entries")
    entryDecoder.Decode(&entries)
    fmt.Println("Writing Entries ", len(entries))
    binary.Write(w, binary.BigEndian, uint32(len(entries)))
    for k, v := range entries {
        binary.Write(w, binary.BigEndian, HashTitle(k))
        binary.Write(w, binary.BigEndian, uint32(len(v)))
        for _, entry := range v {
            binary.Write(w, binary.BigEndian, uint32(len(entry[0])))
            binary.Write(w, binary.BigEndian, []byte(entry[0]))
            binary.Write(w, binary.BigEndian, uint32(len(entry[1])))
            binary.Write(w, binary.BigEndian, []byte(entry[1]))
        }
    }
}