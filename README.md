# Latency Sensitive FaaS

## Requirements
- Docker/podman (I used Docker desktop: https://docs.docker.com/desktop/setup/install/windows-install/)
- minikube: https://minikube.sigs.k8s.io/docs/start/?arch=%2Fwindows%2Fx86-64%2Fstable%2F.exe+download
- kubectl: https://kubernetes.io/docs/tasks/tools/
- istioctl: https://github.com/istio/istio/releases/tag/1.23.2
- knative cli: https://knative.dev/docs/client/install-kn/
- helm: https://helm.sh/docs/intro/install/

  ## Initial setup (Windows)

  1. Open a terminal and cd into the `./kubernetes` folder
  2. Open the `cluster_setup.bat` script. You can set the number of nodes, and the CPU and memory reservations for each node using the variables at the top of the file. Unfortunately, ElasticSearch needs a lot of memory to function, so do not lower the memory limit too much.
  3. Run the `cluster_setup.bat` script and wait for it to finish. This will create the minikube cluster and start a tunnel for DNS services. Keep this terminal window open while using the cluster.
  4. Run the `elastic_with_eck_setup.bat` script (located in the same folder). This will setup the OpenTelemetry operator, the OpenTelemetry collector, and the required components from the Elastic stack. Once the script finishes it will also port forward Kibana to the local port `5601`, so you can reach it on localhost (keep this terminal window open aswell).
  5. Check the `./tools` directory. Extract the `lsfunc.zip` archive and copy the exe to your desired location and add it to the PATH environment variable, so you can use it from the terminal.
  6. Open a terminal and cd to the `./loadbalanced-app` directory. Run the `deploy.bat` script.


  ## Testing

  The easiest way is to use Apache JMeter: https://jmeter.apache.org/download_jmeter.cgi

  1. Open up the JMeter GUI and click on `File -> Open` in the top menu. Import the  `./loadbalanced-app/load-testing/loadbalanced-app.jmx` file.
  2. In the tree on the left side navigate to `Latency Sensitive Faas -> Thread Group -> HTTP Request -> JSR223 PreProcessor`. In the script section change the `filePath` variable to a valid path pointing to an image on your computer.
  3. Click on the Start button. 

  This will periodically send a request with the given image to our application. You can monitor the responses and potential errors with the JMeter Summary report or in Kibana.

  Kibana username: `elastic`
  
  Kibana password: `elastic`


