package mykv

import (
    "log"
    "sync"
)

type KVNode struct {
    KeyValues map[Key]KeyValue

    readIndex int
    writeIndex int
    recentReads [10]KeyValue
    recentWrites [10]KeyValue
    maxHashedKey HashedKey
    nodeMutex sync.Mutex
}

func NewNode(hash HashedKey) *KVNode {
    return &KVNode{
        KeyValues: make(map[Key]KeyValue),
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
    for i := 0; i < 10; i++ {
        entryIndex := (kv.readIndex + i) % 10
        log.Printf("[%v]    %v", i, kv.recentReads[entryIndex])
    }

    log.Println("Writes:")
    for i := 0; i < 10; i++ {
        entryIndex := (kv.writeIndex + i) % 10
        log.Printf("[%v]    %v", i, kv.recentWrites[entryIndex])
    }
}

func (kv *KVNode) Insert(args *KeyValue, reply *bool) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    keyValue, ok := kv.KeyValues[args.Key]
    if !ok {
        kv.KeyValues[args.Key] = *args
    } else if args.Time > keyValue.Time {
        kv.KeyValues[args.Key] = *args
    }
    kv.writeIndex -= 1
    if kv.writeIndex < 0 {
        kv.writeIndex = 9
    }
    kv.recentWrites[kv.writeIndex] = *args
    *reply = true
    return nil
}

func (kv *KVNode) Update(args *KeyValue, reply *bool) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    keyValue, ok := kv.KeyValues[args.Key]
    if !ok {
        *reply = false
        return ErrNoKey
    }
    if args.Time > keyValue.Time {
        kv.KeyValues[args.Key] = *args
    }
    kv.writeIndex -= 1
    if kv.writeIndex < 0 {
        kv.writeIndex = 9
    }
    kv.recentWrites[kv.writeIndex] = *args
    *reply = true
    return nil
}

func (kv *KVNode) Lookup(args *Key, reply *KeyValue) error {
    kv.nodeMutex.Lock()
    defer kv.nodeMutex.Unlock()
    defer kv.Debug()
    v, ok := kv.KeyValues[*args]
    if !ok {
        *reply = KeyValue{ Key:Key(0), Time: Timestamp(0), Value: "does not exist"}
        return ErrNoKey
    }
    kv.readIndex -= 1
    if kv.readIndex < 0 {
        kv.readIndex = 9
    }
    kv.recentReads[kv.readIndex] = v
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
    for k, keyValue := range kv.KeyValues {
        hashedKey := k.Hashed()
        if !hashInRange(prevHash, kv.maxHashedKey, hashedKey) {
            staleKeys = append(staleKeys, keyValue)
        }
    }
    return staleKeys
}
