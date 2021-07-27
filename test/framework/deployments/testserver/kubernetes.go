package testserver

import (
	"fmt"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	opts DeploymentOpts
}

func (k *k8SDeployment) Name() string {
	return "test-server"
}

func (k *k8SDeployment) Deploy(cluster framework.Cluster) error {
	const name = "test-server"
	service := `
apiVersion: v1
kind: Service
metadata:
  name: test-server
  namespace: kuma-test
  annotations:
    80.service.kuma.io/protocol: http
spec:
  ports:
    - port: 80
      name: http
  selector:
    app: test-server
`
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-server
  namespace: kuma-test
  labels:
    app: test-server
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: test-server
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: test-server
    spec:
      containers:
        - name: test-server
          image: %s
          imagePullPolicy: IfNotPresent
          readinessProbe:
            httpGet:
              path: /
              port: 80
            initialDelaySeconds: 3
            periodSeconds: 3
          ports:
            - containerPort: 80
          command: [ "test-server" ]
          args:
            - echo
            - "--instance"
            - "echo"
            - "--port"
            - '80'
          resources:
            limits:
              cpu: 50m
              memory: 128Mi
`
	fn := framework.Combine(
		framework.YamlK8s(service),
		framework.YamlK8s(fmt.Sprintf(deployment, k.opts.Mesh, framework.GetUniversalImage())),
		framework.WaitService(framework.TestNamespace, name),
		framework.WaitNumPods(1, name),
		framework.WaitPodsAvailable(framework.TestNamespace, name),
	)
	return fn(cluster)
}

func (k *k8SDeployment) Delete(cluster framework.Cluster) error {
	return nil // todo
}

var _ Deployment = &k8SDeployment{}
