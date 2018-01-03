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

	kubeconfig      *string
	master          *string
	crdPollInterval *time.Duration

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

	crdPollInterval = fs.Duration("crd-poll-interval", 5*time.Minute, "period at which to check for custom resources")

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

	// Check if the FluxHelmResources exist in the cluster
	//		later on add a check that the CRD itself exists and creat it if not

	fmt.Println("I am functional!")

	// get clientset

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building integrations clientset: %v", err)
	}

	for {
		/*
			// create CRD:
			/*
			A CRD manifest is loaded before helm-operator starts
		*/
		list, err := client.IntegrationsV1().FluxHelmResources("kube-system").List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("Error listing all fluxhelmresources: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		fmt.Printf(">>> found %v items\n", len(list.Items))

		for _, fhr := range list.Items {
			fmt.Printf("fluxhelmresource %s with image %q, tag %q and namespace %q\n", fhr.Name, fhr.Spec.Image, fhr.Spec.ImageTag, fhr.Spec.Namespace)

			fmt.Printf(">>> found %v parameters\n", len(fhr.Spec.Customization))

			for _, cp := range fhr.Spec.Customization {
				fmt.Printf("\t\tcustomization with \n\t\tname %q\n\t\tvalue %q\n\t\ttype %q\n", cp.Name, cp.Value, cp.Type)
			}

		}

		time.Sleep(5 * time.Minute)
	}
}

// set up cluster tools
// kubectl
// cluster ?

// Watch for changes in Flux-Helm CRDs

/*
	// Example Controller
	// Watch for changes in Example objects and fire Add, Delete, Update callbacks
	_, controller := cache.NewInformer(
		crdclient.NewListWatch(),
		&crd.Example{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("add: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("delete: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("Update old: %s \n      New: %s\n", oldObj, newObj)
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	// Wait forever
select {}
*/

// Helper functions
func optionalVar(fs *pflag.FlagSet, value ssh.OptionalValue, name, usage string) ssh.OptionalValue {
	fs.Var(value, name, usage)
	return value
}
