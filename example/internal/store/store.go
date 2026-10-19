// Package store 提供 example 工程的内存版伪数据源。
//
// 目的:让 logic 层有真实可跑的读写对象,而不必引入数据库。
// 这里是手写代码,生成器不会修改本包。
//
// 之所以单独成包而不写在 logic 里:logic 层文件是生成器产物,
// 每次重新生成都会被覆盖;把伪实现放进来会被冲掉。
package store

import (
	"reflect"
	"sort"
	"sync"
	"time"
)

// ============ 通用内存仓库 ============

// Repo 泛型内存仓库,用读写锁保证并发安全
type Repo[T any] struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]T
}

// NewRepo 建一个空仓库
func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{items: make(map[int64]T)}
}

// Seed 预置数据并把 seq 推到最大 id
func (r *Repo[T]) Seed(list ...T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, it := range list {
		if id := idOf(it); id > r.seq {
			r.seq = id
		}
		r.items[idOf(it)] = it
	}
}

// NextId 分配下一个自增 id
func (r *Repo[T]) NextId() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	return r.seq
}

// Put 写入或覆盖
func (r *Repo[T]) Put(id int64, v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[id] = v
}

// Get 按 id 读取
func (r *Repo[T]) Get(id int64) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[id]
	return v, ok
}

// Has 判断是否存在
func (r *Repo[T]) Has(id int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.items[id]
	return ok
}

// Del 删除
func (r *Repo[T]) Del(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return false
	}
	delete(r.items, id)
	return true
}

// Len 元素个数
func (r *Repo[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// All 返回全部元素(浅拷贝切片)
func (r *Repo[T]) All() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]T, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, v)
	}
	return out
}

// Filter 按条件过滤并按 id 升序返回
func (r *Repo[T]) Filter(keep func(T) bool) []T {
	all := r.All()
	out := make([]T, 0, len(all))
	for _, v := range all {
		if keep == nil || keep(v) {
			out = append(out, v)
		}
	}
	SortByID(out)
	return out
}

// Update 存在时回调修改,返回是否存在
func (r *Repo[T]) Update(id int64, apply func(T) T) (T, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.items[id]
	if !ok {
		var zero T
		return zero, false
	}
	nv := apply(v)
	r.items[id] = nv
	return nv, true
}

// SortByID 按 Id 字段升序排序
func SortByID[T any](list []T) {
	sort.SliceStable(list, func(i, j int) bool {
		return idOf(list[i]) < idOf(list[j])
	})
}

// idOf 取实体的 Id 字段。用反射而非给每个实体加方法,少写样板代码
func idOf[T any](v T) int64 {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("Id")
	if !f.IsValid() {
		return 0
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.Int()
	}
	return 0
}

// ============ 各领域实体 ============

// User 用户
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

// Product 商品
type Product struct {
	Id       int64
	Title    string
	Price    float64
	Stock    int
	Tags     []string
	Images   []string
	Desc     string
	Category int64
}

// Order 订单
type Order struct {
	Id       int64
	UserId   int64
	Amount   float64
	Status   string
	Remark   string
	Receiver string
	Phone    string
	Addr     string
	CreateAt time.Time
}

// Article 文章
type Article struct {
	Id        int64
	Title     string
	Content   string
	AuthorId  int64
	Tags      []string
	ReadCount int64
	PublishAt time.Time
	LastEdit  time.Duration
}

// Notify 通知
type Notify struct {
	Id       int64
	UserId   int64
	Title    string
	Content  string
	Read     bool
	CreateAt time.Time
}

// FileInfo 文件元信息
type FileInfo struct {
	Id       int64
	Name     string
	Size     int64
	Url      string
	UploadAt time.Time
}

// Threshold 阈值
type Threshold struct {
	Id    int64
	Name  string
	Score string
}

// Skill 角色技能
type Skill struct {
	Id    int64
	Name  string
	Level []int
}

// Profile 用户资料
type Profile struct {
	Id        int64
	Nickname  string
	Avatar    string
	Level     int
	Tags      []string
	Gender    string
	Signature string
	CreatedAt time.Time
}

// Token 令牌
type Token struct {
	Value    string
	Username string
	ExpireAt time.Time
}

// SmsCode 短信验证码
type SmsCode struct {
	Phone     string
	Scene     string
	Code      string
	ExpireAt  time.Time
	CreatedAt time.Time
}
