package logic

import (
	"strings"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// ArticleService 文章业务逻辑(伪实现:数据存内存)
type ArticleService struct {
	articles *store.ArticleStore
}

// NewArticleService 创建文章服务
func NewArticleService() *ArticleService {
	return &ArticleService{articles: store.NewArticleStore()}
}

// ListArticles 文章列表,支持按作者与标签过滤
func (s *ArticleService) ListArticles(c *gin.Context, req request.ListArticlesReq) (res.ListArticlesRes, error) {
	if req.AuthorId < 0 {
		return res.ListArticlesRes{}, validate.Wrap(validate.OutOfRange, "authorId 不能为负数")
	}
	tag := strings.TrimSpace(req.Tag)
	filtered := s.articles.Filter(func(a store.Article) bool {
		if req.AuthorId > 0 && a.AuthorId != req.AuthorId {
			return false
		}
		if tag != "" && !containsString(a.Tags, tag) {
			return false
		}
		return true
	})

	out := res.ListArticlesRes{Total: int64(len(filtered)), List: make([]*res.ListArticlesResListItem, 0, len(filtered))}
	for _, a := range filtered {
		summary := []rune(a.Content)
		if len(summary) > 20 {
			summary = summary[:20]
		}
		out.List = append(out.List, &res.ListArticlesResListItem{
			Id: a.Id, Title: a.Title, Summary: string(summary), ReadCount: a.ReadCount,
		})
	}
	return out, nil
}

// GetArticleById 文章详情。顺带把阅读数加一
func (s *ArticleService) GetArticleById(c *gin.Context, id int64) (res.GetArticleByIdRes, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return res.GetArticleByIdRes{}, err
	}
	a, ok := s.articles.Get(id)
	if !ok {
		return res.GetArticleByIdRes{}, validate.NotFoundErr("文章", id)
	}
	s.articles.Update(id, func(x store.Article) store.Article {
		x.ReadCount++
		return x
	})
	return res.GetArticleByIdRes{
		Id: a.Id, Title: a.Title, Content: a.Content, AuthorId: a.AuthorId,
		PublishAt: a.PublishAt, LastEdit: a.LastEdit,
	}, nil
}

// PublishArticle 发布文章
func (s *ArticleService) PublishArticle(c *gin.Context, req request.PublishArticleReq) (int64, error) {
	if err := validate.Required("title", req.Title); err != nil {
		return 0, err
	}
	if err := validate.Required("content", req.Content); err != nil {
		return 0, err
	}
	if len([]rune(req.Title)) > 100 {
		return 0, validate.Wrap(validate.Invalid, "title 最多 100 个字符")
	}
	if len(req.Tags) > 10 {
		return 0, validate.Wrap(validate.Invalid, "tags 最多 10 个")
	}
	for _, t := range req.Tags {
		if strings.TrimSpace(t) == "" {
			return 0, validate.Wrap(validate.Invalid, "tags 不能包含空字符串")
		}
	}

	id := s.articles.NextId()
	s.articles.Put(id, store.Article{
		Id: id, Title: req.Title, Content: req.Content,
		AuthorId: 1, Tags: req.Tags, PublishAt: time.Now(),
	})
	return id, nil
}

// GetArticleStats 统计文章数据
func (s *ArticleService) GetArticleStats(c *gin.Context) (res.GetArticleStatsRes, error) {
	st := s.collectStats()
	return res.GetArticleStatsRes{
		Total: st.Total, Today: st.Today, Ratio: st.Ratio,
		Online: st.Online, AvgRead: st.AvgRead, Cost: st.Cost, Extra: "v1",
	}, nil
}

// GetArticleStats1 统计文章数据1。
// 注意:该接口与 GetArticleStats 的 path 完全相同,是 yaml 里重复声明同一 path
// 导致生成器产出了重复路由,属于生成器需要改进的地方(详见测试报告)。
func (s *ArticleService) GetArticleStats1(c *gin.Context) (res.GetArticleStats1Res, error) {
	st := s.collectStats()
	return res.GetArticleStats1Res{
		Total: st.Total, Today: st.Today, Ratio: st.Ratio,
		Online: st.Online, AvgRead: st.AvgRead, Cost: st.Cost, Extra: "v1",
	}, nil
}

// GetArticleStats2 统计文章数据2。同上,path 重复
func (s *ArticleService) GetArticleStats2(c *gin.Context) (res.GetArticleStats2Res, error) {
	st := s.collectStats()
	return res.GetArticleStats2Res{
		Total: st.Total, Today: st.Today, Ratio: st.Ratio,
		Online: st.Online, AvgRead: st.AvgRead, Cost: st.Cost, Extra: "v1",
	}, nil
}

// stats 统计结果中间结构
type stats struct {
	Total   int64
	Today   int64
	Ratio   float64
	Online  bool
	AvgRead float32
	Cost    time.Duration
}

// collectStats 汇总统计,三个 Stats 接口共用
func (s *ArticleService) collectStats() stats {
	start := time.Now()
	all := s.articles.Filter(nil)

	today := int64(0)
	since := time.Now().Add(-24 * time.Hour)
	var sumRead int64
	for _, a := range all {
		if a.PublishAt.After(since) {
			today++
		}
		sumRead += a.ReadCount
	}

	var avg float32
	var ratio float64
	if len(all) > 0 {
		avg = float32(sumRead) / float32(len(all))
		ratio = float64(today) / float64(len(all))
	}
	return stats{
		Total: int64(len(all)), Today: today, Ratio: ratio,
		Online: len(all) > 0, AvgRead: avg, Cost: time.Since(start),
	}
}

// containsString 判断切片是否包含指定字符串(忽略大小写)
func containsString(list []string, want string) bool {
	for _, v := range list {
		if strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}
