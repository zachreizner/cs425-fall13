package mykv

import (
    "log"
    "sync"
)

type KVNode struct {
    KeyValues map[Key]interface{}

    recentReads [10]KeyValue
    recentWrites [10]KeyValue
    maxHashedKey HashedKey
    nodeMutex sync.Mutex
}

func NewNode(hash HashedKey) *KVNode {
    return &KVNode{
        KeyValues: make(map[Key]interface{}),
        maxHashedKey: hash,
    }
}

func (kv *KVNode) Debug() {
    if len(kv.KeyValues) % 10000 == 0 {
        log.Println(len(kv.KeyValues))
    }
}

func (kv *KVNode) Show() {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()

    log.Println("Reads:")
    for i, entry := range kv.recentReads {
        log.Printf("[%v]    %v", i, entry)
    }

    log.Println("Writes:")
    for i, entry := range kv.recentWrites {
        log.Printf("[%v]    %v", i, entry)
    }
}

func (kv *KVNode) Insert(args *KeyValue, reply *bool) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KVNode) Update(args *KeyValue, reply *bool) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    _, ok := kv.KeyValues[args.Key]
    if !ok {
        *reply = false
        return ErrNoKey
    }
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KVNode) Lookup(args *Key, reply *interface{}) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    v, ok := kv.KeyValues[*args]
    if !ok {
        *reply = nil
        return ErrNoKey
    }
    *reply = v
    return nil
}

func (kv *KVNode) Delete(args *Key, reply *bool) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    _, ok := kv.KeyValues[*args]
    if !ok {
        *reply = false
        return ErrNoKey
    }
    delete(kv.KeyValues, *args)
    *reply = true
    return nil
}

func (kv *KVNode) StaleKeys(prevHash HashedKey) []KeyValue {
    staleKeys := make([]KeyValue, 0, 16)
    for k, v := range kv.KeyValues {
        hashedKey := k.Hashed()
        if !hashInRange(prevHash, kv.maxHashedKey, hashedKey) {
            staleKeys = append(staleKeys, KeyValue{k,v})
        }
    }
    return staleKeys
}
