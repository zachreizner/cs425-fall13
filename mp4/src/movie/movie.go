package movie

import (
    "bufio"
    "fmt"
    "hash/fnv"
    "encoding/gob"
    "encoding/binary"
    "io"
    "log"
    "os"
    "strings"
    "time"
    "mykv"
)

type Entry struct {
    Title string
    Year string
}

func HashTitle(title string) mykv.Key {
    hasher := fnv.New32a()
    io.WriteString(hasher, title)
    return mykv.Key(hasher.Sum32())
}

func RunInteractive(g *mykv.KVGraph, seedAddress string) {
    promptReader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("> ")
        line, err := promptReader.ReadString('\n')
        if err != nil {
            log.Println("error reading prompt: ", err)
            break
        }

        cmd := strings.TrimSuffix(line, "\n")
        if len(cmd) == 0 {
            continue
        }

        queryStartTime := time.Now()

        g.Seed(seedAddress)

        keyvalue, err := g.Lookup(HashTitle(cmd), mykv.All)
        if err != nil {
            fmt.Println(err)
        }
        v := keyvalue.Value

        index, ok := v.([]Entry)
        if !ok {
            fmt.Println("invalid value", v)
        }

        for i, entry := range index {
            fmt.Printf("%v (%v)\n", entry.Title, entry.Year)
            if i > 50 {
                fmt.Printf("...and %v more", len(index) - i - 1)
                break
            }
        }

        fmt.Println("cmd finished; took", time.Since(queryStartTime))
    }
}

func LoadEntries(g *mykv.KVGraph, r io.Reader) error {
    log.Println("starting  movieload")
    var keywordCount uint32
    binary.Read(r, binary.BigEndian, &keywordCount)
    for keywordIndex := uint32(0); keywordIndex < keywordCount; keywordIndex++ {
        var k mykv.Key
        var entryCount uint32
        binary.Read(r, binary.BigEndian, &k)
        binary.Read(r, binary.BigEndian, &entryCount)
        entries := make([]Entry, 0, entryCount)
        for entryIndex := uint32(0); entryIndex < entryCount; entryIndex++ {
            var entryTitleLen uint32
            binary.Read(r, binary.BigEndian, &entryTitleLen)
            entryTitle := make([]byte, entryTitleLen)
            r.Read(entryTitle)

            var entryYearLen uint32
            binary.Read(r, binary.BigEndian, &entryYearLen)
            entryYear := make([]byte, entryYearLen)
            r.Read(entryYear)

            entries = append(entries, Entry{string(entryTitle), string(entryYear)})
        }
        g.InsertLocal(mykv.KeyValue{k, mykv.StampNow(), entries})
        if keywordIndex % 10000 == 0 {
            log.Printf("%v%%", float64(keywordIndex) / float64(keywordCount) * 100.0)
        }
    }
    log.Println("finished movie load")
    return nil
}

func init() {
    gob.Register([]Entry{})
}
