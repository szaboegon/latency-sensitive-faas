import requests
import time
import statistics
import json
import sys
import subprocess
from kubernetes import client, config
from typing import Any, List, Dict
import psutil
from pathlib import Path

# -------------------------------------------------------------------
# Config
# -------------------------------------------------------------------
HANDLERS: List[str] = ["main", "foo", "bar"]      # handler file names (without .py)
DOCKER_IMAGE_TEMPLATE: str = "szaboegon/{handler}:latest"
DEPLOY_YAML_TEMPLATE: str = "./server/{handler}_deploy.yaml"
LOCAL_PORT: int = 8080
SVC_PORT: int = 80
NAMESPACE: str = "default"
N: int = 1000  # default number of requests
RESULTS_FILE: Path = Path("./results.json")       # where benchmark results are stored
# -------------------------------------------------------------------


# ------------------------- Docker Build ----------------------------
def build_docker_image(image: str, handler: str) -> None:
    print(f"Building Docker image {image} for handler {handler}...")
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


def port_forward(function_name: str, max_retries: int = 5, delay: int = 3) -> subprocess.Popen[Any]:
    kill_existing_port_forwards()

    for attempt in range(1, max_retries + 1):
        print(f"Port-forward attempt {attempt}/{max_retries} for {function_name}...")
        pf: subprocess.Popen[Any] = subprocess.Popen(
            ["kubectl", "port-forward", f"svc/{function_name}", f"{LOCAL_PORT}:{SVC_PORT}"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        time.sleep(delay)

        if pf.poll() is None:
            print(f"Port-forward established at localhost:{LOCAL_PORT}")
            return pf
        else:
            print(f"Port-forward failed, retrying in {delay} seconds...")
            pf.terminate()
            pf.wait()
            time.sleep(delay)

    raise RuntimeError(f"Failed to port-forward {function_name} after {max_retries} attempts")


# ------------------------ Kubernetes Memory -----------------------
def get_mem_usage(pod_name: str, namespace: str) -> float:
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
def benchmark(url: str, pod_name: str, json_file: str, num_requests: int, delay: float = 0.1) -> Dict[str, Any]:
    payload: Any = load_payload(json_file)
    wall_times: List[float] = []
    server_mems: List[float] = []

    for i in range(num_requests):
        start: float = time.perf_counter()
        resp: requests.Response = requests.post(url, json=payload)
        resp.raise_for_status()
        elapsed: float = time.perf_counter() - start
        wall_times.append(elapsed)

        mem: float = get_mem_usage(pod_name, NAMESPACE)
        server_mems.append(mem)

        if (i + 1) % 50 == 0:
            print(f"Completed {i+1}/{num_requests} requests")
        time.sleep(delay)

    result = {
        "mean_wall_time_sec": statistics.mean(wall_times),
        "p95_wall_time_sec": statistics.quantiles(wall_times, n=100)[94],
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
        print("Usage: python benchmarker.py payload.json [num_requests]")
        sys.exit(1)

    json_file: str = sys.argv[1]
    num_requests: int = N
    if len(sys.argv) >= 3:
        num_requests = int(sys.argv[2])

    try:
        config.load_kube_config()
    except Exception:
        config.load_incluster_config()

    for handler in HANDLERS:
        function_name = f"{handler}-func"
        docker_image = DOCKER_IMAGE_TEMPLATE.format(handler=handler)
        deploy_yaml = DEPLOY_YAML_TEMPLATE.format(handler=handler)

        print(f"\n=== Testing handler: {handler} ===")

        build_docker_image(docker_image, handler)
        deploy_to_k8s(deploy_yaml, function_name)

        pf = port_forward(function_name)
        url = f"http://localhost:{LOCAL_PORT}"

        try:
            pod_name = get_pod_name(function_name, NAMESPACE)
            print(f"Target pod for memory metrics: {pod_name}")
            results = benchmark(url, pod_name, json_file, num_requests)
            save_results(handler, results)
        finally:
            print("Stopping port-forward...")
            pf.terminate()
            pf.wait()


if __name__ == "__main__":
    main()
