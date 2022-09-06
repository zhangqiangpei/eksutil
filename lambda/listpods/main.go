package main

import (
	"context"
	"os"

	eksauth "github.com/chankh/eksutil/pkg/auth"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
)

func main() {
	if os.Getenv("ENV") == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	}

	lambda.Start(handler)
}

func handler(context context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Setup the basic EKS cluster info
	cfg := &eksauth.ClusterConfig{
		ClusterName: os.Getenv("CLUSTER_NAME"),
	}

	clientset, err := eksauth.NewAuthClient(cfg)
	if err != nil {
		log.WithError(err).Fatal(err.Error())
	}

	// group_name

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Error listing nodes")
	}
	for i := range nodes.Items {
		for j := range nodes.Items[i].Spec.Taints {
			taint := nodes.Items[i].Spec.Taints[j]
			if taint.Key == "dedicated" && taint.Value == "custom_ami-20220715023441369500000002" && taint.Effect != corev1.TaintEffectNoExecute {
				nodes.Items[i].Spec.Taints[j].Effect = corev1.TaintEffectNoExecute
				node := &nodes.Items[i]
				_, err := clientset.CoreV1().Nodes().Update(node)
				if err != nil {
					log.WithError(err).Fatal("Error Update node")
				}else {
					log.Infof("[node: %s] Taints: %v \n", nodes.Items[i].Name, nodes.Items[i].Spec.Taints[j])
				}
			}
		}
	}


	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}
