package k8sclient

import (
	"context"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	client *kubernetes.Clientset
}

func NewKubeClient() (*KubeClient, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)

	return &KubeClient{
		client: client,
	}, err

}

func (c *KubeClient) KillCheckoutPod() {

	pods, err := c.client.CoreV1().
		Pods("default").
		List(context.Background(), metav1.ListOptions{
			LabelSelector: "app=checkout-service",
		})

	if err != nil {
		slog.Error("Failed to list checkout-service pods: ", "error", err.Error())
		return
	}

	if len(pods.Items) == 0 {
		slog.Error("checkout-service pods is 0")
		return
	}

	pod := pods.Items[0]
	err = c.client.CoreV1().
		Pods("default").
		Delete(context.Background(), pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64), // Force delete
		})
	if err != nil {
		slog.Error("Pod delete is failed", "error", err.Error())
	}

	slog.Info("Pod deleted", "podName", pod.Name)
}
