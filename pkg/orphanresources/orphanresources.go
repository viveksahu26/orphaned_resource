package orphanresources

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/viveksahu26/orphaned_resource/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ResourceInfo struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
}

type CountOrphanedResource struct {
	totalCount int
	Resources  []string
}

// OrphanedResources is the total no. of orphaned resources
var OrphanedResources = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: metricOrphanedResourceCount,
		Help: helpMetricOrphanedResourceCount,
	},
)

// SendOrphanedResourceMetricsAlertInfo func for monitor orphan resources
func SendOrphanedResourceMetricsAlertInfo() error {
	log.Println("+++++ Inside SendOrphanedResourceMetricsAlertInfo")
	defer log.Println("------ Exit SendOrphanedResourceMetricsAlertInfo")
	orphangeResourceLabels := CountOrphanedResources(config.KubeConfig)
	log.Println("**************** TOTAL TOTAL orphangeResourceLabels.totalCount: ", orphangeResourceLabels.totalCount)
	OrphanedResources.Set(float64(orphangeResourceLabels.totalCount))
	return nil
}

func CountOrphanedResources(kubeconfig *rest.Config) CountOrphanedResource {
	log.Println("+++++ Inside CountOrphanedResources")
	defer log.Println("------ Exit CountOrphanedResources")
	count := CountOrphanedResource{}

	countArgoCDResource := count
	countOrphanedResource := count
	countArgoCDResource, countOrphanedResource = GetOrphanedResource(kubeconfig, countArgoCDResource, countOrphanedResource)

	log.Println("Total ArgoCD Managed Resource: ", countArgoCDResource.totalCount)
	log.Println("Total Orphaned Resource: ", countOrphanedResource.totalCount)
	return countOrphanedResource
}

func GetOrphanedResource(config *rest.Config, totalArgoManagedResource, totalOrphanedResource CountOrphanedResource) (CountOrphanedResource, CountOrphanedResource) {
	log.Println("+++++ Inside GetOrphanedResource")
	defer log.Println("------ Exit GetOrphanedResource")

	// creates new kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("error in creating kuberentes client: %v\n", err)
		os.Exit(1)
	}
	argocdClientset, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Argo CD clientset: %v", err)
	}
	log.Println("argocdClientset: ", argocdClientset)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	discoveryClient := clientset.DiscoveryClient

	// fetches all resources from API Server
	resources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		panic(err.Error())
	}
	totalNumberOfResources := 0
	for _, apiGroup := range resources {
		if len(apiGroup.APIResources) == 0 {
			continue
		}

		grp := strings.Split(apiGroup.GroupVersion, "/")
		if len(grp) == 1 {
			grp = append(grp, "")
		}
		// when resource only have version but not groups
		if (grp[0] == "v1") && (grp[1] == "") {
			grp[1] = grp[0]
			grp[0] = ""
		}

		// get resource group and version
		resGroup, resVersion := grp[0], grp[1]

		// looping around list of API resources
		for _, resource := range apiGroup.APIResources {
			if len(resource.Verbs) == 0 {
				log.Printf("Resource %s doesn't contain any verbs", resource.Name)
				continue
			}
			resInfo := ResourceInfo{
				Group: resGroup, Version: resVersion, Kind: resource.Kind, Name: resource.Name,
			}

			_, err := json.Marshal(&resInfo)
			if err != nil {
				log.Printf("Could not convert resource %s to JSON format: %v", resInfo.Name, err)
				continue
			}

			gvk := schema.GroupVersionResource{
				Group:    resInfo.Group,
				Version:  resInfo.Version,
				Resource: resInfo.Name,
			}

			getAllResource, err := dynamicClient.Resource(gvk).Namespace("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("Could not retrieve resources from namespace: %v", err)
				continue
			}

			for _, el := range getAllResource.Items {
				log.Println("////////////////")
				totalNumberOfResources++

				// get labels for each resource
				resourceLabels := el.GetLabels()

				resLabel1, ok := resourceLabels["app.kubernetes.io/instance"]
				resLabel2, ok1 := resourceLabels["app.kubernetes.io/part-of"]
				resLabel3, ok3 := resourceLabels["app.kubernetes.io/name"]
				if ok && strings.Contains(resLabel1, "argo-cd") || ok1 && strings.Contains(resLabel2, "argocd") || ok3 && strings.Contains(resLabel3, "argocd") {
					totalArgoManagedResource.totalCount++
					totalArgoManagedResource.Resources = append(totalArgoManagedResource.Resources, resInfo.Name)
				} else {
					totalOrphanedResource.Resources = append(totalOrphanedResource.Resources, resInfo.Name)
				}
			}
		}
	}
	log.Println("Total no. of resources present in API Server: ", totalNumberOfResources)
	totalOrphanedResource.totalCount = totalNumberOfResources - totalArgoManagedResource.totalCount
	return totalArgoManagedResource, totalOrphanedResource
}
