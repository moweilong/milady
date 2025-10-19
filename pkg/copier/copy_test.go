package copier

import (
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

type MyUser1 struct {
	Id        uint64
	MyIp      string
	OrderId   int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type MyUser2 struct {
	ID        int64
	MyIP      string
	OrderID   uint32
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type UserReply struct {
	Id        int
	MyIp      string
	OrderId   int
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}

func copyCustom1() (*UserReply, *UserReply) {
	user1 := &MyUser1{
		Id:        123,
		MyIp:      "127.0.0.1",
		OrderId:   888,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Hour),
		DeletedAt: time.Now().Add(-2 * time.Hour),
	}
	user2 := &MyUser2{
		ID:        456,
		MyIP:      "localhost",
		OrderID:   999,
		CreatedAt: &user1.CreatedAt,
		UpdatedAt: &user1.UpdatedAt,
		DeletedAt: &user1.DeletedAt,
	}
	reply1 := &UserReply{}
	reply2 := &UserReply{}

	Copy(reply1, user1)
	Copy(reply2, user2)

	return reply1, reply2
}

func copyStandard1() (*UserReply, *UserReply) {
	user1 := &MyUser1{
		Id:        789,
		MyIp:      "127.0.0.1",
		OrderId:   888,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Hour),
		DeletedAt: time.Now().Add(-2 * time.Hour),
	}
	user2 := &MyUser2{
		ID:        1010,
		MyIP:      "localhost",
		OrderID:   999,
		CreatedAt: &user1.CreatedAt,
		UpdatedAt: &user1.UpdatedAt,
		DeletedAt: &user1.DeletedAt,
	}
	reply1 := &UserReply{}
	reply2 := &UserReply{}

	copier.Copy(reply1, user1)
	copier.Copy(reply2, user2)

	return reply1, reply2
}

func copyCustom2() (*MyUser1, *MyUser2) {
	req := &UserReply{
		Id:        123,
		MyIp:      "127.0.0.1",
		OrderId:   888,
		CreatedAt: "2025-05-25T15:38:20+08:00",
		UpdatedAt: "2025-05-25T16:38:20+08:00",
		DeletedAt: "2025-05-25T17:38:20+08:00",
	}
	user1 := &MyUser1{}
	user2 := &MyUser2{}

	Copy(user1, req)
	Copy(user2, req)

	return user1, user2
}

func copyStandard2() (*MyUser1, *MyUser2) {
	req := &UserReply{
		Id:        456,
		MyIp:      "localhost",
		OrderId:   888,
		CreatedAt: "2025-05-25T15:38:20+08:00",
		UpdatedAt: "2025-05-25T16:38:20+08:00",
		DeletedAt: "2025-05-25T17:38:20+08:00",
	}
	user1 := &MyUser1{}
	user2 := &MyUser2{}

	copier.Copy(user1, req)
	copier.Copy(user2, req)

	return user1, user2
}

func TestCopyCustom1(t *testing.T) {
	reply1, reply2 := copyCustom1()
	assert.Equal(t, reply1.Id, 123)
	assert.Equal(t, reply2.Id, 456)
	t.Log(reply1)
	t.Log(reply2)
}

func TestCopyStandard1(t *testing.T) {
	reply1, reply2 := copyStandard1()
	assert.Equal(t, reply1.Id, 789)
	assert.Equal(t, reply2.Id, 1010)
	t.Log(reply1)
	t.Log(reply2)
}

func TestCopyCustom2(t *testing.T) {
	user1, user2 := copyCustom2()
	assert.Equal(t, int(user1.Id), 123)
	assert.Equal(t, int(user2.ID), 123)
	t.Log(user1)
	t.Log(user2)
}

func TestCopyStandard2(t *testing.T) {
	user1, user2 := copyStandard2()
	assert.Equal(t, int(user1.Id), 456)
	assert.Equal(t, int(user2.ID), 456)
	t.Log(user1)
	t.Log(user2)
}

func BenchmarkCopyStandard1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		copyStandard1()
	}
}

func BenchmarkCopyCustom1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		copyCustom1()
	}
}

func BenchmarkCopyStandard2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		copyStandard2()
	}
}

func BenchmarkCopyCustom2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		copyCustom2()
	}
}

// -------------------------------------------------

func SliceData1() []*MyUser1 {
	users := []*MyUser1{
		{
			Id:        123,
			MyIp:      "127.0.0.1",
			OrderId:   888,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now().Add(time.Hour),
			DeletedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Id:        456,
			MyIp:      "localhost",
			OrderId:   999,
			CreatedAt: time.Now().Add(2 * time.Hour),
			UpdatedAt: time.Now().Add(3 * time.Hour),
			DeletedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			Id:        789,
			MyIp:      "192.168.1.1",
			OrderId:   1010,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * time.Hour),
			DeletedAt: time.Now().Add(3 * time.Hour),
		},
	}

	return users
}

func SliceData2() []*MyUser2 {
	now := time.Now()
	updated := now.Add(time.Hour)
	deleted := now.Add(-2 * time.Hour)
	users := []*MyUser2{
		{
			ID:        123,
			MyIP:      "127.0.0.1",
			OrderID:   888,
			CreatedAt: &now,
			UpdatedAt: &updated,
			DeletedAt: &deleted,
		},
		{
			ID:        456,
			MyIP:      "localhost",
			OrderID:   999,
			CreatedAt: &now,
			UpdatedAt: &updated,
			DeletedAt: &deleted,
		},
		{
			ID:        789,
			MyIP:      "192.168.1.1",
			OrderID:   1010,
			CreatedAt: &now,
			UpdatedAt: &updated,
			DeletedAt: &deleted,
		},
	}

	return users
}

func SliceData3() []*UserReply {
	userReplies := []*UserReply{
		{
			Id:        123,
			MyIp:      "127.0.0.1",
			OrderId:   888,
			CreatedAt: "2025-05-25T15:38:20+08:00",
			UpdatedAt: "2025-05-25T16:38:20+08:00",
			DeletedAt: "2025-05-25T17:38:20+08:00",
		},
		{
			Id:        456,
			MyIp:      "localhost",
			OrderId:   999,
			CreatedAt: "2025-05-25T15:38:20+08:00",
			UpdatedAt: "2025-05-25T16:38:20+08:00",
			DeletedAt: "2025-05-25T17:38:20+08:00",
		},
		{
			Id:        789,
			MyIp:      "192.168.1.1",
			OrderId:   1010,
			CreatedAt: "2025-05-25T15:38:20+08:00",
			UpdatedAt: "2025-05-25T16:38:20+08:00",
			DeletedAt: "2025-05-25T17:38:20+08:00",
		},
	}
	return userReplies
}

func TestSliceCopyCustom1(t *testing.T) {
	users := SliceData1()
	replies := make([]*UserReply, len(users))
	Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyStandard1(t *testing.T) {
	users := SliceData1()
	replies := make([]*UserReply, len(users))
	copier.Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyCustom2(t *testing.T) {
	users := SliceData2()
	replies := make([]*UserReply, len(users))
	Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyStandard2(t *testing.T) {
	users := SliceData2()
	replies := make([]*UserReply, len(users))
	copier.Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyCustom3(t *testing.T) {
	users := SliceData3()
	replies := make([]*MyUser1, len(users))
	Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyStandard3(t *testing.T) {
	users := SliceData3()
	replies := make([]*MyUser1, len(users))
	copier.Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyCustom4(t *testing.T) {
	users := SliceData3()
	replies := make([]*MyUser2, len(users))
	Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func TestSliceCopyStandard4(t *testing.T) {
	users := SliceData3()
	replies := make([]*MyUser2, len(users))
	copier.Copy(&replies, &users)
	for _, reply := range replies {
		t.Log(reply)
	}
}

func BenchmarkSliceCopyStandard1(b *testing.B) {
	users := SliceData1()
	for i := 0; i < b.N; i++ {
		replies := make([]*UserReply, len(users))
		copier.Copy(&replies, &users)
	}
}

func BenchmarkSliceCopyStandardSingle1(b *testing.B) {
	users := SliceData1()
	for i := 0; i < b.N; i++ {
		reply := &UserReply{}
		for _, user := range users {
			copier.Copy(reply, user)
		}
	}
}

func BenchmarkSliceCopyCustom1(b *testing.B) {
	users := SliceData1()
	for i := 0; i < b.N; i++ {
		replies := make([]*UserReply, len(users))
		Copy(&replies, &users)
	}
}

func BenchmarkSliceCopyCustomSingle1(b *testing.B) {
	users := SliceData1()
	for i := 0; i < b.N; i++ {
		reply := &UserReply{}
		for _, user := range users {
			Copy(reply, user)
		}
	}
}

// ---------------------------------------------------------------------

func TestCopyDefault(t *testing.T) {
	user1 := &MyUser1{
		Id:        123,
		MyIp:      "127.0.0.1",
		OrderId:   888,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Hour),
		DeletedAt: time.Now().Add(2 * time.Hour),
	}

	user2 := &MyUser2{}

	err := CopyDefault(user2, user1)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, user2.ID, int64(123))

	t.Log(user2)
}

func TestCopyWithOption(t *testing.T) {
	user1 := &MyUser1{
		Id:        123,
		MyIp:      "127.0.0.1",
		OrderId:   888,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(time.Hour),
		DeletedAt: time.Now().Add(2 * time.Hour),
	}

	user2 := &MyUser2{}
	option := copier.Option{
		DeepCopy: true,
	}
	err := CopyWithOption(user2, user1, option)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, user2.ID, int64(123))

	t.Log(user2)
}

func TestCopyEmbedStruct(t *testing.T) {
	type Model struct {
		ID        uint64         `gorm:"column:id;AUTO_INCREMENT;primary_key" json:"id"`
		CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
		UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
		DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
	}

	type User struct {
		Model `gorm:"embedded"`
		Name  string `gorm:"column:name" json:"name"`
		Age   int    `gorm:"column:age" json:"age"`
	}

	user := &User{
		Model: Model{
			ID:        123,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now().Add(time.Hour),
			DeletedAt: gorm.DeletedAt{Time: time.Now().Add(2 * time.Hour)},
		},
		Name: "test",
		Age:  18,
	}

	type UserReply struct {
		Id        uint64 `json:"id"`
		Age       int    `json:"age"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		DeletedAt string `json:"deletedAt"`
	}

	reply := &UserReply{}
	err := Copy(reply, user)
	if err != nil {
		t.Error(err)
		return
	}
	reply.DeletedAt = user.DeletedAt.Time.Format(time.RFC3339)

	t.Log(*reply)
}
