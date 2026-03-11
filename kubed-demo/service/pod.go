package service

import (
	"context"
	"errors"

	"github.com/wonderivan/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// 定义pod类型和Pod对象，用于包外的调用(包是指service目录)，例如Controller
var Pod pod

type pod struct{}

// 定义列表的返回内容，Items是pod元素列表，Total为pod元素数量
type PodsResp struct {
	Items []corev1.Pod `json:"items"`
	Total int          `json:"total"`
}

// 定义PodsNs类型，返回namespace中pod的数量
type PodsNp struct {
	Namespace string `json:"namespace"`
	PodNum    int    `json:"pod_num"`
}

// 获取pod列表，支持过滤、排序、分页
func (p *pod) GetPods(client *kubernetes.Clientset, filterName, namespace string, limit, page int) (podsResp *PodsResp, err error) {
	//获取podList类型的pod列表
	//context.TODO()用于声明一个空的context上下文，用于List方法内设置这个请求的超时（源码），这里的常用用法
	//metav1.ListOptions{}用于过滤List数据，如使用label，field等
	//kubectl get services --all-namespaces --field-seletor metadata.namespace != default
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(errors.New("获取Pod列表失败, " + err.Error()))
		return nil, errors.New("获取Pod列表失败, " + err.Error())
	}
	//实例化dataSelector对象
	selectableData := &dataSelector{
		GenericDataList: p.toCells(podList.Items),
		dataSelectorQuery: &DataSelectorQuery{
			FilterQuery: &FilterQuery{Name: filterName},
			PaginateQuery: &PaginateQuery{
				Limit: limit,
				Page:  page,
			},
		},
	}
	//先过滤
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	//再排序和分页
	data := filtered.Sort().Paginate()

	//将[]DataCell类型的pod列表转为v1.pod列表
	pods := p.fromCells(data.GenericDataList)

	return &PodsResp{
		Items: pods,
		Total: total,
	}, nil
}

// toCells方法用于将pod类型数组，转换成DataCell类型数组
func (p *pod) toCells(std []corev1.Pod) []DataCell {
	cells := make([]DataCell, len(std))
	for i := range std {
		cells[i] = podCell(std[i])
	}
	return cells
}

// fromCells方法用于将DataCell类型数组，转换成pod类型数组
func (p *pod) fromCells(cells []DataCell) []corev1.Pod {
	pods := make([]corev1.Pod, len(cells))
	for i := range cells {
		pods[i] = corev1.Pod(cells[i].(podCell))
	}

	return pods
}
