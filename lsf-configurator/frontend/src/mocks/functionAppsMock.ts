import type { FunctionApp } from "../models/models";

export const functionAppsMock: FunctionApp[] = [
  {
    id: "1",
    name: "Function App 1",
    runtime: "Node.js",
    latencyLimit: 1000,
    files: ["file1.txt", "file2.txt"],
    components: [
      {
        name: "Component A1",
        memory: 128,
        runtime: 100,
        files: ["file1.txt", "file2.txt"],
      },
      { name: "Component B1", memory: 128, runtime: 100, files: ["file1.txt"] },
      { name: "Component C1", memory: 128, runtime: 100, files: ["file2.txt"] },
    ],
    links: [
      { from: "Component A1", to: "Component B1", invocationRate: 1.0 },
      { from: "Component A1", to: "Component C1", invocationRate: 1.0 },
      { from: "Component B1", to: "Component C1", invocationRate: 1.0 },
    ],
    compositions: [
      {
        id: "composition1a",
        functionAppId: "1",
        files: ["file1.txt", "file2.txt"],
        build: {
          image: "image1a",
          timestamp: "2023-01-01T00:00:00Z",
        },
        status: "pending",
        components: ["Component A1", "Component B1", "Component C1"],
        deployments: [
          {
            id: "deployment1a",
            functionCompositionId: "composition1a",
            node: "Node 1",
            namespace: "Namespace 1",
            routingTable: {
              "Component A1": [
                { to: "Component B1", function: "local" },
                { to: "Component C1", function: "deployment1b" },
              ],
              "Component B1": [
                { to: "Component C1", function: "deployment1b" },
              ],
              "Component C1": [],
            },
          },
        ],
      },
      {
        id: "composition1b",
        functionAppId: "1",
        files: ["file3.js", "file4.js"],
        build: {
          image: "image1b",
          timestamp: "2023-01-02T00:00:00Z",
        },
        status: "built",
        components: ["Component C1"],
        deployments: [
          {
            id: "deployment1b",
            functionCompositionId: "composition1b",
            node: "Node 1",
            namespace: "Namespace 1",
            routingTable: {
              "Component C1": [],
            },
          },
        ],
      },
    ],
  },
  {
    id: "2",
    name: "Function App 2",
    runtime: "Python",
    latencyLimit: 1200,
    components: [
      { name: "Component X2", memory: 128, runtime: 100, files: [] },
      { name: "Component Y2", memory: 128, runtime: 100, files: [] },
      { name: "Component Z2", memory: 128, runtime: 100, files: [] },
    ],
    links: [
      { from: "Component X2", to: "Component Y2", invocationRate: 1.0 },
      { from: "Component X2", to: "Component Z2", invocationRate: 1.0 },
      { from: "Component Y2", to: "Component Z2", invocationRate: 1.0 },
    ],
    compositions: [
      {
        id: "composition2a",
        functionAppId: "2",
        files: ["file3.py", "file4.py"],
        build: {
          image: "image2a",
          timestamp: "2023-02-01T00:00:00Z",
        },
        status: "deployed",
        components: ["Component X2", "Component Y2", "Component Z2"],
        deployments: [
          {
            id: "deployment2a",
            functionCompositionId: "composition2a",
            node: "Node 2",
            namespace: "Namespace 2",
            routingTable: {
              "Component X2": [
                { to: "Component Y2", function: "local" },
                { to: "Component Z2", function: "deployment2b" },
              ],
              "Component Y2": [
                { to: "Component Z2", function: "deployment2b" },
              ],
              "Component Z2": [],
            },
          },
        ],
      },
      {
        id: "composition2b",
        functionAppId: "2",
        files: ["file5.py", "file6.py"],
        build: {
          image: "image2b",
          timestamp: "2023-02-02T00:00:00Z",
        },
        status: "error",
        components: ["Component Z2"],
        deployments: [
          {
            id: "deployment2b",
            functionCompositionId: "composition2b",
            node: "Node 2",
            namespace: "Namespace 2",
            routingTable: {
              "Component Z2": [],
            },
          },
        ],
      },
    ],
  },
  {
    id: "3",
    name: "Function App 3",
    runtime: "Go",
    latencyLimit: 900,
    components: [
      { name: "Component M3", memory: 128, runtime: 100, files: [] },
      { name: "Component N3", memory: 128, runtime: 100, files: [] },
      { name: "Component O3", memory: 128, runtime: 100, files: [] },
    ],
    links: [
      { from: "Component M3", to: "Component N3", invocationRate: 1.0 },
      { from: "Component M3", to: "Component O3", invocationRate: 1.0 },
      { from: "Component N3", to: "Component O3", invocationRate: 1.0 },
    ],
    compositions: [
      {
        id: "composition3a",
        functionAppId: "3",
        files: ["file5.go", "file6.go"],
        build: {
          image: "image3a",
          timestamp: "2023-03-01T00:00:00Z",
        },
        status: "pending",
        components: ["Component M3", "Component N3", "Component O3"],
        deployments: [
          {
            id: "deployment3a",
            functionCompositionId: "composition3a",
            node: "Node 3",
            namespace: "Namespace 3",
            routingTable: {
              "Component M3": [
                { to: "Component N3", function: "local" },
                { to: "Component O3", function: "deployment3b" },
              ],
              "Component N3": [
                { to: "Component O3", function: "deployment3b" },
              ],
              "Component O3": [],
            },
          },
        ],
      },
      {
        id: "composition3b",
        functionAppId: "3",
        files: ["file7.go", "file8.go"],
        build: {
          image: "image3b",
          timestamp: "2023-03-02T00:00:00Z",
        },
        status: "built",
        components: ["Component O3"],
        deployments: [
          {
            id: "deployment3b",
            functionCompositionId: "composition3b",
            node: "Node 3",
            namespace: "Namespace 3",
            routingTable: {
              "Component O3": [],
            },
          },
        ],
      },
    ],
  },
  {
    id: "4",
    name: "Function App 4",
    runtime: "Ruby",
    latencyLimit: 1100,
    components: [
      { name: "Component P4", memory: 128, runtime: 100, files: [] },
      { name: "Component Q4", memory: 128, runtime: 100, files: [] },
    ],
    links: [{ from: "Component P4", to: "Component Q4", invocationRate: 1.0 }],
    compositions: [
      {
        id: "composition4a",
        functionAppId: "4",
        files: ["file9.rb", "file10.rb"],
        build: {
          image: "image4a",
          timestamp: "2023-04-01T00:00:00Z",
        },
        status: "pending",
        components: ["Component P4", "Component Q4"],
        deployments: [
          {
            id: "deployment4a",
            functionCompositionId: "composition4a",
            node: "Node 4",
            namespace: "Namespace 4",
            routingTable: {
              "Component P4": [
                { to: "Component Q4", function: "deployment4b" },
              ],
              "Component Q4": [],
            },
          },
        ],
      },
      {
        id: "composition4b",
        functionAppId: "4",
        files: ["file11.rb", "file12.rb"],
        build: {
          image: "image4b",
          timestamp: "2023-04-02T00:00:00Z",
        },
        status: "built",
        components: ["Component Q4"],
        deployments: [
          {
            id: "deployment4b",
            functionCompositionId: "composition4b",
            node: "Node 4",
            namespace: "Namespace 4",
            routingTable: {
              "Component Q4": [],
            },
          },
        ],
      },
    ],
  },
  {
    id: "5",
    name: "Function App 5",
    runtime: "Java",
    latencyLimit: 1500,
    components: [
      { name: "Component R5", memory: 128, runtime: 100, files: [] },
      { name: "Component S5", memory: 128, runtime: 100, files: [] },
      { name: "Component T5", memory: 128, runtime: 100, files: [] },
      { name: "Component U5", memory: 128, runtime: 100, files: [] },
    ],
    links: [
      { from: "Component R5", to: "Component S5", invocationRate: 1.0 },
      { from: "Component S5", to: "Component T5", invocationRate: 1.0 },
      { from: "Component T5", to: "Component U5", invocationRate: 1.0 },
    ],
    compositions: [
      {
        id: "composition5a",
        functionAppId: "5",
        files: ["file15.java", "file16.java"],
        build: {
          image: "image5a",
          timestamp: "2023-05-01T00:00:00Z",
        },
        status: "pending",
        components: ["Component R5", "Component S5"],
        deployments: [
          {
            id: "deployment5a",
            functionCompositionId: "composition5a",
            node: "Node 5",
            namespace: "Namespace 5",
            routingTable: {
              "Component R5": [
                { to: "Component S5", function: "deployment5b" },
              ],
              "Component S5": [],
            },
          },
        ],
      },
      {
        id: "composition5b",
        functionAppId: "5",
        files: ["file17.java", "file18.java"],
        build: {
          image: "image5b",
          timestamp: "2023-05-02T00:00:00Z",
        },
        status: "pending",
        components: ["Component S5", "Component T5"],
        deployments: [
          {
            id: "deployment5b",
            functionCompositionId: "composition5b",
            node: "Node 5",
            namespace: "Namespace 5",
            routingTable: {
              "Component S5": [
                { to: "Component T5", function: "deployment5c" },
              ],
              "Component T5": [],
            },
          },
        ],
      },
      {
        id: "composition5c",
        functionAppId: "5",
        files: ["file19.java", "file20.java"],
        build: {
          image: "image5c",
          timestamp: "2023-05-03T00:00:00Z",
        },
        status: "pending",
        components: ["Component T5", "Component U5"],
        deployments: [
          {
            id: "deployment5c",
            functionCompositionId: "composition5c",
            node: "Node 5",
            namespace: "Namespace 5",
            routingTable: {
              "Component T5": [
                { to: "Component U5", function: "deployment5d" },
              ],
              "Component U5": [],
            },
          },
        ],
      },
      {
        id: "composition5d",
        functionAppId: "5",
        files: ["file21.java", "file22.java"],
        build: {
          image: "image5d",
          timestamp: "2023-05-04T00:00:00Z",
        },
        status: "built",
        components: ["Component U5"],
        deployments: [
          {
            id: "deployment5d",
            functionCompositionId: "composition5d",
            node: "Node 5",
            namespace: "Namespace 5",
            routingTable: {
              "Component U5": [],
            },
          },
        ],
      },
    ],
  },
];
