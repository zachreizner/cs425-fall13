package mykv

import (
    "log"
)

type KVNode struct {
    KeyValues map[Key]interface{}
    maxHashedKey HashedKey
}

func NewNode(hash HashedKey) *KVNode {
    return &KVNode{
        KeyValues: make(map[Key]interface{}),
        maxHashedKey: hash,
    }
}

func (kv *KVNode) Debug() {
    log.Println(kv.KeyValues)
}

func (kv *KVNode) Insert(args *KeyValue, reply *bool) error {
    defer kv.Debug()
    kv.KeyValues[args.Key] = args.Value
    *reply = true
    return nil
}

func (kv *KVNode) Update(args *KeyValue, reply *bool) error {
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

func (kv *KVNode) StaleKeys(args *HashedKey, reply *[]KeyValue) error {
    staleKeys := make([]KeyValue, 0, 16)
    for k, v := range kv.KeyValues {
        hashedKey := k.Hashed()
        if !hashInRange(*args, kv.maxHashedKey, hashedKey) {
            staleKeys = append(staleKeys, KeyValue{k,v})
        }
    }
    *reply = staleKeys
    return nil
}

func (kv *KVNode) DeleteStaleKeys(args *HashedKey, reply *[]KeyValue) error {
    for k := range kv.KeyValues {
        hashedKey := k.Hashed()
        if !hashInRange(*args, kv.maxHashedKey, hashedKey) {
            delete(kv.KeyValues, k)
        }
    }
    return nil
}