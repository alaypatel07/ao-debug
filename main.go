package main

import (
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"os"
	"strings"
)

func main() {
	podName := flag.String("pod-name", "", "The pod whose logs are to be monitored")
	podNamespace := flag.String("pod-namespace", "default", "The namespace of the pod")
	jobId := flag.String("job-id", "latest", "The filepath of the ansible stdout in the container.")

	flag.Parse()
	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		kubeconfig = "~/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalln("Failed to create config")

	}

	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %+v", err)
	}

	pod, err := restClient.CoreV1().Pods(*podNamespace).Get(*podName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get pod: %+v", err)
	}
	//
	c, err := containerToAttachTo("", pod)
	if err != nil {
		log.Fatalf("Failed to get the container: %+v", err)
	}

	command := "tail -f -n +0 /tmp/ansible-operator/runner/osb.openshift.io/v1alpha1/AutomationBroker/fail/ansible-service-broker/artifacts/" + *jobId + "/stdout"
	req := restClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(*podName).
		Namespace(*podNamespace).
		SubResource("exec")
	s := runtime.NewScheme()
	if err := v1.AddToScheme(s); err != nil {
		panic(err)
	}

	parameterCodec := runtime.NewParameterCodec(s)
	req.VersionedParams(&v1.PodExecOptions{
		Command:   strings.Fields(command),
		Container: c.Name,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	fmt.Println("Request URL:", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Fatalf("Failed to get the executor: %+v", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		log.Fatalf("Failed to run the command: %+v", err)
	}

}

// containerToAttach returns a reference to the container to attach to, given
// by name or the first container if name is empty.
func containerToAttachTo(container string, pod *v1.Pod) (*v1.Container, error) {
	if len(container) > 0 {
		for i := range pod.Spec.Containers {
			if pod.Spec.Containers[i].Name == container {
				return &pod.Spec.Containers[i], nil
			}
		}
		for i := range pod.Spec.InitContainers {
			if pod.Spec.InitContainers[i].Name == container {
				return &pod.Spec.InitContainers[i], nil
			}
		}
		return nil, fmt.Errorf("container not found (%s)", container)
	}
	return &pod.Spec.Containers[0], nil
}

