package main

import (
	"context"
	"kubed-demo/config"
	"kubed-demo/controller"
	"net/http"
	"os"
	"os/signal"
	"time"

	"kubed-demo/service"

	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

func main() {
	// 初始化gin引擎
	r := gin.Default()
	// 初始化k8s client
	service.K8s.Init()
	// 初始化api路由
	controller.Router.InitApiRouter(r)
	// 启动gin server
	srv := &http.Server{
		Addr:    config.ListenAddr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen: %s\n", err)
		}
	}()
	// 优雅关闭server
	// 声明一个系统信号的channel，并监听，如果没有信号，则阻塞，如果有，就继续执行
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	//设置ctx超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//cancel用于释放ctx
	defer cancel()

	// 优雅关闭server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Gin server 关闭异常: %s\n", err)
	}
	logger.Info("Gin server 退出成功")

}
