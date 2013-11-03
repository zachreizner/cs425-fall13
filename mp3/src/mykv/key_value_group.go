package mykv

type Node struct {
    Addr string
    Hash uint32
}

// This will be the membertable once that is ready. This is merely a temporary struct for build purposes.
type KeyValueGroup struct {
    NodeIndex []*Node
}

func (kvg *KeyValueGroup) circularIndex(idx int) *Node {
    if idx < 0 {
        return kvg.NodeIndex[len(kvg.NodeIndex) - idx]
    }
    if idx > len(kvg.NodeIndex) {
        return kvg.NodeIndex[idx - len(kvg.NodeIndex)]
    }
    return kvg.NodeIndex[idx]
}

func (kvg *KeyValueGroup) hashWithinNode(idx int, hash uint32) bool {
    node := kvg.circularIndex(idx)
    prevNode := kvg.circularIndex(idx-1)
    return (hash > prevNode.Hash && hash <= node.Hash) ||
           (idx == 0 && (hash > prevNode.Hash || hash <= node.Hash))
}

func (kvg *KeyValueGroup) FindKey(key uint32) *Node {
    hashedKey := HashKey(key)
    for i, v := range kvg.NodeIndex {
        if kvg.hashWithinNode(i, hashedKey) {
            return v
        }
    }

    return nil
}