package logic

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"codegen-example/internal/model/gen/request"
	"codegen-example/pkg/errorutil"

	"github.com/gin-gonic/gin"
)

// newCtx 造一个 gin 测试上下文
func newCtx() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)
	return c
}

// validCreateReq 一份合法的创建请求
func validCreateReq() request.CreateUserReq {
	return request.CreateUserReq{
		Name: "王五", Age: 30,
		Height: 172, Weight: 65.5, Enabled: true,
		Score: 300, Level: 6, Count: 9, Ratio: 0.75,
		Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		Duration: 45 * time.Minute,
		Nickname: "风一样的男子",
	}
}

// TestCreateUser 正常创建
func TestCreateUser(t *testing.T) {
	s := NewUserService()
	got, err := s.CreateUser(newCtx(), validCreateReq())
	if err != nil {
		t.Fatalf("期望成功,实际报错: %v", err)
	}
	if got.Id <= 0 {
		t.Errorf("期望返回正数 id,实际 %d", got.Id)
	}
	if got.CreatedAt.IsZero() {
		t.Error("期望返回创建时间")
	}
}

// TestCreateUserInvalidParams 校验失败应返回业务错误且不落库
func TestCreateUserInvalidParams(t *testing.T) {
	cases := []struct {
		name string
		req  request.CreateUserReq
	}{
		{"姓名为空", request.CreateUserReq{Age: 20}},
		{"年龄为负", request.CreateUserReq{Name: "甲", Age: -1}},
		{"年龄过大", request.CreateUserReq{Name: "甲", Age: 200}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewUserService()
			if _, err := s.CreateUser(newCtx(), tc.req); err == nil {
				t.Fatal("期望返回参数错误")
			}
			list, _ := s.ListUsers(newCtx())
			if list.Total != 2 { // 只有两条种子数据,说明没写入脏数据
				t.Errorf("参数错误不应写入数据,实际条数 %d", list.Total)
			}
		})
	}
}

// TestGetUser 存在与不存在两种情况
func TestGetUser(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	got, err := s.GetUser(c, 1)
	if err != nil {
		t.Fatalf("查询种子用户应成功: %v", err)
	}
	if got.Name != "张三" {
		t.Errorf("期望返回 张三,实际 %q", got.Name)
	}

	if _, err := s.GetUser(c, 999); err == nil {
		t.Error("查询不存在的用户应报错")
	}
	if _, err := s.GetUser(c, 0); err == nil {
		t.Error("id 非法应报错")
	}
}

// TestListUsers 种子数据应可列出
func TestListUsers(t *testing.T) {
	s := NewUserService()
	got, err := s.ListUsers(newCtx())
	if err != nil {
		t.Fatalf("ListUsers 失败: %v", err)
	}
	if got.Total != 2 {
		t.Errorf("期望 2 条种子数据,实际 %d", got.Total)
	}
	if len(got.List) != int(got.Total) {
		t.Errorf("列表长度与总数不一致: %d vs %d", len(got.List), got.Total)
	}
	// id 应升序
	if got.List[0].Id > got.List[1].Id {
		t.Error("列表未按 id 升序")
	}
}

// TestCreateThenGet 串起创建与查询,验证数据确实落库
func TestCreateThenGet(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	created, err := s.CreateUser(c, validCreateReq())
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetUser(c, created.Id)
	if err != nil {
		t.Fatalf("创建后应能查到: %v", err)
	}
	if got.Name != "王五" || got.Age != 30 {
		t.Errorf("查询结果与创建不一致: %+v", got)
	}
}

// TestUpdateUser 更新存在与不存在的用户
func TestUpdateUser(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	ok, err := s.UpdateUser(c, 1, request.UpdateUserReq{Name: "张三改", Age: 31})
	if err != nil || !ok {
		t.Fatalf("更新应成功: ok=%v err=%v", ok, err)
	}
	got, _ := s.GetUser(c, 1)
	if got.Name != "张三改" || got.Age != 31 {
		t.Errorf("更新未生效: %+v", got)
	}

	if _, err := s.UpdateUser(c, 999, request.UpdateUserReq{Name: "x", Age: 1}); err == nil {
		t.Error("更新不存在的用户应报错")
	}
	if _, err := s.UpdateUser(c, 1, request.UpdateUserReq{Name: "", Age: 10}); err == nil {
		t.Error("姓名为空应报参数错误")
	}
}

// TestPatchUser 局部更新:空昵称视为不修改
func TestPatchUser(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	ok, err := s.PatchUser(c, 1, request.PatchUserReq{Nickname: "新昵称"})
	if err != nil || !ok {
		t.Fatalf("局部更新应成功: ok=%v err=%v", ok, err)
	}
	profile, err := s.GetUserProfileVo(c, 1)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Profile.Nickname != "新昵称" {
		t.Errorf("昵称未更新: %q", profile.Profile.Nickname)
	}

	// 空昵称:成功但不改动
	before, _ := s.GetUserProfileVo(c, 1)
	if ok, err := s.PatchUser(c, 1, request.PatchUserReq{Nickname: "  "}); err != nil || !ok {
		t.Errorf("空昵称应视为成功且不改动: ok=%v err=%v", ok, err)
	}
	after, _ := s.GetUserProfileVo(c, 1)
	if before.Profile.Nickname != after.Profile.Nickname {
		t.Error("空昵称不应改动数据")
	}

	if _, err := s.PatchUser(c, 999, request.PatchUserReq{Nickname: "x"}); err == nil {
		t.Error("更新不存在的用户应报错")
	}
}

// TestDeleteUser 删除后不可再查
func TestDeleteUser(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	ok, err := s.DeleteUser(c, 1)
	if err != nil || !ok {
		t.Fatalf("删除应成功: ok=%v err=%v", ok, err)
	}
	if _, err := s.GetUser(c, 1); err == nil {
		t.Error("删除后不应再查到")
	}
	if _, err := s.DeleteUser(c, 1); err == nil {
		t.Error("重复删除应报错")
	}
}

// TestGetUserProfileVo 详情含标签
func TestGetUserProfileVo(t *testing.T) {
	s := NewUserService()
	c := newCtx()

	got, err := s.GetUserProfileVo(c, 1)
	if err != nil {
		t.Fatalf("查询详情失败: %v", err)
	}
	if got.Id != 1 {
		t.Errorf("期望 id=1,实际 %d", got.Id)
	}
	if got.Profile == nil {
		t.Fatal("Profile 不应为 nil")
	}
	if len(got.Profile.Tags) != 2 {
		t.Errorf("期望 2 个标签,实际 %v", got.Profile.Tags)
	}

	if _, err := s.GetUserProfileVo(c, 999); err == nil {
		t.Error("查询不存在的用户应报错")
	}
	if _, err := s.GetUserProfileVo(c, -1); err == nil {
		t.Error("id 非法应报错")
	}
}

// makeUploadReq 造一个 multipart 请求
func makeUploadReq(userId int64, fileName, content string) request.UploadAvatarReq {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	// userId 是表单字段,必须一起提交,否则绑定出来是 0
	_ = w.WriteField("userId", strconv.FormatInt(userId, 10))
	part, _ := w.CreateFormFile("avatar", fileName)
	part.Write([]byte(content))
	w.Close()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("POST", "/users/avatar", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	c.Request = req

	var r request.UploadAvatarReq
	if err := c.ShouldBind(&r); err != nil {
		return request.UploadAvatarReq{}
	}
	return r
}

// TestUploadAvatar 正常上传与各种异常分支
func TestUploadAvatar(t *testing.T) {
	t.Run("正常上传", func(t *testing.T) {
		s := NewUserService()
		url, err := s.UploadAvatar(newCtx(), makeUploadReq(1, "a.png", "PNGDATA"))
		if err != nil {
			t.Fatalf("上传应成功: %v", err)
		}
		if url != "/static/avatar/1.png" {
			t.Errorf("返回地址不对: %q", url)
		}
	})

	t.Run("用户不存在", func(t *testing.T) {
		s := NewUserService()
		if _, err := s.UploadAvatar(newCtx(), makeUploadReq(999, "a.png", "x")); err == nil {
			t.Error("应为不存在用户报错")
		}
	})

	t.Run("扩展名不支持", func(t *testing.T) {
		s := NewUserService()
		if _, err := s.UploadAvatar(newCtx(), makeUploadReq(1, "a.exe", "x")); err == nil {
			t.Error("应拒绝非图片扩展名")
		}
	})

	t.Run("缺少文件", func(t *testing.T) {
		s := NewUserService()
		if _, err := s.UploadAvatar(newCtx(), request.UploadAvatarReq{UserId: 1}); err == nil {
			t.Error("缺少头像文件应报错")
		}
	})

	t.Run("文件过大", func(t *testing.T) {
		s := NewUserService()
		big := bytes.Repeat([]byte("A"), maxAvatarSize+1)
		if _, err := s.UploadAvatar(newCtx(), makeUploadReq(1, "a.png", string(big))); err == nil {
			t.Error("超大文件应报错")
		}
	})
}

// TestBusinessErrorCode 错误码应能被上层识别,且描述不应被上下文冲掉
func TestBusinessErrorCode(t *testing.T) {
	s := NewUserService()
	_, err := s.GetUser(newCtx(), 999)
	if err == nil {
		t.Fatal("期望报错")
	}
	be := errorutil.ConvertError(err)
	if be.Code != errUserNotFound.Code {
		t.Errorf("错误码不对: 期望 %d,实际 %d", errUserNotFound.Code, be.Code)
	}
	// wrap 应保留原始描述并追加上下文
	if be.Message != "用户不存在: id=999" {
		t.Errorf("错误描述不对: %q", be.Message)
	}
}
