# KubeCon EU 2018

These are my notes for Twitter demo prep. 


### After Setup

#### Switching

```
gcloud container clusters get-credentials n3wscott-ledhouse-demo-cluster --zone us-central1-a --project n3wscott-ledhouse-demo
```

or

```
gcloud container clusters get-credentials n3wscott-ledhouse-complex-cluster --zone us-central1-a --project n3wscott-ledhouse-complex
```

## Setup

### Installing Service Catalog

create gke cluster in ui, then:

```
gcloud container clusters get-credentials n3wscott-ledhouse-complex-cluster --zone us-central1-a --project n3wscott-ledhouse-complex
gcloud config set project n3wscott-ledhouse-complex
gcloud auth application-default login
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
sc install
kubectl get deployment -n service-catalog
sc add-gcp-broker
svcat describe broker gcp-broker
```

```
GCP_PROJECT_ID=$(gcloud config get-value project)
GCP_PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT_ID --format='value(projectNumber)')
gcloud projects add-iam-policy-binding ${GCP_PROJECT_ID} \
    --member serviceAccount:${GCP_PROJECT_NUMBER}@cloudservices.gserviceaccount.com \
    --role=roles/owner
```


### namespace

```
kubectl create namespace ledhouse
```

set the namespace:

```
kubectl config set-context $(kubectl config current-context) --namespace=ledhouse
```

#### Pub/Sub instance for demo

```
svcat provision demo-to-ledhouse-pubsub \
    --class cloud-pubsub \
    --plan beta \
    --param topicId=demo-to-ledhouse \
    --namespace ledhouse

svcat get instance --namespace ledhouse demo-to-ledhouse-pubsub

svcat bind --name demo-to-ledhouse-publisher --namespace ledhouse demo-to-ledhouse-pubsub --params-json '{
  "serviceAccount": "demo-publisher",
  "createServiceAccount": true,
  "roles": [
    "roles/pubsub.publisher",
    "roles/pubsub.viewer"
  ]
}'

svcat bind --name demo-to-ledhouse-subscriber --namespace ledhouse demo-to-ledhouse-pubsub --params-json '{
  "serviceAccount": "demo-subscriber",
  "createServiceAccount": true,
  "roles": [
    "roles/pubsub.subscriber",
    "roles/pubsub.viewer"
  ],
  "subscription": {
    "subscriptionId": "demo-to-ledhouse"
  }
}'

svcat get bindings --namespace ledhouse

svcat describe bindings --namespace ledhouse demo-to-ledhouse-publisher
kubectl get secret --namespace ledhouse demo-to-ledhouse-publisher -oyaml

svcat describe bindings --namespace ledhouse demo-to-ledhouse-subscriber
kubectl get secret --namespace ledhouse demo-to-ledhouse-subscriber -oyaml

```

### Provision pubsub for proxy <--> local

Create the topics,

```
svcat provision proxy-to-local-pubsub \
    --class cloud-pubsub \
    --plan beta \
    --param topicId=proxy-to-local \
    --namespace ledhouse

svcat provision local-to-proxy-pubsub \
    --class cloud-pubsub \
    --plan beta \
    --param topicId=local-to-proxy \
    --namespace ledhouse
```

bind to them, this makes the service accounts.

```
svcat bind proxy-to-local-pubsub --name proxy-to-local-publisher \
  --namespace ledhouse \
  --params-json '{
  "serviceAccount": "broker-proxy",
  "createServiceAccount": true,
  "roles": [
    "roles/pubsub.publisher",
    "roles/pubsub.viewer"
  ]
}'

svcat bind local-to-proxy-pubsub --name local-to-proxy-publisher \
  --namespace ledhouse \
  --params-json '{
  "serviceAccount": "local-broker",
  "createServiceAccount": true,
  "roles": [
    "roles/pubsub.publisher",
    "roles/pubsub.viewer"
  ]
}'

svcat bind proxy-to-local-pubsub --name proxy-to-local-subscriber \
  --namespace ledhouse \
  --params-json '{
  "serviceAccount": "local-broker",
  "roles": [
    "roles/pubsub.subscriber",
    "roles/pubsub.viewer"
  ],
  "subscription": {
    "subscriptionId": "proxy-to-local"
  }
}'

svcat bind local-to-proxy-pubsub --name local-to-proxy-subscriber \
  --namespace ledhouse \
  --params-json '{
  "serviceAccount": "broker-proxy",
  "roles": [
    "roles/pubsub.subscriber",
    "roles/pubsub.viewer"
  ],
  "subscription": {
    "subscriptionId": "local-to-proxy"
  }
}'
```

### Deploy the proxy

```
kubectl create -f ./twitter/manifests/proxy-deployment.yaml
```

### Connect local to proxy

```
kubectl get secret --namespace ledhouse local-to-proxy-publisher -ojson
kubectl get secret --namespace ledhouse proxy-to-local-subscriber -ojson
```

### Register service catalog to the proxy

```
kubectl create -f ./proxy-broker.yaml
```

### Provision and bind the lights

```
kubectl apply -f ./manifests/demo-lights.yaml
kubectl apply -f ./manifests/demo-light-bindings.yaml
```

### Make the twitter secret

```
kubectl create secret generic twitter \
    --from-literal=consumerKey=XXX \
    --from-literal=consumerSecret=YYY \
    --from-literal=token=ZZZ \
    --from-literal=tokenSecret=AAA
```

### Upload the demo image

once: `gcloud auth configure-docker`

```
make upload
```

### deploy the demo app

```
kubectl create -f ./manifests/demo-deployment.yaml
```

## Useful,

### Relist the catalog

```
svcat sync broker broker-proxy
```