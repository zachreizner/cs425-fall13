package mykv

import (
    "log"
    "sort"

    "membertable"
)

const numberOfReplicas = int(3)

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
    g.SetByMembertable(members)
    return nil
}

func (g *KVGraph) SetByMembertable(members []membertable.Member)  {
    g.NodeIndex = make([]*Vertex, 0, len(members))
    for _, member := range members {

        v := &Vertex{
            Addr: member.ID.Address,
            Hash: HashedKey(member.ID.Hashed()),
            LocalNode: nil,
        }
        g.NodeIndex = append(g.NodeIndex, v)
    }
    sort.Sort(g)
}

func (g *KVGraph) FindNode(k HashedKey) *Vertex {
    for _, v := range g.NodeIndex {
        if v.Hash == k {
            return v
        }
    }
    return nil
}

func (g *KVGraph) FindVerticies(k Key) []*Vertex {
    return verticiesHave(k, g.NodeIndex)
}


func (g *KVGraph) Insert(kv KeyValue) error {
    // TODO change for quarrum
    verts := g.FindVerticies(kv.Key)
    err := error(nil)
    for _, v := range verts {
        currentErr := g.insertToVert(kv, v)
        // TODO is this the right way to handle errors?
        if currentErr != nil {
            err = currentErr
        }
    }
    return err
}

func (g *KVGraph) insertToVert(kv KeyValue, v *Vertex) error {
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
    // TODO change for quarrum
    verts := g.FindVerticies(kv.Key)
    err := error(nil)
    for _, v := range verts {
        currentErr := g.insertToVert(kv, v)
        if currentErr != nil {
            err = currentErr
        }
    }
    return err
}

func (g *KVGraph) updateVertex(kv KeyValue, v *Vertex) error {
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
    verts := g.FindVerticies(k)
    var newestData interface{}
    var err error
    for _, v := range verts{
        data, currentErr := g.lookupVertex(k, v)
        // TODO check the date rather than just errors
        if currentErr == nil {
            newestData = data
        }
    }
    return newestData, err
}

func (g *KVGraph) lookupVertex(k Key, v *Vertex) (interface{}, error) {
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
    verts := g.FindVerticies(k)
    err := error(nil)
    for _, v := range verts {
        currentErr := g.deleteFromVertex(k, v)
        if currentErr != nil {
            err = currentErr
        }
    }
    return err
}

func (g *KVGraph) deleteFromVertex(k Key, v *Vertex) error {
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

func verticiesHave(k Key, verts []*Vertex) []*Vertex {
    hashedKey := k.Hashed()
    verticies := make([]*Vertex, 0, numberOfReplicas)
    if len(verts) <= numberOfReplicas {
        verticies = append(verticies, verts...)
        return verticies
    }
    for i, v := range verts {
        prevHash := verts[loop(i - numberOfReplicas, len(verts))].Hash
        if hashInRange(prevHash, v.Hash, hashedKey) {
            verticies = append(verticies, v)
        }
    }
    return verticies
}

func loop(index, length int) int {
    if index >= 0 {
        return index % length
    }
    return length + (index % length)
}

func shouldHave(k Key, verts []*Vertex, h HashedKey) bool {
    candidates := verticiesHave(k, verts)
    for _, vert := range candidates {
        if vert.Hash == h {
            return true
        }
    }
    return false
}

func (g *KVGraph) findLocalNode() *Vertex {
    for _, me := range g.NodeIndex {
        if me.LocalNode != nil {
            return me
        }
    }
    return nil
}

func (g *KVGraph) HandleStaleKeys(changedMembers []membertable.ID, dropped bool) {
    // TODO change to account for multiple replicas
    // TODO maybe do repairs here?


    me := g.findLocalNode()
    verts := make([]*Vertex, 0, 10)
    verts = append(verts, g.NodeIndex...)
    if dropped {
        for _, id := range changedMembers {
            newVertex := Vertex{ Addr: "nowhere",
                                 Hash : HashedKey(id.Hashed()),
                                 LocalNode : nil }
            verts = append(verts, &newVertex)
        }
    }
    for _, id := range changedMembers {
        for k, v := range me.LocalNode.KeyValues {
            if shouldHave(k, verts, HashedKey(id.Hashed())) {
                g.Insert(KeyValue{k,v})
            }
            if !shouldHave(k, verts, me.LocalNode.maxHashedKey) && !dropped {
                var success bool
                me.LocalNode.Delete(&k, &success)
            }
        }
    }
}

func (g *KVGraph) RemoveLocalNodes() {
    // Sort nodes into local and non-local
    var filteredNodeIndex []*Vertex
    var localNodes []*KVNode
    for _, v := range g.NodeIndex {
        if v.LocalNode == nil {
            filteredNodeIndex = append(filteredNodeIndex, v)
        } else {
            localNodes = append(localNodes, v.LocalNode)
        }
    }

    // Set the node index to be only remote nodes and redistribute the keys
    g.NodeIndex = filteredNodeIndex
    for _, node := range localNodes {
        for k, v := range node.KeyValues {
            if err := g.Insert(KeyValue{k,v}); err != nil {
                log.Printf("redis error: %v", err)
            }
        }
    }
}
