// Package store 提供 example 工程的内存版伪数据源。
//
// 目的:让 logic 层有真实可跑的读写对象,而不必引入数据库。
// 这里是手写代码,生成器不会修改本包。
//
// 之所以单独成包而不写在 logic 里:logic 层文件是生成器产物,
// 每次重新生成都会被覆盖;把伪实现放进来会被冲掉。
package store

import (
	"sort"
	"sync"
	"time"
)

// User 内存中的用户记录
type User struct {
	Id        int64
	Name      string
	Age       int
	Height    float64
	Weight    float32
	Enabled   bool
	Score     int64
	Level     int32
	Count     uint
	Ratio     float64
	Birthday  time.Time
	Duration  time.Duration
	Nickname  string
	Avatar    string
	Tags      []string
	CreatedAt time.Time
}

// UserStore 内存用户仓库,用读写锁保证并发安全
type UserStore struct {
	mu     sync.RWMutex
	seq    int64
	users  map[int64]*User
	tagsOf map[int64][]string
}

// NewUserStore 建一个带少量种子数据的仓库,便于直接联调
func NewUserStore() *UserStore {
	s := &UserStore{
		users:  make(map[int64]*User),
		tagsOf: make(map[int64][]string),
	}
	now := time.Now()
	s.users[1] = &User{
		Id: 1, Name: "张三", Age: 20, Height: 175.5, Weight: 68.5,
		Enabled: true, Score: 100, Level: 3, Count: 12, Ratio: 0.85,
		Birthday: time.Date(2005, 3, 15, 0, 0, 0, 0, time.Local),
		Duration: 3 * time.Hour, Nickname: "法外狂徒",
		CreatedAt: now.Add(-72 * time.Hour),
	}
	s.tagsOf[1] = []string{"vip", "active"}
	s.users[2] = &User{
		Id: 2, Name: "李四", Age: 26, Height: 168, Weight: 60,
		Enabled: true, Score: 200, Level: 5, Count: 40, Ratio: 0.6,
		Birthday: time.Date(1999, 7, 1, 0, 0, 0, 0, time.Local),
		Duration: 90 * time.Minute, Nickname: "隔壁老王",
		CreatedAt: now.Add(-24 * time.Hour),
	}
	s.tagsOf[2] = []string{"normal"}
	s.seq = 2
	return s
}

// NextId 分配下一个自增 id
func (s *UserStore) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return s.seq
}

// Create 写入一条新用户
func (s *UserStore) Create(u *User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.Id] = u
}

// Get 按 id 读取,不存在返回 ok=false
func (s *UserStore) Get(id int64) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// Exists 判断 id 是否存在
func (s *UserStore) Exists(id int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.users[id]
	return ok
}

// List 按 id 升序返回全部用户的浅拷贝
func (s *UserStore) List() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

// Update 覆盖式更新(对应 PUT)
func (s *UserStore) Update(u *User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.Id]; !ok {
		return false
	}
	s.users[u.Id] = u
	return true
}

// Patch 局部更新(对应 PATCH),只改非零字段,返回是否有字段被改动
func (s *UserStore) Patch(id int64, apply func(*User)) (bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return false, false
	}
	apply(u)
	return true, true
}

// Delete 删除一条,返回是否存在
func (s *UserStore) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return false
	}
	delete(s.users, id)
	delete(s.tagsOf, id)
	return true
}

// Tags 读取用户标签
func (s *UserStore) Tags(id int64) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.tagsOf[id]...)
}

// SetAvatar 记录头像地址
func (s *UserStore) SetAvatar(id int64, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return false
	}
	u.Avatar = url
	return true
}
