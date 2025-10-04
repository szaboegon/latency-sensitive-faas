#!/usr/bin/env python3
import sys
import json
import networkx as nx
from slambuc.alg.tree.serial.pseudo import pseudo_ltree_partitioning  # type: ignore
from typing import Any


def main() -> None:
    # Read JSON input from stdin
    data = json.load(sys.stdin)

    # Extract parameters, nodes, edges
    params = data.get("params", {})
    nodes = data.get("nodes", [])
    edges = data.get("edges", [])

    # Create directed graph
    tree: nx.DiGraph[Any] = nx.DiGraph()

    # Add a dummy root/platform node
    tree.add_node("P", mem=0, time=0)
    tree.add_edge("P", params.get("root", 1), rate=1, data=0)  # dummy edge

    # Add nodes dynamically
    for node in nodes:
        node_id = node["id"]
        tree.add_node(node_id, mem=node["mem"], time=node["runtime"])

    # Add edges dynamically
    for edge in edges:
        u = edge["from"]
        v = edge["to"]
        attrs = edge.get("attr", {})
        tree.add_edge(u, v, rate=attrs.get("rate", 1), data=attrs.get("data", 0))

    # Combine tree and params for algorithm
    algo_params = dict(tree=tree, **params)

    # Run SLAMBUC layout algorithm
    res = pseudo_ltree_partitioning(**algo_params)

    # Return JSON output
    output = {
        "layout": res[0],
        "opt_cost": (
            float(res[1]) if res[1] is not None and res[1] != float("inf") else -1
        ),  # in GB
        "latency": res[2] if res[2] is not None else -1,  # in ms
    }
    print(json.dumps(output))


if __name__ == "__main__":
    main()
