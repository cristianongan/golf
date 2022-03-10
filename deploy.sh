#! /bin/bash

RELEASE_NAME=$1
NAMESPACE=$2
IMAGE_TAG=$3

KUBE_CONFIG_FILE=$4 

VAR_FILE=./helm/$5 

DEPLOYS=$(helm --kubeconfig ${KUBE_CONFIG_FILE} --namespace=${NAMESPACE} ls | grep $RELEASE_NAME | wc -l)
if [ ${DEPLOYS}  -eq 0 ]; 
then helm install --kubeconfig ${KUBE_CONFIG_FILE} --namespace=${NAMESPACE} -f ${VAR_FILE} --set imageTag=${IMAGE_TAG} ${RELEASE_NAME} --description ${IMAGE_TAG} ./helm; 
else helm upgrade --kubeconfig ${KUBE_CONFIG_FILE} --namespace=${NAMESPACE} -f ${VAR_FILE} --set imageTag=${IMAGE_TAG} ${RELEASE_NAME} --description ${IMAGE_TAG} ./helm; 
fi