import requests
import time
import statistics
import json
import sys
import subprocess
from kubernetes import client, config
from typing import Any, List
import psutil

# -------------------------------------------------------------------
# Config
# -------------------------------------------------------------------
FUNCTION_NAME: str = "my-func"                  # Kubernetes deployment/service name
DOCKER_IMAGE: str = "szaboegon/my-func:latest"     # Docker image to build
DOCKERFILE_PATH: str = "./server"                       # Folder containing Dockerfile + server.py + function
DEPLOY_YAML: str = "./server/deploy.yaml"               # Kubernetes Deployment/Service YAML
LOCAL_PORT: int = 8080                            # Local port to forward to function
SVC_PORT: int = 80                        # Container port in pod
NAMESPACE: str = "default"
N: int = 1000                                  # default number of requests
# -------------------------------------------------------------------

# ------------------------- Docker Build ----------------------------
def build_docker_image() -> None:
    print(f"Building Docker image {DOCKER_IMAGE}...")
    subprocess.run(["docker", "build", "-t", DOCKER_IMAGE, DOCKERFILE_PATH], check=True)
    # Optional: push image if needed
    subprocess.run(["docker", "push", DOCKER_IMAGE], check=True)

# ------------------------ Kubernetes Deploy -----------------------
def deploy_to_k8s() -> None:
    print("Deploying to Kubernetes...")
    subprocess.run(["kubectl", "apply", "-f", DEPLOY_YAML], check=True)
    print(f"Waiting for deployment {FUNCTION_NAME} to be ready...")
    subprocess.run(["kubectl", "rollout", "status", f"deployment/{FUNCTION_NAME}"], check=True)
    
def get_pod_name(deployment_name: str, namespace: str) -> str:
    """
    Return the first pod name for the given deployment.
    """
    v1 = client.CoreV1Api()
    pods = v1.list_namespaced_pod(
        namespace=namespace,
        label_selector=f"app={deployment_name}"
    )
    if not pods.items:
        raise RuntimeError(f"No pods found for deployment {deployment_name}")
    return str(pods.items[0].metadata.name)

# ------------------------ Port Forward ----------------------------
def kill_existing_port_forwards() -> None:
    """
    Kill all running 'kubectl port-forward' processes.
    """
    for proc in psutil.process_iter(["pid", "name", "cmdline"]):
        try:
            cmdline = proc.info["cmdline"]
            if cmdline and "kubectl" in cmdline[0] and "port-forward" in cmdline:
                print(f"Killing existing port-forward PID {proc.info['pid']}: {' '.join(cmdline)}")
                proc.kill()
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            pass
        
def port_forward(max_retries: int = 5, delay: int = 3) -> subprocess.Popen[Any]:
    """
    Port-forward the service to localhost with retries if the pod is not yet ready.
    """
    kill_existing_port_forwards()
    
    for attempt in range(1, max_retries + 1):
        print(f"Port-forward attempt {attempt}/{max_retries} for {FUNCTION_NAME}...")
        pf: subprocess.Popen[Any] = subprocess.Popen(
            ["kubectl", "port-forward", f"svc/{FUNCTION_NAME}", f"{LOCAL_PORT}:{SVC_PORT}"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        # Give a few seconds for port-forward to start
        time.sleep(delay)

        # Check if port-forward is running by seeing if the process is still alive
        if pf.poll() is None:
            print(f"Port-forward established at localhost:{LOCAL_PORT}")
            return pf
        else:
            print(f"Port-forward failed, retrying in {delay} seconds...")
            # Cleanup before retrying
            pf.terminate()
            pf.wait()
            time.sleep(delay)

    raise RuntimeError(f"Failed to port-forward {FUNCTION_NAME} after {max_retries} attempts")

# ------------------------ Kubernetes Memory -----------------------
def get_mem_usage(pod_name: str, namespace: str) -> float:
    """
    Query server-side memory usage via Kubernetes Metrics API.
    Returns memory in MiB.
    """
    metrics_api: client.CustomObjectsApi = client.CustomObjectsApi()
    metrics: Any = metrics_api.list_namespaced_custom_object(
        group="metrics.k8s.io",
        version="v1beta1",
        namespace=namespace,
        plural="pods",
    )

    for pod in metrics["items"]:
        if pod["metadata"]["name"] == pod_name:
            mem_str: str = pod["containers"][0]["usage"]["memory"]
            if mem_str.endswith("Ki"):
                return int(mem_str[:-2]) / 1024
            elif mem_str.endswith("Mi"):
                return int(mem_str[:-2])
            elif mem_str.endswith("Gi"):
                return int(mem_str[:-2]) * 1024
            else:
                raise ValueError(f"Unexpected memory format: {mem_str}")
    raise RuntimeError(f"Pod {pod_name} not found in namespace {namespace}")

# ------------------------ Payload Load ----------------------------
def load_payload(json_file: str) -> Any:
    with open(json_file, "r") as f:
        return json.load(f)

# ------------------------ Benchmark --------------------------------
def benchmark(url: str, pod_name: str, json_file: str, num_requests: int, delay: float = 0.1) -> None:
    payload: Any = load_payload(json_file)
    wall_times: List[float] = []
    server_mems: List[float] = []

    for i in range(num_requests):
        # measure latency
        start: float = time.perf_counter()
        resp: requests.Response = requests.post(url, json=payload)
        resp.raise_for_status()
        elapsed: float = time.perf_counter() - start
        wall_times.append(elapsed)

        # measure server memory
        mem: float = get_mem_usage(pod_name, NAMESPACE)
        server_mems.append(mem)

        if (i + 1) % 50 == 0:
            print(f"Completed {i+1}/{num_requests} requests")
        time.sleep(delay)  # avoid overwhelming the server

    print("\n--- Results ---")
    print(f"Mean wall time: {statistics.mean(wall_times):.4f} s")
    print(f"p95 wall time: {statistics.quantiles(wall_times, n=100)[94]:.4f} s")
    print(f"Mean server memory: {statistics.mean(server_mems):.2f} MiB")
    print(f"Peak server memory: {max(server_mems):.2f} MiB")

# ------------------------ Main -------------------------------------
def main() -> None:
    if len(sys.argv) < 2:
        print("Usage: python benchmark.py payload.json [num_requests]")
        sys.exit(1)

    json_file: str = sys.argv[1]
    num_requests: int = N
    if len(sys.argv) >= 3:
        num_requests = int(sys.argv[2])

    # Build Docker image
    build_docker_image()

    # Deploy to Kubernetes
    deploy_to_k8s()

    # Load Kubernetes config
    try:
        config.load_kube_config()
    except Exception:
        config.load_incluster_config()

    # Port forward
    pf: subprocess.Popen[Any] = port_forward()
    url: str = f"http://localhost:{LOCAL_PORT}"

    try:
        # Run benchmark
        pod_name = get_pod_name(FUNCTION_NAME, NAMESPACE)
        print(f"Target pod for memory metrics: {pod_name}")
        benchmark(url, pod_name, json_file, num_requests)
    finally:
        print("Stopping port-forward...")
        pf.terminate()
        pf.wait()

if __name__ == "__main__":
    main()
