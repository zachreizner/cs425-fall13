package mykv

import (
    "errors"
)

var (
    ErrNoKey = errors.New("no such key")
)


type KeyValueStore struct {
    KeyValues map[int]interface{}
}

type KeyValue struct {
    Key int
    Value interface{}
}

func (kv *KeyValueStore) Put(args *KeyValue, reply *bool) error {
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KeyValueStore) Get(args *int, reply *interface{}) error {
    v, ok := kv.KeyValues[*args]
    if ok {
        *reply = v
        return nil
    } else {
        *reply = nil
        return ErrNoKey
    }
    return nil
}
