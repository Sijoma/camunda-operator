# camunda-operator

https://camunda.com

## Description

Operator for Camunda - Use at your own risk. This is an alpha version.

```yaml
apiVersion: core.camunda.io/v1alpha1
kind: OrchestrationCluster
metadata:
  name: camunda
spec:
  version: 8.7.7
  clusterSize: 3
  partitionCount: 3
  replicationFactor: 3
  resources:
    requests:
      cpu: 1000m
      memory: 1500Mi
    limits:
      cpu: 1500m
      memory: 1500Mi
  database:
    type: elasticsearch
    hostName: "http://elasticsearch-es-http:9200"
    userName: elastic
    password:
      key: elastic
      name: elasticsearch-es-elastic-user
```

## Getting Started

### Deploying the Operator with dependencies

1. Install Prerequisites 
- kubectl 
- Kubernetes cluster (v1.24+ recommended)

2. Install ECK Operator (Elasticsearch) [Here](https://www.elastic.co/docs/deploy-manage/deploy/cloud-on-k8s/install-using-yaml-manifest-quickstart)
```shell
kubectl create -f https://download.elastic.co/downloads/eck/3.0.0/crds.yaml
kubectl apply -f https://download.elastic.co/downloads/eck/3.0.0/operator.yaml
```

3. Deploy Elasticsearch cluster:
```shell
cat <<EOF | kubectl apply -f -
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: elasticsearch
spec:
  version: 8.17.1
  nodeSets:
    - name: default
      count: 1
      config:
        node.store.allow_mmap: false
  http:
    tls:
      selfSignedCertificate:
        disabled: true
EOF
```

4. Install Camunda Operator
```shell
kubectl apply -f https://github.com/Sijoma/camunda-operator/releases/latest/download/install.yaml
```

5. Deploy OrchestrationCluster Resource
```sh
cat <<EOF | kubectl apply -f -
apiVersion: core.camunda.io/v1alpha1
kind: OrchestrationCluster
metadata:
  name: camunda
spec:
  version: 8.7.7
  clusterSize: 3
  partitionCount: 3
  replicationFactor: 3
  database:
    type: elasticsearch
    hostName: "http://elasticsearch-es-http:9200"
    userName: elastic
    password:
      key: elastic
      name: elasticsearch-es-elastic-user
      optional: false
EOF
```

## Contributing

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
