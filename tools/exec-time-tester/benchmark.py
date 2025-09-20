import requests
import time
import statistics
import json
import sys
import subprocess
from kubernetes import client, config
from typing import Any, List, Dict, Tuple
import psutil
from pathlib import Path
import os

# -------------------------------------------------------------------
# Config
# -------------------------------------------------------------------
HANDLERS: List[str] = ["resize", "grayscale", "objectdetect", "cut", "objectdetect2", "tag"]      # handler file names (without .py)
DOCKER_IMAGE_TEMPLATE: str = "szaboegon/{handler}_benchmark:latest"
DEPLOY_YAML_TEMPLATE: str = "./server/deploy.tmpl"
LOCAL_PORT: int = 8080
SVC_PORT: int = 80
NAMESPACE: str = "default"
N: int = 1000  # default number of requests
RESULTS_FILE: Path = Path("./results.json")       # where benchmark results are stored
# -------------------------------------------------------------------


# ------------------------- Docker Build ----------------------------
def build_docker_image(image: str, handler: str) -> None:
    print(f"Building Docker image {image} for handler {handler}...")
    # Use Dockerfile template and format it for the current handler
    dockerfile_template = "./server/Dockerfile.tmpl"
    dockerfile_path = "./server/Dockerfile"
    with open(dockerfile_template, "r") as f:
        dockerfile_content = f.read().format(handler=handler)
    with open(dockerfile_path, "w") as f:
        f.write(dockerfile_content)
    try:
        subprocess.run(
            [
                "docker", "build",
                "-t", image,
                "--build-arg", f"HANDLER={handler}",
                "./server"
            ],
            check=True,
        )
        subprocess.run(["docker", "push", image], check=True)
    finally:
        # Clean up Dockerfile
        if os.path.exists(dockerfile_path):
            os.remove(dockerfile_path)


# ------------------------ Kubernetes Deploy -----------------------
def deploy_to_k8s(deploy_yaml: str, function_name: str) -> None:
    print(f"Deploying {function_name}...")
    subprocess.run(["kubectl", "apply", "-f", deploy_yaml], check=True)
    print(f"Waiting for deployment {function_name} to be ready...")
    subprocess.run(["kubectl", "rollout", "status", f"deployment/{function_name}"], check=True)


def get_pod_name(deployment_name: str, namespace: str) -> str:
    v1 = client.CoreV1Api()
    pods = v1.list_namespaced_pod(
        namespace=namespace,
        label_selector=f"app={deployment_name}"
    )
    if not pods.items:
        raise RuntimeError(f"No pods found for deployment {deployment_name}")
    return str(pods.items[0].metadata.name)


def delete_deployment(function_name: str) -> None:
    print(f"Deleting deployment and service for {function_name}...")
    subprocess.run(["kubectl", "delete", "deployment", function_name], check=False)
    subprocess.run(["kubectl", "delete", "service", function_name], check=False)


# ------------------------ Port Forward ----------------------------
def kill_existing_port_forwards() -> None:
    for proc in psutil.process_iter(["pid", "name", "cmdline"]):
        try:
            cmdline = proc.info["cmdline"]
            if cmdline and "kubectl" in cmdline[0] and "port-forward" in cmdline:
                print(f"Killing existing port-forward PID {proc.info['pid']}: {' '.join(cmdline)}")
                proc.kill()
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            pass


def get_free_port() -> int:
    import socket
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('', 0))
        return int(s.getsockname()[1])


def port_forward(function_name: str, max_retries: int = 5, delay: int = 3) -> Tuple[subprocess.Popen[Any], int]:
    kill_existing_port_forwards()
    local_port = get_free_port()

    for attempt in range(1, max_retries + 1):
        print(f"Port-forward attempt {attempt}/{max_retries} for {function_name} on port {local_port}...")
        pf: subprocess.Popen[Any] = subprocess.Popen(
            ["kubectl", "port-forward", f"svc/{function_name}", f"{local_port}:{SVC_PORT}"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        time.sleep(delay)

        if pf.poll() is None:
            print(f"Port-forward established at localhost:{local_port}")
            return pf, local_port
        else:
            print(f"Port-forward failed, retrying in {delay} seconds...")
            pf.terminate()
            pf.wait()
            time.sleep(delay)
            local_port = get_free_port()

    raise RuntimeError(f"Failed to port-forward {function_name} after {max_retries} attempts")


# ------------------------ Kubernetes Memory -----------------------
def get_mem_usage(pod_name: str, namespace: str, retries: int = 10, delay: float = 1.0) -> float:
    metrics_api: client.CustomObjectsApi = client.CustomObjectsApi()
    for attempt in range(retries):
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
        if attempt < retries - 1:
            time.sleep(delay)
    raise RuntimeError(f"Pod {pod_name} not found in namespace {namespace} after {retries} attempts")


# ------------------------ Payload Load ----------------------------
def load_payload(json_file: str) -> Any:
    with open(json_file, "r") as f:
        return json.load(f)


# ------------------------ Benchmark --------------------------------
def benchmark(url: str, pod_name: str, json_file: str, num_requests: int, delay: float = 0.1) -> Dict[str, Any]:
    payload: Any = load_payload(json_file)
    wall_times: List[float] = []
    server_mems: List[float] = []

    for i in range(num_requests):
        start: float = time.perf_counter()
        try:
            resp: requests.Response = requests.post(url, json=payload)
            resp.raise_for_status()
        except requests.exceptions.RequestException as e:
            print(f"Request failed at iteration {i+1}: {e}")
            continue
        elapsed: float = time.perf_counter() - start
        wall_times.append(elapsed)

        mem: float = get_mem_usage(pod_name, NAMESPACE)
        server_mems.append(mem)

        if (i + 1) % 50 == 0:
            print(f"Completed {i+1}/{num_requests} requests")
        time.sleep(delay)

    if not wall_times:
        raise RuntimeError("No successful requests were completed.")

    result = {
        "mean_wall_time_sec": statistics.mean(wall_times),
        "p95_wall_time_sec": statistics.quantiles(wall_times, n=100)[94] if len(wall_times) >= 100 else max(wall_times),
        "mean_server_memory_mib": statistics.mean(server_mems),
        "peak_server_memory_mib": max(server_mems),
    }

    print("\n--- Results ---")
    for k, v in result.items():
        print(f"{k}: {v}")

    return result


# ------------------------ Results Writer --------------------------
def save_results(handler: str, results: Dict[str, Any]) -> None:
    all_results: Dict[str, Any] = {}
    if RESULTS_FILE.exists():
        with open(RESULTS_FILE, "r") as f:
            try:
                all_results = json.load(f)
            except json.JSONDecodeError:
                all_results = {}

    all_results[handler] = results

    with open(RESULTS_FILE, "w") as f:
        json.dump(all_results, f, indent=2)

    print(f"Results written to {RESULTS_FILE}")


# ------------------------ Main -------------------------------------
def main() -> None:
    if len(sys.argv) < 2:
        print("Usage: python benchmarker.py [num_requests]")
        sys.exit(1)

    num_requests: int = N
    if len(sys.argv) >= 2:
        num_requests = int(sys.argv[1])

    try:
        config.load_kube_config()
    except Exception:
        config.load_incluster_config()

    # Build all images first
    for handler in HANDLERS:
        docker_image = DOCKER_IMAGE_TEMPLATE.format(handler=handler)
        build_docker_image(docker_image, handler)

    # Then deploy and test each handler
    for handler in HANDLERS:
        function_name = f"{handler}-func"
        docker_image = DOCKER_IMAGE_TEMPLATE.format(handler=handler)
        with open(DEPLOY_YAML_TEMPLATE, "r") as f:
            deploy_yaml_content = f.read().format(handler=handler)
        deploy_yaml_path = "./server/deploy.yaml"
        with open(deploy_yaml_path, "w") as f:
            f.write(deploy_yaml_content)

        json_file = f"./inputs/{handler}.json"

        print(f"\n=== Testing handler: {handler} ===")

        try:
            deploy_to_k8s(deploy_yaml_path, function_name)

            pf, local_port = port_forward(function_name)
            url = f"http://localhost:{local_port}"

            try:
                pod_name = get_pod_name(function_name, NAMESPACE)
                print(f"Target pod for memory metrics: {pod_name}")
                results = benchmark(url, pod_name, json_file, num_requests)
                save_results(handler, results)
            finally:
                print("Stopping port-forward...")
                pf.terminate()
                pf.wait()
        finally:
            if os.path.exists(deploy_yaml_path):
                os.remove(deploy_yaml_path)
            # Delete deployment and service after test
            delete_deployment(function_name)


if __name__ == "__main__":
    main()
