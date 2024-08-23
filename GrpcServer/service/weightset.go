package service

// 用来与 Kubernetes 集群和 Istio 的 VirtualService 进行交互的。
// 主要功能包括列出 VirtualService 资源。
// 并对某个 VirtualService 的路由权重进行修改
import (
	"context"       // 用于控制多个 API 的调用上下文，例如取消请求或设置超时。
	"encoding/json" // 用于将Go结构体编码为JSON，或者从JSON解码为Go结构体。
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// Kubernetes API 中的元数据接口，例如列表选项、对象的通用字段。
	"k8s.io/apimachinery/pkg/runtime/schema"
	// Kubernetes API 中定义资源的版本和组信息。
	"k8s.io/apimachinery/pkg/types"
	// 提供 Kubernetes 中通用的类型定义，例如 Patch 类型。
	"k8s.io/client-go/dynamic"
	// Kubernetes 客户端，用于动态地与集群中的资源交互。
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// 用于加载 Kubeconfig 配置文件并生成客户端配置。
	"k8s.io/klog"
	// Kubernetes 的日志库。
)

// JSON Patch 结构体定义
type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

// 函数返回 Kubeconfig 文件的路径，用于配置 Kubernetes 客户端。
func getKubeConfig() string {
	//home := homedir.HomeDir()
	kubeconfig := "./admin.conf"
	return kubeconfig
}

// 基于指定的 Kubeconfig 文件生成 Kubernetes 客户端配置。
// 如果出错，返回 nil 和错误信息。
func clientConfig(kubeconfig string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// 设置 VirtualService 路由权重的函数
func setVirtualServiceWeights(client dynamic.Interface, service string, scheduleDesion [][]int, namespace string) error {
	//  Create a GVR which represents an Istio Virtual Service.
	// 创建一个 GVR（GroupVersionResource），指定要操作的资源类型为 Istio 的 VirtualService。
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}

	nodeNum := len(scheduleDesion)
	fmt.Println("nodeNum: ", nodeNum)
	// 遍历网关设置
	for i := 0; i < nodeNum; i++ {
		// ：创建一个 JSON Patch 操作列表，修改指定路径下的权重。
		patchPayload := make([]patchUInt32Value, nodeNum)
		// patchPayload 是一个 patchUInt32Value 类型的切片（数组），长度为 nodeNum。
		// 这个切片用于存储nodeNum个 JSON Patch 操作。
		// 每个元素表示对 VirtualService 的一个具体的路径进行修改操作。
		for j := 0; j < nodeNum; j++ {
			patchPayload[j].Op = "replace"
			// Op 字段表示 JSON Patch 操作的类型。
			// 这里，"replace" 表示替换操作。
			// 意思是，指定路径下的值将被替换为新的值。
			patchPayload[j].Path = "/spec/http/0/route/" + strconv.Itoa(j) + "/weight"
			// Path 字段表示 JSON Patch 操作所作用的对象路径
			// "/spec/http/0/route/0/weight" 表示 VirtualService 配置中的第一个路由的权重。
			// "/spec/http/0/route/1/weight" 表示 VirtualService 配置中的第二个路由的权重。
			patchPayload[j].Value = uint32(scheduleDesion[i][j])
			// Value 字段表示要设置的新值。
			// patchPayload[0].Value 将第一个路由的权重设置为 weight1。
			// patchPayload[1].Value 将第二个路由的权重设置为 weight2

		}
		patchBytes, _ := json.Marshal(patchPayload)
		// 将 patchPayload 转换为 JSON 格式。
		// patchBytes 是经过 json.Marshal 操作后得到的 JSON 字节数组。
		// patchPayload 被转换为 JSON 格式的数据
		// 以便后续通过 API 请求发送到 Kubernetes 集群，应用到 VirtualService 上。
		virtualServiceName := service + "-vs-nuc" + strconv.Itoa(i+1)
		fmt.Println("virtualServiceName: ", virtualServiceName)
		_, err := client.Resource(virtualServiceGVR).Namespace(namespace).Patch(context.Background(), virtualServiceName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
		// 对 VirtualService 应用 JSON Patch 更新权重。
		if err != nil {
			fmt.Println("Failed to patch VirtualService: ", err)
			return err
		}
	}
	return nil

}

// 输入服务名，调度决策，命名空间
func WeightSet(service string, scheduleDesion [][]int, namespace string) error {
	kubeconfig := getKubeConfig() // 获取 Kubeconfig 文件路径。

	config, err := clientConfig(kubeconfig) // 创建 Kubernetes 客户端配置。
	if err != nil {
		klog.Fatalf("Failed to create client config: %v", err)
		return err
	}

	// 为 Kubernetes 集群的 API Server 地址。
	config.Host = "https://168.11.7.126:6443" // 替换为虚拟机IP

	// 创建动态客户端。
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// 接下来就是调用dynamicClient与k8s的组件进行交互

	err = setVirtualServiceWeights(dynamicClient, service, scheduleDesion, namespace)
	if err != nil {
		fmt.Println("Failed to set VirtualService weights: ", err)
		return err
	} else {
		return nil
	}

}
