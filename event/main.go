/*
@Time : 2019/10/24 10:51
@Author : Tux
@File : main
@Description :
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "tux-readonly.kubeconfig"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// 指定 namespace label
	namespace := flag.String("namespace", "default", "select a namespace")
	// label := flag.String("label", "", "label select")

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
		// LabelSelector: *label,
	}

	// err = listDeployment(clientset, listOptions)
	// if err != nil {
	// 	panic(err)
	// }

	// 创建一个 event 的 watcher
	watcher, err := clientset.CoreV1().Events(*namespace).Watch(listOptions)
	if err != nil {
		log.Panic(err)
	}

	// ADDED  MODIFIED  DELETED
	resultChan := watcher.ResultChan()
	log.Println("watching...")

	fmt.Printf("%T\n", resultChan)
	for result := range resultChan {
		// fmt.Println(result.Type)
		// fmt.Printf("类型 %T\n", result)

		// fmt.Println(result.Object)
		// 断言 针对不同的event
		event, ok := result.Object.(*corev1.Event)
		// replicas := deployment.Status.Replicas
		// availableReplicas := deployment.Status.AvailableReplicas
		if !ok {
			log.Panicf("type assertion err,%T", result.Object)
		}
		switch result.Type {
		case watch.Added:
			// log.Printf("cluster event %s added", event.Name)
			// log.Printf("reason %s message %s",event.Reason,event.Message)
			logrus.WithFields(logrus.Fields{
				"name":    event.Name,
				"reason":  event.Reason,
				"message": event.Message,
			}).Print("added")
		case watch.Modified:
			logrus.WithFields(logrus.Fields{
				"name":    event.Name,
				"reason":  event.Reason,
				"message": event.Message,
			}).Print("Modified")

		case watch.Deleted:
			logrus.WithFields(logrus.Fields{
				"name":    event.Name,
				"reason":  event.Reason,
				"message": event.Message,
			}).Print("Deleted")

		case watch.Error:
			logrus.WithFields(logrus.Fields{
				"name":    event.Name,
				"reason":  event.Reason,
				"message": event.Message,
			}).Print("Error")

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
