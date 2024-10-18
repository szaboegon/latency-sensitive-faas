minikube build -t localhost/faas-loadbalancer -f .\deploy\Dockerfile .
kubectl apply -f .\deploy\faas-loadbalancer.yaml