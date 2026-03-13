package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/wonderivan/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

var Deployment deployment

type deployment struct{}

// 定义列表的返回内容，Items是deployment元素列表，Total为deployment元素数量
type DeploymentsResp struct {
	Items []appsv1.Deployment `json:"items"`
	Total int                 `json:"total"`
}

type DeployCreate struct {
	Namespace     string            `json:"namespace"`
	Name          string            `json:"name"`
	Replicas      int32             `json:"replicas"`
	Image         string            `json:"image"`
	Label         map[string]string `json:"label"`
	ContainerPort int32             `json:"container_port"`
	Cpu           string            `json:"cpu"`
	Memory        string            `json:"memory"`
	HealthCheck   bool              `json:"health_check"`
	HealthPath    string            `json:"health_path"`
	Cluster       string            `json:"cluster"`
}

// 定义每个namespace的deployment数量
type DeploysNp struct {
	Namespace string `json:"namespace"`
	DeployNum int    `json:"deployment_num"`
}

// 获取deployment列表，支持过滤、排序、分页
func (d *deployment) GetDeployments(client *kubernetes.Clientset, filterName, namespace string, limit, page int) (deploymentsResp *DeploymentsResp, err error) {
	//获取deployment列表
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(errors.New("获取Deployment列表失败, " + err.Error()))
		return nil, errors.New("获取Deployment列表失败, " + err.Error())
	}
	//将deploymentList.Items转换成DataCell类型
	selectableData := &dataSelector{
		GenericDataList: d.toCells(deploymentList.Items),
		dataSelectorQuery: &DataSelectorQuery{
			FilterQuery:   &FilterQuery{Name: filterName},
			PaginateQuery: &PaginateQuery{Limit: limit, Page: page},
		},
	}
	//过滤出符合条件的deployment列表
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	//排序和分页
	data := filtered.Sort().Paginate()
	//将DataCell类型转换成Deployment类型
	deployments := d.fromCells(data.GenericDataList)
	return &DeploymentsResp{
		Items: deployments,
		Total: total,
	}, nil
}

// 获取deployment详情
func (d *deployment) GetDeploymentDetail(client *kubernetes.Clientset, deploymentName, namespace string) (deployment *appsv1.Deployment, err error) {
	deployment, err = client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		logger.Error(errors.New("获取Deployment详情失败, " + err.Error()))
		return nil, errors.New("获取Deployment详情失败, " + err.Error())
	}

	return deployment, nil
}

// 删除Deployment
func (d *deployment) DeleteDeployment(client *kubernetes.Clientset, namespace, deploymentName string) (err error) {
	err = client.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		logger.Error(errors.New("删除Deployment失败, " + err.Error()))
		return errors.New("删除Deployment失败, " + err.Error())
	}
	return nil
}

// 更新Deployment
func (d *deployment) UpdateDeployment(client *kubernetes.Clientset, namespace, content string) (err error) {
	var deployment = &appsv1.Deployment{}
	err = json.Unmarshal([]byte(content), deployment)
	if err != nil {
		logger.Error(errors.New("反序列化Deployment失败, " + err.Error()))
		return errors.New("反序列化Deployment失败, " + err.Error())
	}
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(errors.New("更新Deployment失败, " + err.Error()))
		return errors.New("更新Deployment失败, " + err.Error())
	}
	return nil
}

// 重启deployment
func (d *deployment) RestartDeployment(client *kubernetes.Clientset, deploymentName, namespace string) (err error) {
	//此功能等同于一下kubectl命令
	//kubectl deployment ${service} -p \
	//'{"spec":{"template":{"spec":{"containers":[{"name":"'"${service}"'","env":[{"name":"RESTART_","value":"'$(date +%s)'"}]}]}}}}'

	//使用patchData Map组装数据
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{"name": deploymentName,
							"env": []map[string]string{{
								"name":  "RESTART_",
								"value": strconv.FormatInt(time.Now().Unix(), 10),
							}},
						},
					},
				},
			},
		},
	}
	//序列化为字节，因为patch方法只接收字节类型参数
	patchByte, err := json.Marshal(patchData)
	if err != nil {
		logger.Error(errors.New("json序列化失败, " + err.Error()))
		return errors.New("json序列化失败, " + err.Error())
	}
	//调用patch方法更新deployment
	_, err = client.AppsV1().Deployments(namespace).Patch(context.TODO(), deploymentName, "application/strategic-merge-patch+json", patchByte, metav1.PatchOptions{})
	if err != nil {
		logger.Error(errors.New("重启Deployment失败, " + err.Error()))
		return errors.New("重启Deployment失败, " + err.Error())
	}

	return nil
}

// 获取每个namespace的deployment数量
func (d *deployment) GetDeployNumPerNp(client *kubernetes.Clientset) (deploysNps []*DeploysNp, err error) {
	namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaceList.Items {
		deploymentList, err := client.AppsV1().Deployments(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		deploysNp := &DeploysNp{
			Namespace: namespace.Name,
			DeployNum: len(deploymentList.Items),
		}

		deploysNps = append(deploysNps, deploysNp)
	}
	return deploysNps, nil
}

// 设置deployment副本数
func (d *deployment) ScaleDeployment(client *kubernetes.Clientset, deploymentName, namespace string, scaleNum int) (replica int32, err error) {
	//获取autoscalingv1.Scale类型的对象，能点出当前的副本数
	scale, err := client.AppsV1().Deployments(namespace).GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		logger.Error(errors.New("获取Deployment副本数信息失败, " + err.Error()))
		return 0, errors.New("获取Deployment副本数信息失败, " + err.Error())
	}
	//修改副本数
	scale.Spec.Replicas = int32(scaleNum)
	//更新副本数，传入scale对象
	newScale, err := client.AppsV1().Deployments(namespace).UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(errors.New("更新Deployment副本数信息失败, " + err.Error()))
		return 0, errors.New("更新Deployment副本数信息失败, " + err.Error())
	}

	return newScale.Spec.Replicas, nil
}

// 创建deployment,接收DeployCreate对象
func (d *deployment) CreateDeployment(client *kubernetes.Clientset, data *DeployCreate) (err error) {
	//将data中的属性组装成appsv1.Deployment对象
	deployment := &appsv1.Deployment{
		//ObjectMeta中定义资源名、命名空间以及标签
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: data.Namespace,
			Labels:    data.Label,
		},
		//Spec中定义副本数、选择器、以及pod属性
		Spec: appsv1.DeploymentSpec{
			Replicas: &data.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: data.Label,
			},
			Template: corev1.PodTemplateSpec{
				//定义pod名和标签
				ObjectMeta: metav1.ObjectMeta{
					Name:   data.Name,
					Labels: data.Label,
				},
				//定义容器名、镜像和端口
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  data.Name,
							Image: data.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
		//Status定义资源的运行状态，这里由于是新建，传入空的appsv1.DeploymentStatus{}对象即可
		Status: appsv1.DeploymentStatus{},
	}
	//判断是否打开健康检查功能，若打开，则定义ReadinessProbe和LivenessProbe
	if data.HealthCheck {
		//设置第一个容器的ReadinessProbe，因为我们pod中只有一个容器，所以直接使用index 0即可
		//若pod中有多个容器，则这里需要使用for循环去定义了
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: data.HealthPath,
					//intstr.IntOrString的作用是端口可以定义为整型，也可以定义为字符串
					//Type=0则表示表示该结构体实例内的数据为整型，转json时只使用IntVal的数据
					//Type=1则表示表示该结构体实例内的数据为字符串，转json时只使用StrVal的数据
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: data.ContainerPort,
					},
				},
			},
			//初始化等待时间
			InitialDelaySeconds: 5,
			//超时时间
			TimeoutSeconds: 5,
			//执行间隔
			PeriodSeconds: 5,
		}
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: data.HealthPath,
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: data.ContainerPort,
					},
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       5,
		}
		//定义容器的limit和request资源
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse(data.Cpu),
			corev1.ResourceMemory: resource.MustParse(data.Memory),
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse(data.Cpu),
			corev1.ResourceMemory: resource.MustParse(data.Memory),
		}
	}
	//调用sdk创建deployment
	_, err = client.AppsV1().Deployments(data.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		logger.Error(errors.New("创建Deployment失败, " + err.Error()))
		return errors.New("创建Deployment失败, " + err.Error())
	}

	return nil
}

func (d *deployment) toCells(std []appsv1.Deployment) []DataCell {
	cells := make([]DataCell, len(std))
	for i := range std {
		cells[i] = deploymentCell(std[i])
	}
	return cells
}

// fromCells方法用于将DataCell类型数组，转换成deployment类型数组
func (d *deployment) fromCells(cells []DataCell) []appsv1.Deployment {
	deployments := make([]appsv1.Deployment, len(cells))
	for i := range cells {
		deployments[i] = appsv1.Deployment(cells[i].(deploymentCell))
	}
	return deployments
}
