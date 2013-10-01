package main

import "testing"

func testQueryMatch(t *testing.T, message string, query string) {
    l := Log{
        Message: message,
    }
    if !QueryLog(l, query) {
        t.Errorf("query for \"%s\" on message \"%s\" was expected to match", query, message)
    }
}

func testQueryNotMatch(t *testing.T, message string, query string) {
    l := Log{
        Message: message,
    }
    if QueryLog(l, query) {
        t.Errorf("query for \"%s\" on message \"%s\" was expected to not match", query, message)
    }
}

func TestQueryHello(t *testing.T) {
    testQueryMatch(t, "hello", "hello")
}

func TestQueryHello2(t *testing.T) {
    testQueryNotMatch(t, "hello", "hello2")
}


func TestQueryHello3(t *testing.T) {
    testQueryMatch(t, "hello3", "hello")
}

func TestQueryHello4(t *testing.T) {
    testQueryMatch(t, "4hello", "hello")
}

func TestQueryWild(t *testing.T) {
    testQueryMatch(t, "hello jlo", "hel*lo")
}

func TestQuerySuffix(t *testing.T) {
    testQueryMatch(t, "moneypenny", "penny$")
}

func TestQueryNotSuffix(t *testing.T) {
    testQueryNotMatch(t, "pennypincher", "penny$")
}

func TestQueryPrefix(t *testing.T) {
    testQueryMatch(t, "nickel and dime", "^nickel")
}

func TestQueryNotPrefix(t *testing.T) {
    testQueryNotMatch(t, "one hundred nickels", "^nickel")
}