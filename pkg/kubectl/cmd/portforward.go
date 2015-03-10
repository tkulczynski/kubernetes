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

package cmd

import (
	"os"
	"os/signal"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/portforward"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const (
	portforward_example = `$ kubectl port-forward -p mypod 5000 6000
<listens on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod>

$ kubectl port-forward -p mypod 8888:5000
<listens on port 8888 locally, forwarding to 5000 in the pod>

$ kubectl port-forward -p mypod :5000
<listens on a random port locally, forwarding to 5000 in the pod>

$ kubectl port-forward -p mypod 0:5000
<listens on a random port locally, forwarding to 5000 in the pod> `
)

func (f *Factory) NewCmdPortForward() *cobra.Command {
	flags := &struct {
		pod       string
		container string
	}{}

	cmd := &cobra.Command{
		Use:     "port-forward -p <pod> [<local port>:]<remote port> [<port>...]",
		Short:   "Forward 1 or more local ports to a pod.",
		Long:    "Forward 1 or more local ports to a pod.",
		Example: portforward_example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(flags.pod) == 0 {
				usageError(cmd, "<pod> is required for exec")
			}

			if len(args) < 1 {
				usageError(cmd, "at least 1 <port> is required for port-forward")
			}

			namespace, err := f.DefaultNamespace(cmd)
			util.CheckErr(err)

			client, err := f.Client(cmd)
			util.CheckErr(err)

			pod, err := client.Pods(namespace).Get(flags.pod)
			util.CheckErr(err)

			if pod.Status.Phase != api.PodRunning {
				glog.Fatalf("Unable to execute command because pod is not running. Current status=%v", pod.Status.Phase)
			}

			config, err := f.ClientConfig(cmd)
			util.CheckErr(err)

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			defer signal.Stop(signals)

			stopCh := make(chan struct{}, 1)
			go func() {
				<-signals
				close(stopCh)
			}()

			req := client.RESTClient.Get().
				Prefix("proxy").
				Resource("minions").
				Name(pod.Status.Host).
				Suffix("portForward", namespace, flags.pod)

			pf, err := portforward.New(req, config, args, stopCh)
			util.CheckErr(err)

			err = pf.ForwardPorts()
			util.CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&flags.pod, "pod", "p", "", "Pod name")
	// TODO support UID
	return cmd
}
