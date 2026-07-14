/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"sync"

	kruise "github.com/openkruise/kruise-api/apps/v1alpha1"
	commonconfig "github.com/koderover/zadig/v2/pkg/config"
	"github.com/koderover/zadig/v2/pkg/setting"
	"github.com/koderover/zadig/v2/pkg/tool/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

var once sync.Once

var c cluster.Cluster
var clusterDisabled bool

func init() {
	ctrl.SetLogger(klog.Background())
}

// Cluster is a singleton, it will be initialized only once.
func Cluster() cluster.Cluster {
	once.Do(func() {
		if commonconfig.Mode() == setting.DebugMode {
			clusterDisabled = true
			log.Warnf("skip kube cluster init in debug mode")
			return
		}
		restConfig, err := ctrl.GetConfig()
		if err != nil {
			panic(err)
		}
		var initErr error
		c, initErr = initCluster(restConfig)
		if initErr != nil {
			panic(initErr)
		}
	})

	return c
}

func Client() client.Client {
	if clusterDisabled || c == nil {
		return nil
	}
	return Cluster().GetClient()
}

func APIReader() client.Reader {
	if clusterDisabled || c == nil {
		return nil
	}
	return Cluster().GetAPIReader()
}

func RESTConfig() *rest.Config {
	if clusterDisabled || c == nil {
		return nil
	}
	return Cluster().GetConfig()
}

func Scheme() *runtime.Scheme {
	if clusterDisabled || c == nil {
		return nil
	}
	return Cluster().GetScheme()
}

func Start(ctx context.Context) error {
	if clusterDisabled || c == nil {
		return nil
	}
	return Cluster().Start(ctx)
}

// apiClient is similar with the default Client(), but it always gets objects from API server.
type apiClient struct {
	client.Client

	apiReader client.Reader
}

func (c *apiClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return c.apiReader.Get(ctx, key, obj, opts...)
}

func (c *apiClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.apiReader.List(ctx, list, opts...)
}

func initCluster(restConfig *rest.Config) (cluster.Cluster, error) {
	scheme := runtime.NewScheme()

	// add all known types
	// if you want to support custom types, call _ = yourCustomAPIGroup.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kruise.AddToScheme(scheme)

	c, err := cluster.New(restConfig, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = scheme
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init client")
	}

	return c, nil
}
