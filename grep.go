package main

import (
    "fmt"
    "regexp"
)

func QueryLog(log Log, query string) bool {
    var regexQuery, err = regexp.Compile(query)
    if err != nil {
        fmt.Errorf("invalid regular expression. %v", err)
        return false
    }

    return regexQuery.MatchString(log.Message)
}
