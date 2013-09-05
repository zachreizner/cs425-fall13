package main

import "fmt"

func main() {
    log, err := ReadLog("1234567890123456789|hello there")
    fmt.Println(log)
    fmt.Println(err)
}