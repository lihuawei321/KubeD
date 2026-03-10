package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 实例化router对象，可以使用该对象点出首字母大写的方法（跨包调用
var Router router

// 定义router结构体
type router struct{}

// 初始化路由，创建测试api接口
func (*router) InitApiRouter(r *gin.Engine) {
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "testapi success",
		})
	})
}
