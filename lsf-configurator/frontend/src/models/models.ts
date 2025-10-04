export interface Component {
  name: string;
  memory: number;
  runtime: number;
  files: string[];
}

export interface ComponentLink {
  from: string;
  to: string;
  invocationRate: number;
  dataDelay: number;
}

export interface FunctionApp {
  id?: string;
  name: string;
  runtime: string;
  components?: Component[];
  links?: ComponentLink[];
  files?: string[];
  compositions?: FunctionComposition[];
  sourcePath?: string;
  latencyLimit: number;
}

export interface FunctionComposition {
  id?: string;
  functionAppId?: string;
  components: string[];
  files: string[];
  build: Build;
  status: "pending" | "built" | "deployed" | "error";
  deployments: Deployment[];
}

export interface Deployment {
  id?: string;
  functionCompositionId?: string;
  namespace: string;
  node: string;
  routingTable: RoutingTable;
}

export interface Route {
  to: string;
  function: string;
}

export type RoutingTable = Record<string, Route[]>;

export interface Build {
  image?: string;
  timestamp?: string;
}
