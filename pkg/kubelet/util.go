/*
Copyright 2014 Google Inc. All rights reserved.

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

package kubelet

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/capabilities"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/record"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

// TODO: move this into a pkg/tools/etcd_tools
func EtcdClientOrDie(etcdServerList util.StringList, etcdConfigFile string) tools.EtcdClient {
	if len(etcdServerList) > 0 {
		return etcd.NewClient(etcdServerList)
	} else if etcdConfigFile != "" {
		etcdClient, err := etcd.NewClientFromFile(etcdConfigFile)
		if err != nil {
			glog.Fatalf("Error with etcd config file: %v", err)
		}
		return etcdClient
	}
	return nil
}

// TODO: move this into pkg/capabilities
func SetupCapabilities(allowPrivileged bool) {
	capabilities.Initialize(capabilities.Capabilities{
		AllowPrivileged: allowPrivileged,
	})
}

// TODO: Split this up?
func SetupLogging() {
	etcd.SetLogger(util.NewLogger("etcd "))
	// Log the events locally too.
	record.StartLogging(glog.Infof)
}

func SetupEventSending(client *client.Client, hostname string) {
	glog.Infof("Sending events to api server.")
	record.StartRecording(client.Events(""))
}
