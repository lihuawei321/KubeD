package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"kubed-demo/config"

	"github.com/wonderivan/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var K8s k8s

type k8s struct {
	// 提供多集群client
	ClientMap map[string]*kubernetes.Clientset
	// 提供集群列表功能
	KubeConfMap map[string]string
}

// 根据集群名获取client
func (k *k8s) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	// 判断clientMap中是否存在该集群
	if client, ok := k.ClientMap[clusterName]; ok {
		return client, nil
	} else {
		return nil, errors.New("cluster not found")
	}
}

// 初始化client
func (k *k8s) Init() {
	mp := make(map[string]string, 0)
	// 初始化clientMap
	k.ClientMap = make(map[string]*kubernetes.Clientset, 0)
	//反序列化
	if err := json.Unmarshal([]byte(config.Kubeconfigs), &mp); err != nil {
		logger.Error("反序列化kubeconfigs失败: %s\n", err)
		return
	}
	k.KubeConfMap = mp
	for key, value := range mp {
		conf, err := clientcmd.BuildConfigFromFlags("", value)
		if err != nil {
			panic(fmt.Sprintf("集群%s: 创建K8s 配置失败 %v\n", key, err))
		}
		clientSet, err := kubernetes.NewForConfig(conf)
		if err != nil {
			panic(fmt.Sprintf("集群%s: 创建K8sClient失败 %v\n", key, err))
		}

		k.ClientMap[key] = clientSet
		logger.Info(fmt.Sprintf("集群%s: 创建K8sClient成功 ", key))
	}
}
