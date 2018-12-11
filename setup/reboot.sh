#!/bin/bash

PROJECT_ID=joe-does-knative
CLUSTER_NAME=knative-backup
CLUSTER_ZONE=us-east1-d
CLUSTER_INGRESS_IP=35.185.37.171
CLUSTER_DOMAIN_NAME=backup.josephburnett.com

# Cleanup
gcloud container clusters delete $CLUSTER_NAME \
       --project=$PROJECT_ID \
       --zone=$CLUSTER_ZONE \
       --quiet

set -e

# Create Cluster
# Based on https://github.com/knative/docs/blob/master/install/Knative-with-GKE.md#creating-a-kubernetes-cluster
gcloud container clusters create $CLUSTER_NAME \
       --project=$PROJECT_ID \
       --zone=$CLUSTER_ZONE \
       --cluster-version=latest \
       --machine-type=n1-standard-4 \
       --enable-autoscaling --min-nodes=1 --max-nodes=10 \
       --enable-autorepair \
       --scopes=service-control,service-management,compute-rw,storage-ro,cloud-platform,logging-write,monitoring-write,pubsub,datastore \
       --num-nodes=3 \
       --quiet
kubectl create clusterrolebinding cluster-admin-binding \
	--clusterrole=cluster-admin \
	--user=$(gcloud config get-value core/account)

# Deploy Knative Serving
cd ~/go/src/github.com/knative/serving
# Based on https://github.com/knative/serving/blob/master/DEVELOPMENT.md
kubectl apply -f ./third_party/istio-1.0.4/istio-crds.yaml
while [ $(kubectl get crd gateways.networking.istio.io -o jsonpath='{.status.conditions[?(@.type=="Established")].status}') != 'True' ]; do
  echo "Waiting on Istio CRDs"; sleep 1
done
kubectl apply -f ./third_party/istio-1.0.4/istio.yaml
kubectl apply -f ./third_party/config/build/release.yaml
KO_DOCKER_REPO=gcr.io/$PROJECT_ID ko apply -f config/

# Attach static IP to ingress gateway
# Based on https://github.com/knative/serving/blob/master/docs/setting-up-ingress-static-ip.md
kubectl patch svc istio-ingressgateway \
	--namespace=istio-system \
	--patch="{\"spec\": { \"loadBalancerIP\": \"$CLUSTER_INGRESS_IP\" }}"

# Adjust Autoscaler Cluster Parameters
kubectl patch configmap config-autoscaler \
	--namespace=knative-serving \
	--patch '{"data":{"container-concurrency-target-default":"10"}}'
kubectl patch configmap config-autoscaler \
	--namespace=knative-serving \
	--patch '{"data":{"scale-to-zero-threshold":"1m"}}'
kubectl patch configmap config-autoscaler \
	--namespace=knative-serving \
	--patch '{"data":{"scale-to-zero-grace-period":"30s"}}'

# Setup Custom Domain
kubectl patch configmap config-domain \
	--namespace=knative-serving \
	--type json \
	--patch "[{\"op\":\"remove\",\"path\":\"/data/example.com\"},{\"op\":\"add\",\"path\":\"/data/$CLUSTER_DOMAIN_NAME\",\"value\":\"\"}]"

# Deploy Yolo Controller
cd ~/go/src/github.com/josephburnett/kubecon-seattle-2018
ko apply -f yolo/artifacts/

# Deploy Sample App
sleep 60 # I get app-00000 if I create the service too soon !?
kubectl create ns kubecon-seattle-2018
kubectl apply -f app/service.yaml
sleep 30 # Let the app startup
echo
curl "http://app.kubecon-seattle-2018.${CLUSTER_DOMAIN_NAME}/"
