package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kubed-demo/config"

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
	Total int          `json:"total"` //Total为pod元素数量,用于前端分页
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
	//实例化dataSelector对象,用于封装排序，过滤，分页的数据类型
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
	//先过滤出符合条件的pod列表
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	//再排序和分页，排序和分页后返回数据
	data := filtered.Sort().Paginate()

	//将[]DataCell类型的pod列表转为v1.pod列表，用于返回给前端
	pods := p.fromCells(data.GenericDataList)

	return &PodsResp{
		Items: pods,
		Total: total,
	}, nil
}

// 获取pod详情
func (p *pod) GetPodDetail(client *kubernetes.Clientset, namespace, PodName string) (podDetail *corev1.Pod, err error) {
	//获取pod详情
	//context.TODO()用于声明一个空的context上下文，用于Get方法内设置这个请求的超时（源码），这里的常用用法
	//metav1.GetOptions{}用于获取Pod详情，如使用label，field等
	podDetail, err = client.CoreV1().Pods(namespace).Get(context.TODO(), PodName, metav1.GetOptions{})
	if err != nil {
		logger.Error(errors.New("获取Pod详情失败, " + err.Error()))
		return nil, errors.New("获取Pod详情失败, " + err.Error())
	}
	return podDetail, nil
}

// 删除pod
func (p *pod) DeletePod(client *kubernetes.Clientset, namespace, podName string) (err error) {
	//删除pod
	//context.TODO()用于声明一个空的context上下文，用于Delete方法内设置这个请求的超时（源码），这里的常用用法
	//metav1.DeleteOptions{}用于删除Pod，如使用label，field等
	fmt.Println("--------namespace, podName--------", namespace, podName)
	err = client.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		logger.Error(errors.New("删除Pod失败, " + err.Error()))
		return errors.New("删除Pod失败, " + err.Error())
	}
	return nil
}

// 更新pod
func (p *pod) UpdatePod(client *kubernetes.Clientset, namespace, PodName, content string) (err error) {
	var pod = &corev1.Pod{}
	// 反序列化为pod对象
	err = json.Unmarshal([]byte(content), pod)
	if err != nil {
		logger.Error(errors.New("反序列化Pod失败, " + err.Error()))
		return errors.New("反序列化Pod失败, " + err.Error())
	}
	//更新pod
	_, err = client.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(errors.New("更新Pod失败, " + err.Error()))
		return errors.New("更新Pod失败, " + err.Error())
	}
	return nil
}

// 获取pod中容器名
func (p *pod) GetPodContainer(client *kubernetes.Clientset, podName, namespace string) (containers []string, err error) {
	//获取pod详情
	pod, err := p.GetPodDetail(client, podName, namespace)
	if err != nil {
		return nil, err
	}
	//从pod对象中拿到容器名
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return containers, nil
}

// 获取pod内容器日志
func (p *pod) GetPodLog(client *kubernetes.Clientset, containerName, podName, namespace string) (log string, err error) {
	//设置日志的配置，容器名、tail的行数
	lineLimit := int64(config.PodLogTailLine)
	option := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &lineLimit,
	}
	//获取request实例
	req := client.CoreV1().Pods(namespace).GetLogs(podName, option)
	//发起request请求，返回一个io.ReadCloser类型（等同于response.body）
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		logger.Error(errors.New("获取PodLog失败, " + err.Error()))
		return "", errors.New("获取PodLog失败, " + err.Error())
	}
	defer podLogs.Close()
	//将response body写入到缓冲区，目的是为了转成string返回
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		logger.Error(errors.New("复制PodLog失败, " + err.Error()))
		return "", errors.New("复制PodLog失败, " + err.Error())
	}

	return buf.String(), nil
}

// 获取每个namespace的pod数量
func (p *pod) GetPodNumPerNp(client *kubernetes.Clientset) (podsNps []*PodsNp, err error) {
	//获取namespace列表
	namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaceList.Items {
		//获取pod列表
		podList, err := client.CoreV1().Pods(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		//组装数据
		podsNp := &PodsNp{
			Namespace: namespace.Name,
			PodNum:    len(podList.Items),
		}
		//添加到podsNps数组中
		podsNps = append(podsNps, podsNp)
	}
	return podsNps, nil
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
