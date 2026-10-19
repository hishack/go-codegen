package logic

import (
	"strconv"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// OrderService 订单业务逻辑(伪实现:下单时按商品单价计算金额)
type OrderService struct {
	orders   *store.OrderStore
	products *store.ProductStore
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	return &OrderService{
		orders:   store.NewOrderStore(),
		products: store.NewProductStore(),
	}
}

// 订单状态
const (
	orderStatusUnpaid = "unpaid"
	orderStatusPaid   = "paid"
	orderStatusClosed = "closed"
)

// CreateOrder 创建订单。金额由服务端按商品单价计算,不信任前端传入
func (s *OrderService) CreateOrder(c *gin.Context, req request.CreateOrderReq) (res.CreateOrderRes, error) {
	if err := validate.RequiredID("userId", req.UserId); err != nil {
		return res.CreateOrderRes{}, err
	}
	if err := validate.RequiredID("addressId", req.AddressId); err != nil {
		return res.CreateOrderRes{}, err
	}
	if err := validate.NonEmpty("items", len(req.Items)); err != nil {
		return res.CreateOrderRes{}, err
	}
	if len(req.Items) > 50 {
		return res.CreateOrderRes{}, validate.Wrap(validate.Invalid, "items 最多 50 项")
	}
	if len([]rune(req.Remark)) > 200 {
		return res.CreateOrderRes{}, validate.Wrap(validate.Invalid, "remark 最多 200 个字符")
	}

	var amount float64
	for i, it := range req.Items {
		if it == nil {
			return res.CreateOrderRes{}, validate.Wrap(validate.Invalid, itemField(i)+" 不能为空")
		}
		if err := validate.RequiredID(itemField(i)+".productId", it.ProductId); err != nil {
			return res.CreateOrderRes{}, err
		}
		if err := validate.RangeInt(itemField(i)+".count", it.Count, 1, 999); err != nil {
			return res.CreateOrderRes{}, err
		}
		p, ok := s.products.Get(it.ProductId)
		if !ok {
			return res.CreateOrderRes{}, validate.NotFoundErr("商品", it.ProductId)
		}
		if p.Stock < it.Count {
			return res.CreateOrderRes{}, validate.Wrap(validate.Invalid, itemField(i)+" 库存不足")
		}
		amount += p.Price * float64(it.Count)
	}

	id := s.orders.NextId()
	s.orders.Put(id, store.Order{
		Id: id, UserId: req.UserId, Amount: amount,
		Status: orderStatusUnpaid, Remark: req.Remark,
		Receiver: "默认收货人", Phone: "13800000000", Addr: "默认地址",
		CreateAt: time.Now(),
	})
	return res.CreateOrderRes{OrderId: id, Amount: amount}, nil
}

// ListOrders 订单列表,支持按用户与状态过滤
func (s *OrderService) ListOrders(c *gin.Context, req request.ListOrdersReq) (res.ListOrdersRes, error) {
	if req.PageSize > 100 {
		return res.ListOrdersRes{}, validate.Wrap(validate.OutOfRange, "pageSize 最大 100")
	}
	filtered := s.orders.Filter(func(o store.Order) bool {
		if req.UserId > 0 && o.UserId != req.UserId {
			return false
		}
		if req.Status != "" && o.Status != req.Status {
			return false
		}
		return true
	})

	total := int64(len(filtered))
	page := validate.Slice(filtered, req.Page, req.PageSize)

	out := res.ListOrdersRes{Total: total, List: make([]*res.ListOrdersResListItem, 0, len(page))}
	for _, o := range page {
		out.List = append(out.List, &res.ListOrdersResListItem{
			OrderId: o.Id, Amount: o.Amount, Status: o.Status, CreateAt: o.CreateAt,
		})
	}
	return out, nil
}

// GetOrderInfo 订单详情
func (s *OrderService) GetOrderInfo(c *gin.Context, orderId int64) (res.GetOrderInfoRes, error) {
	if err := validate.RequiredID("orderId", orderId); err != nil {
		return res.GetOrderInfoRes{}, err
	}
	o, ok := s.orders.Get(orderId)
	if !ok {
		return res.GetOrderInfoRes{}, validate.NotFoundErr("订单", orderId)
	}
	return res.GetOrderInfoRes{
		OrderId: o.Id, UserId: o.UserId, Amount: o.Amount, Status: o.Status,
		Address: &res.GetOrderInfoResAddress{
			Receiver: o.Receiver, Phone: o.Phone, Detail: o.Addr,
		},
	}, nil
}

// CancelOrder 取消订单。只有 unpaid/paid 可取消
func (s *OrderService) CancelOrder(c *gin.Context, orderId int64) (bool, error) {
	if err := validate.RequiredID("orderId", orderId); err != nil {
		return false, err
	}
	o, ok := s.orders.Get(orderId)
	if !ok {
		return false, validate.NotFoundErr("订单", orderId)
	}
	if o.Status != orderStatusUnpaid && o.Status != orderStatusPaid {
		return false, validate.Wrap(validate.Invalid,
			"当前状态 "+o.Status+" 不允许取消(仅 unpaid/paid 可取消)")
	}
	_, ok = s.orders.Update(orderId, func(x store.Order) store.Order {
		x.Status = orderStatusClosed
		return x
	})
	return ok, nil
}

// itemField 生成"第 N 项"的可读字段名
func itemField(i int) string {
	return "items[" + strconv.Itoa(i) + "]"
}
