package mykv

type Vertex struct {
    Addr string
    Hash HashedKey
    LocalNode *KVNode
}

type KVGraph struct {
    NodeIndex []*Vertex
    Connector RPCConnector
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
        return g.NodeIndex[len(g.NodeIndex) - idx]
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