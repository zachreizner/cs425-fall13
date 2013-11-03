package mykv

import (
    "hash/fnv"
    "encoding/binary"
    "errors"
)

var (
    ErrNoKey = errors.New("no such key")
)

type KeyValueStore struct {
    KeyValues map[int]interface{}
    MyHash uint32
    NextHash uint32

}

func HashKey(key uint32) uint32 {
    hasher := fnv.New32a()
    binary.Write(hasher, binary.BigEndian, key)
    return hasher.Sum32()
}

type KeyValue struct {
    Key int
    Value interface{}
}

func (kv *KeyValueStore) Insert(args *KeyValue, reply *bool) error {
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KeyValueStore) Update(args *KeyValue, reply *bool) error {
    _, ok := kv.KeyValues[args.Key]
    if !ok {
        *reply = false
        return ErrNoKey
    }
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KeyValueStore) Lookup(args *int, reply *interface{}) error {
    v, ok := kv.KeyValues[*args]
    if !ok {
        *reply = nil
        return ErrNoKey
    }
    *reply = v
    return nil
}

func (kv *KeyValueStore) Delete(args *int, reply *bool) error {
    _, ok := kv.KeyValues[*args]
    if !ok {
        *reply = false
        return ErrNoKey
    }
    delete(kv.KeyValues, *args)
    *reply = true
    return nil
}

