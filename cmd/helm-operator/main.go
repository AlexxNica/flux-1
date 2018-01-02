package main

import (
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"fmt"
	"os"
	"os/signal"

	"github.com/go-kit/kit/log"
	/*
		"github.com/coreos/etcd-operator/pkg/client"
		"github.com/coreos/etcd-operator/pkg/controller"
		"github.com/coreos/etcd-operator/pkg/debug"
		"github.com/coreos/etcd-operator/pkg/util/constants"
		"github.com/coreos/etcd-operator/pkg/util/k8sutil"
		"github.com/coreos/etcd-operator/pkg/util/probe"
		"github.com/coreos/etcd-operator/pkg/util/retryutil"
		"github.com/coreos/etcd-operator/version"
	*/ //	"github.com/prometheus/client_golang/prometheus"
	//	"github.com/sirupsen/logrus"

	//"github.com/weaveworks/flux/git"

	"github.com/weaveworks/flux/ssh"

	"github.com/golang/glog"

	clientset "github.com/weaveworks/flux/integrations/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	fs      *pflag.FlagSet
	err     error
	logger  log.Logger
	kubectl string

	kubeconfig *string
	master     *string

	customKubectl *string
	gitURL        *string
	gitBranch     *string
	gitPath       *string

	k8sSecretName            *string
	k8sSecretVolumeMountPath *string
	k8sSecretDataKey         *string
	sshKeyBits               ssh.OptionalValue
	sshKeyType               ssh.OptionalValue

	name       *string
	listenAddr *string
	gcInterval *time.Duration

	createCRD *bool
)

func init() {
	// Flags processing
	fs = pflag.NewFlagSet("default", pflag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "DESCRIPTION\n")
		fmt.Fprintf(os.Stderr, "  helm-operator is a Kubernetes operator for Helm integration into flux.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		fs.PrintDefaults()
	}

	kubeconfig = fs.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master = fs.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	customKubectl = fs.String("kubernetes-kubectl", "", "Optional, explicit path to kubectl tool")
	gitURL = fs.String("git-url", "", "URL of git repo with Kubernetes manifests; e.g., git@github.com:weaveworks/flux-example")
	gitBranch = fs.String("git-branch", "master", "branch of git repo to use for Kubernetes manifests")
	gitPath = fs.String("git-path", "", "path within git repo to locate Kubernetes manifests (relative path)")

	// k8s-secret backed ssh keyring configuration
	k8sSecretName = fs.String("k8s-secret-name", "flux-git-deploy", "Name of the k8s secret used to store the private SSH key")
	k8sSecretVolumeMountPath = fs.String("k8s-secret-volume-mount-path", "/etc/fluxd/ssh", "Mount location of the k8s secret storing the private SSH key")
	k8sSecretDataKey = fs.String("k8s-secret-data-key", "identity", "Data key holding the private SSH key within the k8s secret")
	// SSH key generation
	sshKeyBits = optionalVar(fs, &ssh.KeyBitsValue{}, "ssh-keygen-bits", "-b argument to ssh-keygen (default unspecified)")
	sshKeyType = optionalVar(fs, &ssh.KeyTypeValue{}, "ssh-keygen-type", "-t argument to ssh-keygen (default unspecified)")

	// Setup logging
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
}

func main() {

	fs.Parse(os.Args)

	// Shutdown
	errc := make(chan error)

	// Shutdown trigger for goroutines
	shutdown := make(chan struct{})
	shutdownWg := &sync.WaitGroup{}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	defer func() {
		// wait until stopping
		logger.Log("exiting...", <-errc)
		close(shutdown)
		shutdownWg.Wait()
	}()

	fmt.Println("I am functional!")

	/*
		// Platform component.

		var clusterVersion string
		var sshKeyRing ssh.KeyRing
		var k8s cluster.Cluster
		var k8sManifests cluster.Manifests
	*/
	{
		// get client config

		/*
			restClientConfig, err := rest.InClusterConfig()
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}

			restClientConfig.QPS = 50.0
			restClientConfig.Burst = 100
		*/

		// get clientset

		cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
		if err != nil {
			glog.Fatalf("Error building kubeconfig: %v", err)
		}

		client, err := clientset.NewForConfig(cfg)
		if err != nil {
			glog.Fatalf("Error building integrations clientset: %v", err)
		}

		list, err := client.IntegrationsV1().FluxHelmResources("kube-system").List(metav1.ListOptions{})
		if err != nil {
			glog.Fatalf("Error listing all fluxhelmresources: %v", err)
		}

		fmt.Printf(">>> found %v items\n", len(list.Items))

		for {
			for _, fhr := range list.Items {
				fmt.Printf("fluxhelmresource %s with image %q, tag %q\n", fhr.Name, fhr.Spec.Image, fhr.Spec.ImageTag)
			}
			time.Sleep(5 * time.Minute)
		}
		/*
			clientset, err := kubernetes.NewForConfig(restClientConfig)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}

			for {
				fmt.Printf(">>> %#v\n\n", clientset.Endpoints("default"))
				fmt.Printf(">>> %#v\n\n", restClientConfig)

				time.Sleep(time.Duration(10 * time.Hour))
			}
		*/
	}
	/*
		// set up cluster tools
			// kubectl
			// cluster ?

		// create CRD

		// create CRD client interface

		// Watch for changes in Flux-Helm CRDs

	*/
}

// Helper functions
func optionalVar(fs *pflag.FlagSet, value ssh.OptionalValue, name, usage string) ssh.OptionalValue {
	fs.Var(value, name, usage)
	return value
}
