package orphanresources

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viveksahu26/orphaned_resource/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
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

func SendOrphanedResourceMetricsAlertInfo() error {
	orphangeResourceLabels := CountOrphanedResources(config.DiscoveryClient, config.DynamicClient)

	OrphanedResources.Set(float64(orphangeResourceLabels.totalCount))
	return nil
}

func CountOrphanedResources(discoveryClient *discovery.DiscoveryClient, dynamicClient *dynamic.DynamicClient) CountOrphanedResource {
	count := CountOrphanedResource{}

	countArgoCDResource := count
	countOrphanedResource := count
	countArgoCDResource, countOrphanedResource = GetOrphanedResource(discoveryClient, dynamicClient, countArgoCDResource, countOrphanedResource)

	log.Println("Total ArgoCD Managed Resource: ", countArgoCDResource)
	log.Println("Total Orphaned Resource: ", countOrphanedResource)
	return countOrphanedResource
}

func GetOrphanedResource(discoveryClient *discovery.DiscoveryClient, dynamicClient *dynamic.DynamicClient, totalArgoManagedResource, totalOrphanedResource CountOrphanedResource) (CountOrphanedResource, CountOrphanedResource) {
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
				fmt.Printf("Could not convert resource %s to JSON format: %v", resInfo.Name, err)
				continue
			}

			gvk := schema.GroupVersionResource{
				Group:    resInfo.Group,
				Version:  resInfo.Version,
				Resource: resInfo.Name,
			}

			getAllResource, err := dynamicClient.Resource(gvk).Namespace("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not retrieve resources from namespace 'default': %v", err)
				continue
			}

			for _, el := range getAllResource.Items {
				totalNumberOfResources++

				// get labels for each resource
				resourceLabels := el.GetLabels()

				labelValue, ok := resourceLabels["app.kubernetes.io/name"]
				if ok && strings.Contains(labelValue, "argocd") {
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
