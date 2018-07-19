// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	capacity_v1 "github.com/supergiant/capacity/pkg/apis/capacity/v1"
	versioned "github.com/supergiant/capacity/pkg/client/clientset/versioned"
	internalinterfaces "github.com/supergiant/capacity/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/supergiant/capacity/pkg/client/listers/capacity/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterCapacityInformer provides access to a shared informer and lister for
// ClusterCapacities.
type ClusterCapacityInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ClusterCapacityLister
}

type clusterCapacityInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewClusterCapacityInformer constructs a new informer for ClusterCapacity type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterCapacityInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterCapacityInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredClusterCapacityInformer constructs a new informer for ClusterCapacity type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterCapacityInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CapacityV1().ClusterCapacities(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CapacityV1().ClusterCapacities(namespace).Watch(options)
			},
		},
		&capacity_v1.ClusterCapacity{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterCapacityInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterCapacityInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterCapacityInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&capacity_v1.ClusterCapacity{}, f.defaultInformer)
}

func (f *clusterCapacityInformer) Lister() v1.ClusterCapacityLister {
	return v1.NewClusterCapacityLister(f.Informer().GetIndexer())
}
