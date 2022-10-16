package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	//rbacv1 "k8s.io/api/rbac/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/watch"
	"k8s.io/client-go/util/homedir"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	// Create kubeclient
	var kubeconfig string
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Define a watch function that ListWatcher will use later.
	// Here we only look pod in "default" namespaces.
	watchFunc := func(options metav1.ListOptions) (apiWatch.Interface, error) {
		// Return watcher with 10m timeout
		// timeOut := int64(10)
		// return clientset.CoreV1().Pods("default").Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeOut})
		// Return watcher on lebelSelector app=swag
		return clientset.CoreV1().Pods("default").Watch(context.Background(), metav1.ListOptions{LabelSelector: "app=swag"})
	}

	// Start watching
	fmt.Printf("----Start watching----\n")
	w, err := watch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		panic(err)
	}

	for {
		event, ok := <-w.ResultChan()
		if !ok {
			panic(fmt.Errorf("ERROR, Channel is closed"))
		}
		m, ok := event.Object.(*corev1.Pod)
		if !ok {
			panic(fmt.Errorf("Type mismatch"))
		}
		creationTime := m.GetCreationTimestamp()
		if event.Type == apiWatch.Added && creationTime.Time.Before(time.Now().Add(-20*time.Minute)) {
			fmt.Printf("Skip older events. CreationTime: %s CurrentTime: %s\n", creationTime.Time.String(), time.Now().Add(20*time.Minute).String())
			continue
		}

		// Handle event
		eventHandler(event)

		// fmt.Printf("----INCOMING EVENT\n%#v %#v\n----\n", event.Type, event.Object)
		// time.Sleep(20 * time.Millisecond)
	}
}

func eventHandler(event apiWatch.Event) {
	// Serialize struct to json
	b, err := json.Marshal(event.Object)
	if err != nil {
		fmt.Println("error on json", err)
		return
	}

	// Post to endpoint.
	notifyUrl, _ := url.Parse("https://398e-59-124-114-73.jp.ngrok.io")

	querys := notifyUrl.Query()
	querys.Add("type", string(event.Type))

	notifyUrl.RawQuery = querys.Encode()

	req, err := http.NewRequest("POST", notifyUrl.String(), bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))

}
