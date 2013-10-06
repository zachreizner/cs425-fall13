package testmembertable

import (
    "membertable"
    "testing"
    "time"
)

func TestJoinMember(t *testing.T) {
    var table membertable.Table
    table.Init()

    var mem membertable.Member
    mem.ID = 1
    mem.Name = "Cool"
    mem.Address = "107.11.112.1:8888"
    mem.HeartbeatID = 1

    table.JoinMember(&mem)

    failed, exists := table.IsFailed[1]

    if !exists {
        t.Error("Did not add the member to the map")
    }

    if failed {
        t.Error("Member 1 added, but initialized wrong")
    }
}

func TestMerge(t *testing.T) {
    var table membertable.Table
    table.Init()

    inputArray := make([]membertable.Member, 4)

    // try to add members
    var temp membertable.Member
    temp.ID = 0
    temp.Name = "alpha"
    temp.Address = "1.1.1.1"
    temp.HeartbeatID = 1
    inputArray[0] = temp

    temp.ID = 1
    temp.Name = "beta"
    temp.Address = "1.1.1.2"
    temp.HeartbeatID = 1
    inputArray[1] = temp

    temp.ID = 2
    temp.Name = "delta"
    temp.Address = "1.1.1.3"
    temp.HeartbeatID = 1
    inputArray[2] = temp

    temp.ID = 3
    temp.Name = "gamma"
    temp.Address = "1.1.1.4"
    temp.HeartbeatID = 1
    inputArray[3] = temp

    start := membertable.StampNow()
    table.MergeTables(inputArray)
    end := membertable.StampNow()

    for id := range table.Members {
        if start > table.GetTime(id) {
            t.Error("Added member before the merge")
        }

        if end < table.GetTime(id) {
            t.Error("Added member after the merge")
        }
    }

    _, exists0 := table.Members[0]
    _, exists1 := table.Members[1]
    _, exists2 := table.Members[2]
    _, exists3 := table.Members[3]

    if !exists0 {
        t.Error("Member 0 not added")
    }

    if !exists1 {
        t.Error("Member 1 not added")
    }

    if !exists2 {
        t.Error("Member 2 not added")
    }

    if !exists3 {
        t.Error("Member 3 not added")
    }

    // try to update memebers
    temp.ID = 0
    temp.Name = "alpha"
    temp.Address = "1.1.1.1"
    temp.HeartbeatID = 5
    inputArray[0] = temp

    temp.ID = 1
    temp.Name = "beta"
    temp.Address = "1.1.1.2"
    temp.HeartbeatID = 6
    inputArray[1] = temp

    temp.ID = 2
    temp.Name = "delta"
    temp.Address = "1.1.1.3"
    temp.HeartbeatID = 0
    inputArray[2] = temp

    temp.ID = 3
    temp.Name = "gamma"
    temp.Address = "1.1.1.4"
    temp.HeartbeatID = 0
    inputArray[3] = temp

    start2 := membertable.StampNow()
    table.MergeTables(inputArray)
    end2 := membertable.StampNow()

    if start2 > table.GetTime(0) || end2 < table.GetTime(0) {
        t.Error("ID 0 not updated")
    }

    if table.Members[0].HeartbeatID != 5 {
        t.Error("ID 0 heartbeat not updated")
    }

    if start2 > table.GetTime(1) || end2 < table.GetTime(1) {
        t.Error("ID 1 not updated")
    }

    if start > table.GetTime(2) || end < table.GetTime(2) {
        t.Error("ID 2 time not in the first merge")
    }

    if start > table.GetTime(3) || end < table.GetTime(3) {
        t.Error("ID 3 time not if the first merge")
    }
}

func TestRemoveDead(t *testing.T) {
    var table membertable.Table
    table.Init()

    var temp membertable.Member

    temp.ID = 0
    temp.Name = "alpha"
    temp.Address = "1.1.1.1"
    temp.HeartbeatID = 0
    table.MergeMember(temp)

    temp.ID = 1
    temp.Name = "beta"
    temp.Address = "1.1.1.2"
    temp.HeartbeatID = 0
    table.MergeMember(temp)

    time.Sleep(time.Duration(membertable.TFail) * time.Nanosecond)

    temp.ID = 3
    temp.Name = "delta"
    temp.Address = "1.1.1.3"
    temp.HeartbeatID = 0
    table.MergeMember(temp)

    table.RemoveDead()

    if _, exists0 := table.Members[0]; exists0 && !table.IsFailed[0] {
        t.Error("Member0 is still in the table")
    }

    if _, exists1 := table.Members[1]; exists1 && !table.IsFailed[1]{
        t.Error("Member1 is still in the table")
    }

    if _, exists3 := table.Members[3]; !exists3  || table.IsFailed[3] {
        t.Error("Member3 is not in the table, but should be")
    }
}

func TestIsFailed(t *testing.T) {
    var table membertable.Table
    table.Init()

    var temp membertable.Member

    temp.ID = 0
    temp.Name = "alpha"
    temp.Address = "1.1.1.1"
    temp.HeartbeatID = 0
    table.MergeMember(temp)

    temp.ID = 1
    temp.Name = "beta"
    temp.Address = "1.1.1.2"
    temp.HeartbeatID = 0
    table.MergeMember(temp)

    if table.IsDead(0) {
        t.Error("ID 0 died early")
    }

    if table.IsDead(1) {
        t.Error("ID 0 died early")
    }

    time.Sleep(time.Duration(membertable.TFail))
    table.RemoveDead()

    if !table.IsDead(0) {
        t.Error("ID 0 should be dead")
    }

    if !table.IsDead(1) {
        t.Error("ID 1 should be dead")
    }
}
