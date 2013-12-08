package mykv

import (
    "hash/fnv"
    "encoding/binary"
    "errors"
    "time"
    "net/rpc"
)

var (
    ErrNoKey = errors.New("no such key")
)

type RPCConnector interface {
    Connect(addr string) (*rpc.Client, error)
}

type Timestamp int64

func StampNow() Timestamp{
    return Timestamp(time.Now().UnixNano())
}

type Key uint32
type HashedKey uint32

func (k Key) Hashed() HashedKey {
    hasher := fnv.New32a()
    binary.Write(hasher, binary.BigEndian, k)
    return HashedKey(hasher.Sum32())
}

type KeyValue struct {
    Key Key
    Time Timestamp
    Value interface{}
}

func (kv KeyValue) HashedKey() HashedKey {
    return kv.Key.Hashed()
}

func hashInRange(prev HashedKey, current HashedKey, hash HashedKey) bool {
    return (hash > prev && hash <= current) ||
           ((prev >= current) && (hash > prev || hash <= current))
}
