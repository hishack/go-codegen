package store

import (
	"strings"
	"sync"
	"time"
)

// ============ 各领域仓库 ============

// UserStore 内存用户仓库
type UserStore struct{ Repo[User] }

// NewUserStore 建一个带少量种子数据的仓库,便于直接联调
func NewUserStore() *UserStore {
	now := time.Now()
	s := &UserStore{Repo[User]{items: map[int64]User{}}}
	s.Seed(
		User{
			Id: 1, Name: "张三", Age: 20, Height: 175.5, Weight: 68.5,
			Enabled: true, Score: 100, Level: 3, Count: 12, Ratio: 0.85,
			Birthday: time.Date(2005, 3, 15, 0, 0, 0, 0, time.Local),
			Duration: 3 * time.Hour, Nickname: "法外狂徒",
			Tags: []string{"vip", "active"}, CreatedAt: now.Add(-72 * time.Hour),
		},
		User{
			Id: 2, Name: "李四", Age: 26, Height: 168, Weight: 60,
			Enabled: true, Score: 200, Level: 5, Count: 40, Ratio: 0.6,
			Birthday: time.Date(1999, 7, 1, 0, 0, 0, 0, time.Local),
			Duration: 90 * time.Minute, Nickname: "隔壁老王",
			Tags: []string{"normal"}, CreatedAt: now.Add(-24 * time.Hour),
		},
	)
	return s
}

// Tags 读取用户标签
func (s *UserStore) Tags(id int64) []string {
	u, ok := s.Get(id)
	if !ok {
		return nil
	}
	return append([]string(nil), u.Tags...)
}

// SetAvatar 记录头像地址
func (s *UserStore) SetAvatar(id int64, url string) bool {
	_, ok := s.Update(id, func(u User) User {
		u.Avatar = url
		return u
	})
	return ok
}

// ProductStore 商品仓库
type ProductStore struct{ Repo[Product] }

// NewProductStore 带种子数据
func NewProductStore() *ProductStore {
	s := &ProductStore{Repo[Product]{items: map[int64]Product{}}}
	s.Seed(
		Product{Id: 1, Title: "机械键盘", Price: 399, Stock: 20, Tags: []string{"数码"}, Images: []string{"/img/kb.png"}, Desc: "87 键", Category: 1},
		Product{Id: 2, Title: "人体工学椅", Price: 1299, Stock: 5, Tags: []string{"家具"}, Images: []string{"/img/chair.png"}, Desc: "久坐不累", Category: 1},
		Product{Id: 3, Title: "咖啡豆", Price: 88.5, Stock: 100, Tags: []string{"食品", "饮料"}, Desc: "中深烘", Category: 2},
	)
	return s
}

// OrderStore 订单仓库
type OrderStore struct{ Repo[Order] }

// NewOrderStore 带种子数据
func NewOrderStore() *OrderStore {
	now := time.Now()
	s := &OrderStore{Repo[Order]{items: map[int64]Order{}}}
	s.Seed(
		Order{Id: 1001, UserId: 1, Amount: 399, Status: "unpaid", Receiver: "张三", Phone: "13800000000", Addr: "杭州市西湖区", CreateAt: now.Add(-2 * time.Hour)},
		Order{Id: 1002, UserId: 1, Amount: 1299, Status: "paid", Receiver: "张三", Phone: "13800000000", Addr: "杭州市西湖区", CreateAt: now.Add(-24 * time.Hour)},
		Order{Id: 1003, UserId: 2, Amount: 88.5, Status: "shipped", Receiver: "李四", Phone: "13900000000", Addr: "上海市浦东新区", CreateAt: now.Add(-48 * time.Hour)},
	)
	return s
}

// ArticleStore 文章仓库
type ArticleStore struct{ Repo[Article] }

// NewArticleStore 带种子数据
func NewArticleStore() *ArticleStore {
	now := time.Now()
	s := &ArticleStore{Repo[Article]{items: map[int64]Article{}}}
	s.Seed(
		Article{Id: 1, Title: "Go 泛型实战", Content: "正文内容 A", AuthorId: 1, Tags: []string{"go", "后端"}, ReadCount: 120, PublishAt: now.Add(-72 * time.Hour), LastEdit: 2 * time.Hour},
		Article{Id: 2, Title: "Gin 最佳实践", Content: "正文内容 B", AuthorId: 1, Tags: []string{"go", "web"}, ReadCount: 300, PublishAt: now.Add(-96 * time.Hour), LastEdit: time.Hour},
		Article{Id: 3, Title: "性能调优手册", Content: "正文内容 C", AuthorId: 2, Tags: []string{"性能"}, ReadCount: 50, PublishAt: now.Add(-1 * time.Hour)},
	)
	return s
}

// NotifyStore 通知仓库
type NotifyStore struct{ Repo[Notify] }

// NewNotifyStore 带种子数据
func NewNotifyStore() *NotifyStore {
	now := time.Now()
	s := &NotifyStore{Repo[Notify]{items: map[int64]Notify{}}}
	s.Seed(
		Notify{Id: 1, UserId: 1, Title: "订单提醒", Content: "您的订单已发货", Read: false, CreateAt: now.Add(-1 * time.Hour)},
		Notify{Id: 2, UserId: 1, Title: "活动通知", Content: "双十一大促开启", Read: true, CreateAt: now.Add(-10 * time.Hour)},
		Notify{Id: 3, UserId: 2, Title: "安全提醒", Content: "异地登录", Read: false, CreateAt: now.Add(-30 * time.Minute)},
	)
	return s
}

// FileStore 文件仓库
type FileStore struct{ Repo[FileInfo] }

// NewFileStore 空仓库
func NewFileStore() *FileStore { return &FileStore{Repo[FileInfo]{items: map[int64]FileInfo{}}} }

// ThresholdStore 阈值仓库
type ThresholdStore struct{ Repo[Threshold] }

// NewThresholdStore 带种子数据
func NewThresholdStore() *ThresholdStore {
	s := &ThresholdStore{Repo[Threshold]{items: map[int64]Threshold{}}}
	s.Seed(
		Threshold{Id: 1, Name: "满 100 减 10", Score: "easy"},
		Threshold{Id: 2, Name: "满 500 减 50", Score: "normal"},
	)
	return s
}

// ProfileStore 资料仓库
type ProfileStore struct{ Repo[Profile] }

// NewProfileStore 带种子数据
func NewProfileStore() *ProfileStore {
	now := time.Now()
	s := &ProfileStore{Repo[Profile]{items: map[int64]Profile{}}}
	s.Seed(Profile{
		Id: 1, Nickname: "法外狂徒", Level: 3,
		Tags: []string{"vip", "active"}, Gender: "male",
		Signature: "开心就好", CreatedAt: now.Add(-720 * time.Hour),
	})
	return s
}

// SkillStore 技能仓库(按角色 id 存技能列表)
type SkillStore struct {
	mu     sync.RWMutex
	byChar map[int64][]Skill
}

// NewSkillStore 带种子数据
func NewSkillStore() *SkillStore {
	return &SkillStore{byChar: map[int64][]Skill{
		1: {
			{Id: 101, Name: "火球术", Level: []int{1, 2}},
			{Id: 102, Name: "冰霜护甲", Level: []int{1}},
		},
		2: {
			{Id: 201, Name: "闪电链", Level: []int{1, 2, 3}},
		},
	}}
}

// List 返回某角色的技能列表
func (s *SkillStore) List(charId int64) ([]Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list, ok := s.byChar[charId]
	if !ok {
		return nil, false
	}
	out := make([]Skill, len(list))
	copy(out, list)
	return out, true
}

// TokenStore 令牌仓库
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]Token
}

// NewTokenStore 空仓库
func NewTokenStore() *TokenStore { return &TokenStore{tokens: map[string]Token{}} }

// Save 保存令牌
func (s *TokenStore) Save(t Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[t.Value] = t
}

// Get 按令牌值查询
func (s *TokenStore) Get(value string) (Token, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[value]
	return t, ok
}

// Del 删除令牌
func (s *TokenStore) Del(value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tokens[value]; !ok {
		return false
	}
	delete(s.tokens, value)
	return true
}

// SmsStore 验证码仓库
type SmsStore struct {
	mu    sync.RWMutex
	codes map[string]SmsCode
}

// NewSmsStore 空仓库
func NewSmsStore() *SmsStore { return &SmsStore{codes: map[string]SmsCode{}} }

// Save 保存验证码
func (s *SmsStore) Save(c SmsCode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[c.Phone+"|"+c.Scene] = c
}

// Get 按手机号+场景查询
func (s *SmsStore) Get(phone, scene string) (SmsCode, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.codes[strings.TrimSpace(phone)+"|"+strings.TrimSpace(scene)]
	return c, ok
}
