package membertable

import (
    "io"
    "log"
    "time"
    "encoding/gob"
)

const TFail = Timestamp(1000)
const TDrop = Timestamp(3 * TFail)

type ID int32
type Timestamp int64 // time.Now().UnixNano()

type Member struct {
    ID ID
    Name string
    Address string
    HeartbeatID int64
}

type Table struct {
    Members map[ID]Member
    timeStamps map[ID]Timestamp
    isFailed map[ID]bool
}

// returns a timestamp for the current time when called
func StampNow() Timestamp {
    return Timestamp(time.Now().UnixNano())
}

func (t *Table) JoinMember(m *Member) {
    // set m.LastHeartbeat to now
    // add m to t.Members
    t.Members[m.ID] = *m
    t.timeStamps[m.ID] = StampNow()
    t.isFailed[m.ID] = false
}

func (t *Table) HeartbeatMember(id ID) {
    // update the timestamp of the member of the given id
    _, exists := t.timeStamps[id]

    if !exists {
        log.Println("Tried to update timestamp of a nonmember")
    }

    if failed := t.isFailed[id]; failed {
        log.Println("Tried to update timestamp of a failed member")
    }

    t.timeStamps[id] = StampNow()
}

func (t *Table) RemoveDead() {
    // remove dead members
    for id, _ := range t.Members {
        time, exists := t.timeStamps[id]
        if !exists {
            log.Println("While removing dead, found a member with no timestamp.")
        }
        curTime := StampNow()
        if curTime - time > TFail {
            // process not heard from, mark as failed
            log.Println("member", id, "has failed")
            t.isFailed[id] = true
        }

        if curTime - time > TDrop {
            delete(t.Members, id)
            delete(t.timeStamps, id)
        }
    }
}

func (t *Table) ActiveMembers() []Member {
    memberArray := make([]Member, len(t.Members))
    index := 0
    for _, member := range t.Members {
        if !t.isFailed[member.ID] {
            memberArray[index] = member
            index += 1
        }
    }
    return memberArray
}

func (t *Table) WriteTo(w io.Writer) error {
    // remove the dead
    // Write out t.Members as an array using gob. Might require converting the map to an array
    t.RemoveDead()
    enc := gob.NewEncoder(w)
    data := t.ActiveMembers()
    return enc.Encode(data)
}

func (t *Table) mergeMember(member Member) {
    myInfo, exists := t.Members[member.ID]
    if exists {
        failed := t.isFailed[member.ID]
        if myInfo.HeartbeatID < member.HeartbeatID  && !failed {
            myInfo.HeartbeatID = member.HeartbeatID
            t.Members[member.ID] = myInfo
            t.timeStamps[member.ID] = StampNow()
        }
    } else {
        t.JoinMember(&member)
    }
}
func (t *Table) mergeTables(members []Member) {
    // apply the offsets of timeOffsetss to the members array
    for _, member := range members {
        t.mergeMember(member)
    }
}

func (t *Table) Update(id ID, r io.Reader) error {
    // read the input of a Table.Write
    // apply offsets given that this came from id
    // merge the results into t.Members; beware of timestamps in the future
    // remove the dead
    defer t.RemoveDead()

    dec := gob.NewDecoder(r)

    var memberArray []Member
    err := dec.Decode(memberArray)
    if err != nil {
        return err
    }

    t.mergeTables(memberArray)
    return nil
}
