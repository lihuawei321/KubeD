package controller

import (
	"fmt"
	"kubed-demo/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

var Deployment deployment

type deployment struct{}

// 获取deployment列表，支持过滤、排序、分页
func (d *deployment) GetDeployments(ctx *gin.Context) {
	params := new(struct {
		FilterName string `form:"filter_name"`
		Namespace  string `form:"namespace"`
		Page       int    `form:"page"`
		Limit      int    `form:"limit"`
		Cluster    string `form:"cluster"`
	})
	if err := ctx.Bind(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		logger.Error("获取K8sClient失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	data, err := service.Deployment.GetDeployments(client, params.FilterName, params.Namespace, params.Limit, params.Page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "获取Deployment列表成功",
		"data": data,
	})
}

// 获取deployment详情
func (d *deployment) GetDeploymentDetail(ctx *gin.Context) {
	params := new(struct {
		Namespace      string `form:"namespace"`
		DeploymentName string `form:"deployment_name"`
		Cluster        string `form:"cluster"`
	})
	if err := ctx.Bind(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		logger.Error("获取K8sClient失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	data, err := service.Deployment.GetDeploymentDetail(client, params.DeploymentName, params.Namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "获取Deployment详情成功",
		"data": data,
	})
}

// 删除Deployment
func (d *deployment) DeleteDeployment(ctx *gin.Context) {
	params := new(struct {
		Namespace      string `form:"namespace"`
		DeploymentName string `form:"deployment_name"`
		Cluster        string `form:"cluster"`
	})
	if err := ctx.ShouldBindJSON(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		logger.Error("获取K8sClient失败, " + err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	err = service.Deployment.DeleteDeployment(client, params.Namespace, params.DeploymentName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "删除Deployment成功",
		"data": nil,
	})
}

// 更新Deployment
func (d *deployment) UpdateDeployment(ctx *gin.Context) {
	params := new(struct {
		Namespace string `json:"namespace"`
		Cluster   string `json:"cluster"`
		Content   string `json:"content"`
	})
	if err := ctx.ShouldBindJSON(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		logger.Error("获取K8sClient失败, " + err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	err = service.Deployment.UpdateDeployment(client, params.Namespace, params.Content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "更新Deployment成功",
		"data": nil,
	})
}

// 重启deployment
func (d *deployment) RestartDeployment(ctx *gin.Context) {
	params := new(struct {
		DeploymentName string `json:"deployment_name"`
		Namespace      string `json:"namespace"`
		Cluster        string `json:"cluster"`
	})
	//PUT请求，绑定参数方法改为ctx.ShouldBindJSON
	if err := ctx.ShouldBindJSON(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	err = service.Deployment.RestartDeployment(client, params.DeploymentName, params.Namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "重启Deployment成功",
		"data": nil,
	})
}

// 获取每个namespace的pod数量
func (d *deployment) GetDeployNumPerNp(ctx *gin.Context) {
	params := new(struct {
		Cluster string `form:"cluster"`
	})
	if err := ctx.Bind(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	data, err := service.Deployment.GetDeployNumPerNp(client)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "获取每个namespace的deployment数量成功",
		"data": data,
	})
}

// 创建deployment
func (d *deployment) CreateDeployment(ctx *gin.Context) {
	var (
		deployCreate = new(service.DeployCreate)
		err          error
	)

	if err = ctx.ShouldBindJSON(deployCreate); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(deployCreate.Cluster)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	if err = service.Deployment.CreateDeployment(client, deployCreate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "创建Deployment成功",
		"data": nil,
	})
}

// 设置deployment副本数
func (d *deployment) ScaleDeployment(ctx *gin.Context) {
	params := new(struct {
		DeploymentName string `json:"deployment_name"`
		Namespace      string `json:"namespace"`
		ScaleNum       int    `json:"scale_num"`
		Cluster        string `json:"cluster"`
	})
	//PUT请求，绑定参数方法改为ctx.ShouldBindJSON
	if err := ctx.ShouldBindJSON(params); err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	client, err := service.K8s.GetClient(params.Cluster)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	data, err := service.Deployment.ScaleDeployment(client, params.DeploymentName, params.Namespace, params.ScaleNum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "设置Deployment副本数成功",
		"data": fmt.Sprintf("最新副本数: %d", data),
	})
}
