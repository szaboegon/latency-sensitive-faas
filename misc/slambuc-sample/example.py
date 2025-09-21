from slambuc.alg.tree.serial.pseudo import pseudo_ltree_partitioning
from slambuc.misc.random import get_random_tree
from networkx import DiGraph

# Get input parameters
tree: DiGraph = get_random_tree(nodes=3)  # Assuming random memory demands are in GB
params = dict(tree=tree,
              root=1,  # Root node ID
              M=6,  # Memory upper limit
              L=450,  # Latency upper limit
              cp_end=10,  # Critical path: [root -> cp_end]
              delay=10  # Platform delay in ms
              )

# Partitioning
res = pseudo_ltree_partitioning(**params)
print(f"Part: {res[0]}, opt. cost: {params['M'] * (res[1] / 1000)} GBs, latency: {res[2]} ms")
"Part: [[1, 2], [3, 4], [5, 6, 7, 8], [9], [10]], opt. cost: 7.512 GBs, latency: 449 ms"