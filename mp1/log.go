package main

import  (
    "bufio"
    "errors"
    "io"
    "strings"
)

type Log struct {
    Key string
    Message string
}

type LogReader struct {
    reader *bufio.Reader
}

// Creates a new structure for the buffered reading of logs.
func NewLogReader(r io.Reader) *LogReader {
    return &LogReader{bufio.NewReader(r)}
}

// Reads and returns just one log in the reader.
func (r *LogReader) ReadLog() (*Log, error) {
    // We reuse bufio.Reader between calls because bufio often reads more than it returns.
    logLine, err := r.reader.ReadString('\n')
    logLine = strings.TrimSuffix(logLine, "\n")

    // If we reached an empty line, it is either a blank line, or the file has run out.
    if len(logLine) == 0 {
        if err == io.EOF {
            return nil, io.EOF
        }
        return r.ReadLog()
    }

    // Split the log into its parameters
    logParts := strings.SplitN(logLine, ":", 2)
    if len(logParts) != 2 {
        return nil, errors.New("invalid log: no key")
    }

    return &Log{logParts[0], logParts[1]}, nil
}
