/*
Copyright 2015 The Kubernetes Authors.
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

package e2e

import (
	"fmt"
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = framework.KubeDescribe("GPU job processing", func() {
	f := framework.NewDefaultFramework("gpu")
	var cs kubernetes.Interface

	BeforeEach(func() {
		cs = f.ClientSet
	})

	It("Should run a job on a gpu node [Slow] [Zalando] [GPU]", func() {
		ns := f.Namespace.Name
		nameprefix := "gpu-test-"
		labels := map[string]string{
			"application": "vector-add",
		}

		By("Creating a vector pod which runs on a GPU node")
		pod := createVectorPod(nameprefix, ns, labels)
		_, err := cs.CoreV1().Pods(ns).Create(pod)
		framework.ExpectNoError(err, fmt.Errorf("Could not create POD %s", pod.Name))
		framework.ExpectNoError(f.WaitForPodRunning(pod.Name))
		for {
			p, err := cs.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
				framework.ExpectNoError(err, fmt.Errorf("Could not get POD %s", pod.Name))
				return
			}
			if p.Status.ContainerStatuses[0].State.Terminated == nil {
				time.Sleep(10 * time.Second)
				continue
			}
			n := p.Status.ContainerStatuses[0].State.Terminated.ExitCode
			if n < 1 {
				logs, err := getPodLogs(cs, ns, pod.Name, "cuda-vector-add", false)
				framework.ExpectNoError(err, "Should be able to get logs for pod %v", pod.Name)
				regex := regexp.MustCompile("PASSED")
				if regex.MatchString(logs) {
					return
				}
				framework.ExpectNoError(err, "Expected vector job to succeed")
				return
			}
			framework.ExpectNoError(fmt.Errorf("Expected POD %s to terminate with exit code 0", pod.Name))
			return
		}
	})
})
