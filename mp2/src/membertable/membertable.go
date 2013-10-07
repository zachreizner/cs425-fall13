package membertable

import (
    "io"
    "log"
    "time"
    "encoding/gob"
)

const TFail = Timestamp(2 * time.Second)
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
    IsFailed map[ID]bool
}

// initialize the Table struct to be empty
func (t *Table) Init() {
    t.Members = make(map[ID]Member)
    t.timeStamps = make(map[ID]Timestamp)
    t.IsFailed = make(map[ID]bool)
}

func (t *Table) GetTime(id ID) Timestamp {
    return t.timeStamps[id]
}

func (t *Table) IsDead(id ID) bool {
    failed, exists := t.IsFailed[id]

    return !exists || failed
}

// returns a timestamp for the current time when called
func StampNow() Timestamp {
    return Timestamp(time.Now().UnixNano())
}

func (t *Table) JoinMember(m *Member) {
    // set m.LastHeartbeat to now
    // add m to t.Members
    log.Println("Adding member: Id=", m.ID, ", name=", m.Name)
    t.Members[m.ID] = *m
    t.timeStamps[m.ID] = StampNow()
    t.IsFailed[m.ID] = false
}

func (t *Table) HeartbeatMember(id ID) {
    // update the timestamp of the member of the given id
    _, exists := t.timeStamps[id]

    if !exists {
        log.Println("Tried to update timestamp of a nonmember")
    }

    if failed := t.IsFailed[id]; failed {
        log.Println("Tried to update timestamp of a failed member")
    }

    t.timeStamps[id] = StampNow()
}

func (t *Table) dropMember(id ID) {
    delete(t.Members, id)
    delete(t.timeStamps, id)
    delete(t.IsFailed, id)
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
            t.IsFailed[id] = true
        }

        if curTime - time > TDrop {
            t.dropMember(id)
        }
    }
}

func (t *Table) ActiveMembers() []Member {
    t.RemoveDead()
    memberArray := make([]Member, len(t.Members))
    index := 0
    for _, member := range t.Members {
        if !t.IsFailed[member.ID] {
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

func (t *Table) MergeMember(member Member) {
    myInfo, exists := t.Members[member.ID]
    if exists {
        failed := t.IsFailed[member.ID]
        if myInfo.HeartbeatID < member.HeartbeatID  && !failed {
            myInfo.HeartbeatID = member.HeartbeatID
            t.Members[member.ID] = myInfo
            t.timeStamps[member.ID] = StampNow()
        }
    } else {
        t.JoinMember(&member)
    }
}

func (t *Table) MergeTables(members []Member) {
    // apply the offsets of timeOffsetss to the members array
    for _, member := range members {
        t.MergeMember(member)
    }
}

func (t *Table) Update(r io.Reader) error {
    // read the input of a Table.Write
    // merge the results into t.Members; beware of timestamps in the future
    // remove the dead
    defer t.RemoveDead()

    dec := gob.NewDecoder(r)

    var memberArray []Member
    err := dec.Decode(&memberArray)
    if err != nil {
        return err
    }

    t.MergeTables(memberArray)
    return nil
}
