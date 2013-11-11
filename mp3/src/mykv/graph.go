package mykv

import (
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
    return nil
}

func (g *KVGraph) FindVertex(k Key) *Vertex {
    hashedKey := k.Hashed()
    for i, v := range g.NodeIndex {
        prevHash := g.circularIndex(i-1).Hash
        if hashInRange(prevHash, v.Hash, hashedKey) {
            return v
        }
    }
    return nil
}


func (g *KVGraph) Insert(kv KeyValue) error {
    v := g.FindVertex(kv.Key)
    remoteNode, err := g.Connector.Connect(v.Addr)
    if err != nil {
        return err
    }
    defer remoteNode.Close()

    var reply bool
    err = remoteNode.Call("KVNode.Insert", &kv, &reply)
    if err != nil {
        return err
    }
    return nil
}

func (g *KVGraph) Update(kv KeyValue) error {
    v := g.FindVertex(kv.Key)
    remoteNode, err := g.Connector.Connect(v.Addr)
    if err != nil {
        return err
    }
    defer remoteNode.Close()

    var reply bool
    err = remoteNode.Call("KVNode.Update", &kv, &reply)
    if err != nil {
        return err
    }
    return nil
}

func (g *KVGraph) Lookup(k Key) (interface{}, error) {
    v := g.FindVertex(k)
    remoteNode, err := g.Connector.Connect(v.Addr)
    if err != nil {
        return nil, err
    }
    defer remoteNode.Close()

    var value interface{}
    err = remoteNode.Call("KVNode.Lookup", &k, &value)
    if err != nil {
        return nil, err
    }
    return value, nil
}

func (g *KVGraph) Delete(k Key) error {
    v := g.FindVertex(k)
    remoteNode, err := g.Connector.Connect(v.Addr)
    if err != nil {
        return err
    }
    defer remoteNode.Close()

    var reply bool
    err = remoteNode.Call("KVNode.Delete", &k, &reply)
    if err != nil {
        return err
    }
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