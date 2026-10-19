package logic

import (
	"strings"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// ProductService 商品业务逻辑(伪实现)
type ProductService struct {
	products *store.ProductStore
}

// NewProductService 创建商品服务
func NewProductService() *ProductService {
	return &ProductService{products: store.NewProductStore()}
}

// ListProducts 商品列表。支持关键字/分类过滤 + 分页
func (s *ProductService) ListProducts(c *gin.Context, req request.ListProductsReq) (res.ListProductsRes, error) {
	if req.Page < 0 {
		return res.ListProductsRes{}, validate.Wrap(validate.OutOfRange, "page 不能为负数")
	}
	if req.PageSize < 0 {
		return res.ListProductsRes{}, validate.Wrap(validate.OutOfRange, "pageSize 不能为负数")
	}
	if req.PageSize > 100 {
		return res.ListProductsRes{}, validate.Wrap(validate.OutOfRange, "pageSize 最大 100")
	}

	kw := strings.TrimSpace(req.Keyword)
	filtered := s.products.Filter(func(p store.Product) bool {
		if req.CategoryId > 0 && p.Category != req.CategoryId {
			return false
		}
		if kw != "" && !strings.Contains(p.Title, kw) {
			return false
		}
		return true
	})

	total := int64(len(filtered))
	page := validate.Slice(filtered, req.Page, req.PageSize)

	out := res.ListProductsRes{Total: total, List: make([]*res.ListProductsResListItem, 0, len(page))}
	for _, p := range page {
		tags := p.Tags
		if tags == nil {
			tags = []string{}
		}
		out.List = append(out.List, &res.ListProductsResListItem{
			Id: p.Id, Title: p.Title, Price: p.Price, Tags: tags,
		})
	}
	return out, nil
}

// GetProduct 商品详情
func (s *ProductService) GetProduct(c *gin.Context, id int64) (res.GetProductRes, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return res.GetProductRes{}, err
	}
	p, ok := s.products.Get(id)
	if !ok {
		return res.GetProductRes{}, validate.NotFoundErr("商品", id)
	}
	images := p.Images
	if images == nil {
		images = []string{}
	}
	return res.GetProductRes{
		Id: p.Id, Title: p.Title, Price: p.Price, Stock: p.Stock,
		Detail: &res.GetProductResDetail{Images: images, Desc: p.Desc},
	}, nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(c *gin.Context, req request.CreateProductReq) (int64, error) {
	if err := validate.Required("title", req.Title); err != nil {
		return 0, err
	}
	if err := validate.RangeFloat("price", req.Price, 0); err != nil {
		return 0, err
	}
	if req.Price > 1e8 {
		return 0, validate.Wrap(validate.OutOfRange, "price 最大 100000000")
	}
	if len(req.Images) > 9 {
		return 0, validate.Wrap(validate.Invalid, "images 最多 9 张")
	}
	for _, p := range s.products.Filter(nil) {
		if strings.EqualFold(p.Title, req.Title) {
			return 0, validate.Wrap(validate.Existed, "商品标题已存在: "+req.Title)
		}
	}

	id := s.products.NextId()
	p := store.Product{Id: id, Title: req.Title, Price: req.Price, Stock: 100, Images: req.Images}
	if req.Detail != nil {
		p.Desc = req.Detail.Desc
	}
	s.products.Put(id, p)
	return id, nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(c *gin.Context, id int64, req request.UpdateProductReq) (bool, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return false, err
	}
	if err := validate.Required("title", req.Title); err != nil {
		return false, err
	}
	if err := validate.RangeFloat("price", req.Price, 0); err != nil {
		return false, err
	}
	if _, ok := s.products.Get(id); !ok {
		return false, validate.NotFoundErr("商品", id)
	}
	for _, p := range s.products.Filter(nil) {
		if p.Id != id && strings.EqualFold(p.Title, req.Title) {
			return false, validate.Wrap(validate.Existed, "商品标题已存在: "+req.Title)
		}
	}

	_, ok := s.products.Update(id, func(p store.Product) store.Product {
		p.Title = req.Title
		p.Price = req.Price
		return p
	})
	return ok, nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(c *gin.Context, id int64) (bool, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return false, err
	}
	if !s.products.Del(id) {
		return false, validate.NotFoundErr("商品", id)
	}
	return true, nil
}
