package hvc

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// IntegrationTest skips the current Test if the TEST_INTEGRATION environment
// variable is empty.
func IntegrationTest(t *testing.T) {
	if v := os.Getenv("TEST_INTEGRATION"); len(v) == 0 {
		t.Skip()
	}
}

// PortForwardPodRequest is a structure that captures all of the port forwarding
// request details.
type PortForwardPodRequest struct {
	RestConfig *rest.Config
	Pod        corev1.Pod
	LocalPort  int
	PodPort    int
	Streams    genericclioptions.IOStreams
	StopCh     <-chan struct{}
	ReadyCh    chan struct{}
}

func NewPortForwardPodRequest(t *testing.T, localPort, podPort int, podName, podNamespace string, stopCh, readyCh chan struct{}) *PortForwardPodRequest {
	return &PortForwardPodRequest{
		RestConfig: getKubernetesConfig(t),
		LocalPort:  localPort,
		PodPort:    podPort,
		Pod: corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: podNamespace,
			},
		},
		StopCh:  stopCh,
		ReadyCh: readyCh,
		Streams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// PortForwardAPod establishes a port forwarding connection between a Pod and
// a local port.
func PortForwardAPod(req *PortForwardPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

func getKubernetesConfigFile(t *testing.T) string {
	kubeConfigPath := ""
	if v := os.Getenv("TEST_KUBECONFIG_PATH"); v != "" {
		kubeConfigPath = v
	}

	if kubeConfigPath == "" {
		usr, err := user.Current()
		require.NoError(t, err)

		kubeConfigPath = fmt.Sprintf("%s/.kube/config", usr.HomeDir)
	}

	return kubeConfigPath
}

func getKubernetesConfig(t *testing.T) *rest.Config {
	if kubernetesConfig == nil {
		var err error
		kubernetesConfig, err = clientcmd.BuildConfigFromFlags("", getKubernetesConfigFile(t))
		require.NoError(t, err)
	}

	return kubernetesConfig
}

func getKubernetesClientSet(t *testing.T) *kubernetes.Clientset {
	if kubernetesClientset == nil {
		var err error
		kubernetesClientset, err = kubernetes.NewForConfig(getKubernetesConfig(t))
		require.NoError(t, err)
	}

	return kubernetesClientset
}
