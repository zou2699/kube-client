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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// kubeconfig 配置文件
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "dev-readonly.kubeconfig"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// 指定 namespace label
	namespace := flag.String("namespace", "default", "select a namespace")
	label := flag.String("label", "", "label select")

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
		LabelSelector: *label,
	}

	err = listDeployment(clientset, listOptions)
	if err != nil {
		panic(err)
	}

	// 为AppsV1 下面的deployment创建一个watcher
	watcher, err := clientset.AppsV1().Deployments(*namespace).Watch(listOptions)
	if err != nil {
		panic(err)
	}

	// ADDED  MODIFIED  DELETED
	resultChan := watcher.ResultChan()
	fmt.Println("watching...")

	for event := range resultChan {
		// fmt.Println(event.Type)

		// 断言 针对不同的event
		deployment, ok := event.Object.(*v1.Deployment)
		// replicas := deployment.Status.Replicas
		// availableReplicas := deployment.Status.AvailableReplicas
		if !ok {
			panic(err)
		}
		switch event.Type {
		case watch.Added:
			log.Printf("deployment %s added", deployment.Name)
			err := listDeployment(clientset, listOptions)
			if err != nil {
				panic(err)
			}
		case watch.Modified:
			log.Printf("deployment %s modified", deployment.Name)
			var deploymentCondition v1.DeploymentCondition
			if len(deployment.Status.Conditions) == 0 {
				break
			}
			for _,deploymentCondition = range deployment.Status.Conditions {
				// 打印出modified message
				log.Printf("deployment modified message: %s", deploymentCondition.Message)
				// deployment是否可用
				log.Printf("deployment available status: %s", deploymentCondition.Status)
				// 输出pod的可用情况
				name := deployment.Name
				replicas := deployment.Status.Replicas
				availableReplicas := deployment.Status.AvailableReplicas
				unavailableReplicas := deployment.Status.UnavailableReplicas

				log.Printf("name: %s replicas: %d AvailableReplicas: %d unavailableReplicas: %d", name, replicas, availableReplicas, unavailableReplicas)
				if replicas != 0 && replicas == availableReplicas && unavailableReplicas == 0 {
					log.Printf("deployment %s successfully modified",name)
					err := listDeployment(clientset, listOptions)
					if err != nil {
						panic(err)
					}
				}
			}
		case watch.Deleted:
			log.Printf("deployment %s deleted", deployment.Name)
			err := listDeployment(clientset, listOptions)
			if err != nil {
				panic(err)
			}
		}

	}
}

func listDeployment(clientset *kubernetes.Clientset, listOptions metav1.ListOptions) error {
	deploymentList, err := clientset.AppsV1().Deployments(metav1.NamespaceDefault).List(listOptions)
	if err != nil {
		return err
	}

	printDeployments(deploymentList)
	return nil
}

func printDeployments(list *v1.DeploymentList) {
	const template = "%-16s%-16s%-16s\n"
	if len(list.Items) == 0 {
		fmt.Println("没有这个deployment或者pod个数为空")
		return
	}
	log.Println("当前deploymentList: ")
	log.Printf(template, "NAME", "Replicas", "AvailableReplicas")
	for _, deployment := range list.Items {

		replicas := deployment.Status.Replicas
		availableReplicas := deployment.Status.AvailableReplicas
		log.Printf(
			template,
			deployment.Name,
			strconv.FormatInt(int64(replicas), 10),
			strconv.FormatInt(int64(availableReplicas), 10))
	}
}
