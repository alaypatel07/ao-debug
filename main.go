package main

import (
	"flag"
	"fmt"
	"io"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"os"
	"strings"
	"sync"
)

func getDeployment(r io.Reader) *v12.Deployment {
	d := yaml.NewYAMLOrJSONDecoder(r, 100000)
	var dep v12.Deployment
	err := d.Decode(&dep)
	if err != nil {
		log.Fatalf("Failed to decode deployment file: %#v\n", err)
	}
	return &dep
}

func getCustomResource(r io.Reader) *unstructured.Unstructured {
	d := yaml.NewYAMLOrJSONDecoder(r, 100000)
	var u unstructured.Unstructured
	err := d.Decode(&u)
	if err != nil {
		log.Fatalf("Failed to decode the cr: %#v\n", err)
	}
	return &u
}

func getReader(filepath string) *os.File {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Could not open file:  %#v\n", err)
		return nil
	}
	return f

}

func main() {
	operatorYAML := flag.String("deployment-filepath", "./deploy/operator.yaml", "filepath of ansible-operator deployment file. Defaults to `./deploy/operator.yaml`")
	operatorNamespace := flag.String("namespace", "default", "The namespace in which operator is running. Defaults to `default`")
	crYAML := flag.String("cr-filepath", "./deploy/cr.yaml", "filepath of the cr yaml. Defaults to ./deploy/cr.yaml")
	jobID := flag.String("job-id", "latest", "The job id for which logs needs to be displayed")

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "filepath to the kubeconfig")

	flag.Parse()

	deployment := getDeployment(getReader(*operatorYAML))
	cr := getCustomResource(getReader(*crYAML))

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create config: %#v\n", err)
	}

	restClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %+v", err)
	}

	labelSelector := deployment.Spec.Selector

	watchInterface, err := restClient.CoreV1().Pods(*operatorNamespace).Watch(metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})

	if err != nil {
		log.Fatalf("Failed to watch the pods: %+v\n", err)
	}
	waitChan := make(chan bool)
	var pod *v1.Pod
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for {
			select {
			case e := <-watchInterface.ResultChan():
				if e.Type == watch.Added {
					pod, _ = e.Object.(*v1.Pod)
					fmt.Printf("Pod %+v added\n", pod.Name)
					waitChan <- true
				}

			}
		}
		wg.Done()
	}()
	<-waitChan
	fmt.Println("Signal received")
	func() {
		c, err := containerToAttachTo("", pod)
		if err != nil {
			log.Fatalf("Failed to get the container: %+v", err)
		}

		command := "cat /tmp/ansible-operator/runner/" + cr.GroupVersionKind().GroupVersion().String() + "/" + cr.GroupVersionKind().Kind + "/" + *operatorNamespace + "/" + cr.GetName() + "/artifacts/" + *jobID + "/stdout"
		req := restClient.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
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
	}()
	return
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
