package mykv

import (
    "fmt"
    "sort"

    "membertable"
)

type Vertex struct {
    Addr string
    Hash HashedKey
    LocalNode *KVNode
}

type KVGraph struct {
    NodeIndex []*Vertex
    Connector RPCConnector
}

func (g *KVGraph) Len() int {
    return len(g.NodeIndex)
}

func (g *KVGraph) Less(i, j int) bool {
    return g.NodeIndex[i].Hash < g.NodeIndex[j].Hash
}

func (g *KVGraph) Swap(i, j int) {
    tmp := g.NodeIndex[i]
    g.NodeIndex[i] = g.NodeIndex[j]
    g.NodeIndex[j] = tmp
}


func (g *KVGraph) Seed(seedAddr string) error {
    client, err := g.Connector.Connect(seedAddr)
    if err != nil {
        return err
    }
    defer client.Close()
    var dummy int
    var members []membertable.Member
    client.Call("Table.RPCGetActiveMembers", dummy, &members)
    fmt.Println(members)
    g.NodeIndex = make([]*Vertex, 0, 16)
    for _, member := range members {

        v := &Vertex{
            Addr: member.ID.Address,
            Hash: HashedKey(member.ID.Hashed()),
            LocalNode: nil,
        }
        g.NodeIndex = append(g.NodeIndex, v)
    }
    sort.Sort(g)
    for _, v := range g.NodeIndex {
        fmt.Println(v.Addr, v.Hash)
    }
    return nil
}


func (g *KVGraph) Insert(kv KeyValue) error {
    hashedKey := kv.HashedKey()
    for i, v := range g.NodeIndex {
        prevHash := g.circularIndex(i-1).Hash
        if hashInRange(prevHash, v.Hash, hashedKey) {
            remoteNode, err := g.Connector.Connect(v.Addr)
            if err != nil {
                return err
            }
            var reply bool
            err = remoteNode.Call("KVNode.Insert", &kv, &reply)
            if err != nil {
                return err
            }
            return nil
        }
    }
    return nil
}

func (g *KVGraph) Update(kv KeyValue) error {
    return nil
}

func (g *KVGraph) Lookup(k Key) (interface{}, error) {
    return 0, nil
}

func (g *KVGraph) Delete(k Key) error {
    return nil
}

func (g *KVGraph) circularIndex(idx int) *Vertex {
    if idx < 0 {
        return g.NodeIndex[len(g.NodeIndex) + idx]
    }
    if idx > len(g.NodeIndex) {
        return g.NodeIndex[idx - len(g.NodeIndex)]
    }
    return g.NodeIndex[idx]
}

func (g *KVGraph) HandleStaleKeys() {
    for i, v := range g.NodeIndex {
        if v.LocalNode != nil {
            prevHash := g.circularIndex(i-1).Hash

            var staleKeys []KeyValue
            v.LocalNode.StaleKeys(&prevHash, &staleKeys)

            for _, kv := range staleKeys {
                g.Insert(kv)
                var success bool
                v.LocalNode.Delete(&kv.Key, &success)
            }
        }
    }
}