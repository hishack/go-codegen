1. 修改handler中的状态码:
func (u *ArrayHandler) CreateBatchUser(c *gin.Context) {
	var req request.CreateBatchUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](err))
		return
	}

	data, err := u.arrayService.CreateBatchUser(c, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int64](data))
}

参数出问题,应该返回400,