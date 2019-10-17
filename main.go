/*
@Time : 2019/10/17 14:06
@Author : Tux
@File : main
@Description :
*/

/*
@Time : 2019/10/9 15:46
@Author : Tux
@File : main
@Description : k8s client
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config_tux"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// 根据kubeconfig文件生成 *restclient.Config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// 根据*restclient.Config 生成 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 创建一个deployment client
	listOptions := metav1.ListOptions{
		LabelSelector:   "",
	}

	deploymentList, err := clientset.AppsV1().Deployments(metav1.NamespaceDefault).List(listOptions)
	if err != nil {
		panic(err)
	}

	printDeployments(deploymentList)

	// 为AppsV1 下面的deployment创建一个watcher
	watcher, err := clientset.AppsV1().Deployments(metav1.NamespaceDefault).Watch(listOptions)
	if err != nil {
		panic(err)
	}

	// ADDED  MODIFIED DELETED
	resultChan := watcher.ResultChan()
	fmt.Println("watching...")

	for event := range resultChan {
		// fmt.Println(event.Type)

		// 断言 针对不同的event
		deployment,ok := event.Object.(*v1.Deployment)
		// replicas := deployment.Status.Replicas
		// availableReplicas := deployment.Status.AvailableReplicas
		if !ok {
			panic(err)
		}
		switch event.Type {
		case watch.Added:
			log.Printf("deployment %s added",deployment.Name)
		case watch.Modified:
			log.Printf("deployment %s modified",deployment.Name)
			var conditionsStatus corev1.ConditionStatus
			if len(deployment.Status.Conditions) >0 {
				conditionsStatus = deployment.Status.Conditions[0].Status
			}
			log.Printf("conditions available status %s",conditionsStatus)
			name:=deployment.Name
			replicas := deployment.Status.Replicas
			availableReplicas := deployment.Status.AvailableReplicas
			log.Printf("name: %s replicas: %d AvailableReplicas: %d",name,replicas,availableReplicas)
		case watch.Deleted:
			log.Printf("deployment %s deleted",deployment.Name)
		}
	}
}

func watchDeployment()  {

}


func printDeployments(list *v1.DeploymentList)  {
	template := "%-16s%-16s%-16s\n"
	fmt.Printf(template, "NAME", "Replicas", "AvailableReplicas")
	for _, deployment := range list.Items {

		replicas := deployment.Status.Replicas
		availableReplicas := deployment.Status.AvailableReplicas
		fmt.Printf(
			template,
			deployment.Name,
			strconv.FormatInt(int64(replicas),10),
			strconv.FormatInt(int64(availableReplicas),10))
	}
}