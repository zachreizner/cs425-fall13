package membertable

import (
    "os"
    "io"
    "log"
    "time"
    "net/rpc"
    "math/rand"
    "encoding/gob"
    "encoding/binary"
)

const TFail = Timestamp(2 * time.Second)
const TDrop = Timestamp(3 * TFail)

type IDNum int32

type ID struct{
    Num IDNum
    Name string
    Address string
}

func (id *ID) GetColor() string {
    num := uint(id.Num) + dHash(id.Address) + dHash(id.Name)
    switch num % 6 {
        case 0: return "1;31";
        case 1: return "1;32";
        case 2: return "1;34";
        case 3: return "1;33";
        case 4: return "1;35";
        case 5: return "1;36";
    }
    return "0";
}

// djb2 hashing algorithm
func dHash(s string) uint {
    hash := uint(5381)
    var c uint
    i := uint(0)
    for i < uint(len(s)) {
        hash = ((hash << 5) + hash) + c
        i++
    }
    return hash
}

type Timestamp int64 // time.Now().UnixNano()

type Member struct {
    ID ID
    HeartbeatID int64
    TimeStamp Timestamp
    IsFailed bool
}

type Table struct {
    Members map[ID]Member
    myID ID
}

// initialize the Table struct with only myself as a member
// timestamp is time of init
// HeartbeatID is set to 0
// IsFailed is false
func (t *Table) Init(me ID) {
    t.Members = make(map[ID]Member)
    t.myID = me

    member := Member{
        ID: me,
        HeartbeatID: 0,
        TimeStamp: StampNow(),
        IsFailed: false,
    }
    t.Members[me] = member
}

func (t *Table) GetTime(id ID) Timestamp {
    return t.Members[id].TimeStamp
}

func (t *Table) IsDead(id ID) bool {
    mem, exists := t.Members[id]

    return !exists || mem.IsFailed
}

// returns a timestamp for the current time when called
func StampNow() Timestamp {
    return Timestamp(time.Now().UnixNano())
}

func (t *Table) JoinMember(m *Member) {
    // set m.LastHeartbeat to now
    // add m to t.Members
    log.Println("Adding member: Id=", m.ID.Num, ", name=", m.ID.Name)
    m.TimeStamp = StampNow()
    m.IsFailed = false
    t.Members[m.ID] = *m
}

func (t *Table) HeartbeatMember(id ID) {
    // update the timestamp of the member of the given id
    mem, exists := t.Members[id]

    if !exists {
        log.Println("Tried to update timestamp of a nonmember")
    }

    if mem.IsFailed {
        log.Println("Tried to update timestamp of a failed member")
    }
    mem.TimeStamp = StampNow()
    t.Members[id] = mem
}

func (t *Table) dropMember(id ID) {
    delete(t.Members, id)
}

func (t *Table) RemoveDead() {
    // remove dead members
    for id, mem := range t.Members {
        curTime := StampNow()
        time := mem.TimeStamp
        if !mem.IsFailed && curTime - time > TFail {
            // process not heard from, mark as failed
            log.Println("member", id, "has failed")
            mem.IsFailed = true
            t.Members[id] = mem
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
        if !member.IsFailed {
            memberArray[index] = member
            index += 1
        }
    }
    return memberArray[0:index]
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
        failed := myInfo.IsFailed
        if myInfo.HeartbeatID < member.HeartbeatID  && !failed {
            myInfo.HeartbeatID = member.HeartbeatID
            myInfo.TimeStamp = StampNow()
            t.Members[member.ID] = myInfo
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

func (t *Table) RpcUpdate(members []Member, dummy *int) error {
    // a second parameter as a pointer is needed, but i have no use for it
    defer t.RemoveDead()
    t.MergeTables(members)
    *dummy = 0
    return nil
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

func IncrementIDFile(filename string) (IDNum, error) {
    f, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE, os.ModePerm)
    if err != nil {
        return 0, err
    }

    defer f.Close()

    fi, err := f.Stat()
    if err != nil {
        return 0, err
    }

    // A zero file size means there was no ID file, so we should create one with the second ID
    // (which is 1) and return the starting ID (which is 1).
    if fi.Size() == 0 {
        err = binary.Write(f, binary.BigEndian, IDNum(1))
        return 0, err
    }

    // Read the id which we will return
    var id IDNum
    if err = binary.Read(f, binary.BigEndian, &id); err != nil {
        return 0, err
    }

    // Move to the beginning of the file and erase it
    if _, err = f.Seek(0, 0); err != nil {
        return id, err
    }

    if err = f.Truncate(0); err != nil {
        return id, err
    }

    // Write out what the next id should be, but do not return it
    err = binary.Write(f, binary.BigEndian, IDNum(id + 1))
    return id, err
}

////////////// Heartbeating ////////////////////
func (t *Table) SendHeartbeatToAddress(addr string) error {
    // Connect to the given member
    client, err := rpc.DialHTTP("tcp", addr)
    if err != nil {
        return err
    }

    defer client.Close()

    var reply int
    data := t.ActiveMembers()
    callErr := client.Call("Table.RpcUpdate", data, &reply)
    if callErr != nil {
        log.Print("Error while sending heardbeat")
        return callErr
    }
    return nil
}

func (t *Table) SendHeartbeat() error {
    // Get a list of members we can send our hearbeat to
    memberList := t.ActiveMembers()

    // We are alone on this earth :(
    if len(memberList) == 0 || (len(memberList) == 1 && memberList[0].ID == t.myID) {
        log.Println("So allooone")
        return nil
    }

    // Choose a member at random and send their heartbeat
    var sendToMember *Member
    for sendToMember == nil || sendToMember.ID == t.myID {
        sendToMember = &memberList[rand.Int() % len(memberList)]
    }

    return t.SendHeartbeatToAddress(sendToMember.ID.Address)
}

func (t *Table) SendHeartbeatProcess(fatalChan chan bool) {
    for {
        mem := t.Members[t.myID]
        mem.HeartbeatID++
        t.MergeMember(mem)
        err := t.SendHeartbeat()
        if err != nil {
            log.Println(err)
        }
        time.Sleep(100 * time.Millisecond)
    }
    fatalChan <- true
}

