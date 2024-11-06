@echo off

echo echo Starting minikube cluster...
minikube start -p knative
kubectl apply -f knative\disable_scale_to_zero.yaml
kubectl apply -f knative\serving_features.yaml
echo Starting tunnel...
minikube tunnel -p knative