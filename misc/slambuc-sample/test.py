import json
import networkx as nx
from slambuc.alg.tree.serial.pseudo import pseudo_ltree_partitioning

with open("tree.json") as f:
    data = json.load(f)

# Build DiGraph
tree = nx.DiGraph()
tree.add_node("P", mem=0, time=0)  # dummy platform node
tree.add_edge("P", 1, rate=1, data=7)   # dummy edge

# Add nodes
for node in data["nodes"]:
    tree.add_node(node["id"], mem=node["mem"], time=node["runtime"])

# Add edges
for edge in data["edges"]:
    u, v, attrs = edge
    tree.add_edge(u, v, rate=attrs["rate"], data=attrs["data"])

# Extract params
params = dict(tree=tree, **data["params"])

# Run algorithm
res = pseudo_ltree_partitioning(**params)

print(f"Part: {res[0]}, opt. cost: {params['M'] * (res[1] / 1000)} GBs, latency: {res[2]} ms")
