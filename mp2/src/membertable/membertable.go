package membertable

import (
    "io"
)

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
}

func (t *Table) JoinMember(m *Member) {
    // set m.LastHeartbeat to now
    // add m to t.Members
}

func (t *Table) HeartbeatMember(id ID) {
    // update the timestamp of the member of the given id
}

func (t *Table) RemoveDead() {
    // remove dead members
}

func (t *Table) Write(w io.Writer) error {
    // remove the dead
    // Write out t.Members as an array using gob. Might require converting the map to an array
}

func (t *Table) mergeTables(members []Member) {
    // apply the offsets of timeOffsetss to the members array
}

func (t *Table) Update(id ID, r io.Reader) error {
    // read the input of a Table.Write
    // apply offsets given that this came from id
    // merge the results into t.Members; beware of timestamps in the future
    // remove the dead
}