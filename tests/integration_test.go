package hvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var kubernetesConfig *rest.Config
var kubernetesClientset *kubernetes.Clientset

const targetVaultPort int = 8300
const sourceVaultPort int = 8400

func TestIntegrationSuite(t *testing.T) {
	IntegrationTest(t)

	setupOptions := setupComponents(t)
	defer terraform.Destroy(t, setupOptions)

	stopCh := setupPortForwards(t)
	defer close(stopCh)

	configureOptions := configureComponents(t)
	defer terraform.Destroy(t, configureOptions)

	t.Run("Single Secret", verifySingleSecret)
	t.Run("Single Secret from multiple Sources", verifySingleSecretMultipleSources)
	t.Run("Multiple Secrets", verifyMultipleSecrets)
	t.Run("Kubernetes Authentication", verifyKubernetesAuthentication)
	t.Run("Access Denied Error", verifyAccessDenied)
}

func setupComponents(t *testing.T) *terraform.Options {
	options := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./setup",
		Vars:         map[string]interface{}{},
	})

	_, err := terraform.InitAndApplyAndIdempotentE(t, options)
	require.NoError(t, err)

	return options
}

func setupPortForwards(t *testing.T) chan struct{} {
	stopCh := make(chan struct{}, 1)

	sourceReq := NewPortForwardPodRequest(t, sourceVaultPort, 8200, "source-vault", "default", stopCh, make(chan struct{}))
	targetReq := NewPortForwardPodRequest(t, targetVaultPort, 8200, "target-vault", "default", stopCh, make(chan struct{}))

	go func() {
		require.NoError(t, PortForwardAPod(sourceReq))
	}()

	go func() {
		require.NoError(t, PortForwardAPod(targetReq))
	}()

	return stopCh
}

func configureComponents(t *testing.T) *terraform.Options {
	options := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./configure",
		Vars: map[string]interface{}{
			"target_vault_local_port": targetVaultPort,
			"source_vault_local_port": sourceVaultPort,
		},
	})

	_, err := terraform.InitAndApplyAndIdempotentE(t, options)
	require.NoError(t, err)

	return options
}

func newHVCJob(name string) *batchv1.Job {
	namespace := name
	parallelism := int32(1)
	completions := int32(1)
	backoffLimit := int32(0)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "copyjobspec",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "copyjob",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "spec.json",
											Path: "./spec.json",
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "hvc",
							Image: "marcboudreau/hvc:test",
							Args: []string{
								"copy",
								"/mnt/spec.json",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "copyjobspec",
									MountPath: "/mnt",
								},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: "hvc",
				},
			},
		},
	}
}

func verifySingleSecret(t *testing.T) {
	clientset := getKubernetesClientSet(t)

	job, err := clientset.BatchV1().Jobs("test1").Create(context.Background(), newHVCJob("test1"), metav1.CreateOptions{})
	require.NoError(t, err)

	time.Sleep(time.Duration(5 * time.Second))

	job, err = clientset.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), job.Status.Succeeded)
	t.Logf("Job for test %s duration: %s", t.Name(), job.Status.CompletionTime.Sub(job.Status.StartTime.Time))

	vaultClient, err := vaultapi.NewClient(nil)
	require.NoError(t, err)

	vaultClient.SetToken("root")
	vaultClient.SetAddress(fmt.Sprintf("http://localhost:%d", targetVaultPort))

	secret, err := vaultClient.Logical().Read("kv/data/tc1/secret1")
	require.NoError(t, err)
	assert.NotNil(t, secret)

	data := secret.Data["data"].(map[string]interface{})
	assert.Contains(t, data, "k1")
	assert.Equal(t, "path1/secret1", data["k1"].(string))
	assert.Contains(t, data, "k2")
	assert.Equal(t, "path2/secret1", data["k2"].(string))
	assert.Contains(t, data, "k3")
	assert.Equal(t, "path3/secret1", data["k3"].(string))
}

func verifySingleSecretMultipleSources(t *testing.T) {
	clientset := getKubernetesClientSet(t)

	job, err := clientset.BatchV1().Jobs("test2").Create(context.Background(), newHVCJob("test2"), metav1.CreateOptions{})
	require.NoError(t, err)

	time.Sleep(time.Duration(5 * time.Second))

	job, err = clientset.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), job.Status.Succeeded)
	t.Logf("Job for test %s duration: %s", t.Name(), job.Status.CompletionTime.Sub(job.Status.StartTime.Time))

	vaultClient, err := vaultapi.NewClient(nil)
	require.NoError(t, err)

	vaultClient.SetToken("root")
	vaultClient.SetAddress(fmt.Sprintf("http://localhost:%d", targetVaultPort))

	secret, err := vaultClient.Logical().Read("kv/data/tc2/secret1")
	require.NoError(t, err)
	assert.NotNil(t, secret)

	data := secret.Data["data"].(map[string]interface{})
	assert.Contains(t, data, "k1")
	assert.Equal(t, "path1/secret1", data["k1"].(string))
	assert.Contains(t, data, "k2")
	assert.Equal(t, "path2/secret1", data["k2"].(string))
	assert.Contains(t, data, "k3")
	assert.Equal(t, "path3/secret1", data["k3"].(string))
}

func verifyMultipleSecrets(t *testing.T) {
	clientset := getKubernetesClientSet(t)

	job, err := clientset.BatchV1().Jobs("test3").Create(context.Background(), newHVCJob("test3"), metav1.CreateOptions{})
	require.NoError(t, err)

	time.Sleep(time.Duration(5 * time.Second))

	job, err = clientset.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), job.Status.Succeeded)
	t.Logf("Job for test %s duration: %s", t.Name(), job.Status.CompletionTime.Sub(job.Status.StartTime.Time))

	vaultClient, err := vaultapi.NewClient(nil)
	require.NoError(t, err)

	vaultClient.SetToken("root")
	vaultClient.SetAddress(fmt.Sprintf("http://localhost:%d", targetVaultPort))

	secret1, err := vaultClient.Logical().Read("kv/data/tc3/secret1")
	require.NoError(t, err)
	assert.NotNil(t, secret1)

	data1 := secret1.Data["data"].(map[string]interface{})
	assert.Contains(t, data1, "k1")
	assert.Equal(t, "path1/secret1", data1["k1"].(string))
	assert.Contains(t, data1, "k2")
	assert.Equal(t, "path2/secret1", data1["k2"].(string))
	assert.Contains(t, data1, "k3")
	assert.Equal(t, "path3/secret1", data1["k3"].(string))

	secret2, err := vaultClient.Logical().Read("kv/data/tc3/secret2")
	require.NoError(t, err)
	assert.NotNil(t, secret2)

	data2 := secret2.Data["data"].(map[string]interface{})
	assert.Contains(t, data2, "k1")
	assert.Equal(t, "path1/secret2", data2["k1"].(string))
	assert.Contains(t, data2, "k2")
	assert.Equal(t, "path2/secret2", data2["k2"].(string))
	assert.Contains(t, data2, "k3")
	assert.Equal(t, "path3/secret2", data2["k3"].(string))
}

func verifyKubernetesAuthentication(t *testing.T) {
	clientset := getKubernetesClientSet(t)

	job, err := clientset.BatchV1().Jobs("test4").Create(context.Background(), newHVCJob("test4"), metav1.CreateOptions{})
	require.NoError(t, err)

	time.Sleep(time.Duration(5 * time.Second))

	job, err = clientset.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), job.Status.Succeeded)
	t.Logf("Job for test %s duration: %s", t.Name(), job.Status.CompletionTime.Sub(job.Status.StartTime.Time))

	vaultClient, err := vaultapi.NewClient(nil)
	require.NoError(t, err)

	vaultClient.SetToken("root")
	vaultClient.SetAddress(fmt.Sprintf("http://localhost:%d", targetVaultPort))

	secret, err := vaultClient.Logical().Read("kv/data/tc4/secret1")
	require.NoError(t, err)
	assert.NotNil(t, secret)

	data := secret.Data["data"].(map[string]interface{})
	assert.Contains(t, data, "k")
	assert.Equal(t, "path1/secret1", data["k"].(string))
}

func verifyAccessDenied(t *testing.T) {
	clientset := getKubernetesClientSet(t)

	job, err := clientset.BatchV1().Jobs("test5").Create(context.Background(), newHVCJob("test5"), metav1.CreateOptions{})
	require.NoError(t, err)

	time.Sleep(time.Duration(10 * time.Second))

	job, err = clientset.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(1), job.Status.Failed)
	t.Logf("%#v", job.Status)
}
