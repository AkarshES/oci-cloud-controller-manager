package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func calculateExponentialBackoff(currentInterval time.Duration, maxInterval time.Duration, jitterFactor float64) time.Duration {
	// Calculate the next interval using exponential backoff formula
	nextInterval := currentInterval * 2
	if nextInterval > maxInterval {
		nextInterval = maxInterval
	}

	// Apply random jitter to the next interval
	jitter := time.Duration(float64(nextInterval) * jitterFactor)
	minJitter := time.Duration(float64(nextInterval) * (1 - jitterFactor))

	jitterRange := jitter - minJitter

	if jitterRange > 0 {
		nextInterval += time.Duration(rand.Int63n(int64(jitterRange)) + int64(minJitter))
	} else {
		nextInterval += minJitter
	}

	return nextInterval
}

func getHealth(controller string) (string, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://[::]:%s/healthz/%s", MetricsPort, controller)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "unknown", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown", err
	}
	return string(body), nil
}

func getLiveness(controller string) (string, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://[::]:%s/readyz/%s", HealthPort, controller)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "unknown", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown", err
	}
	return string(body), nil
}

func shouldStartControllerManager(crdEnabled ...bool) bool {
	for _, v := range crdEnabled {
		if v == true {
			return true
		}
	}
	return false
}

func getOCIClient(logger *zap.SugaredLogger, config *providercfg.Config) client.Interface {
	c, err := client.GetClient(logger, config)

	if err != nil {
		logger.With(zap.Error(err)).Fatal("client can not be generated.")
	}
	return c
}

func buildRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	// Check for KUBECONFIG env var
	kubeconfigEnv := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if kubeconfigEnv != "" {
		// Use kubeconfig from environment variable
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Fallback: Use explicitly provided kubeconfigPath if set
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Fall back to in-cluster config if nothing else is provided
	return rest.InClusterConfig()
}
