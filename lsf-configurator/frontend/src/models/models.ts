export interface Component {
  name: string;
  memory: number;
  runtime: number;
  files: string[];
}

export interface ComponentLink {
  from: string;
  to: string;
  invocationRate: InvocationRate;
  dataDelay: number;
}

export interface InvocationRate {
  min: number;
  max: number;
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
  layoutCandidates?: Record<string, Layout>; // key: strategy name
  activeLayoutKey?: string;
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

export interface ComponentProfile {
  name: string;
  runtime: number;
  memory: number;
  requiredReplicas: number;
}

export interface CompositionInfo {
  componentProfiles: ComponentProfile[];
  requiredReplicas: number;
  totalEffectiveMemory: number;
  totalMCPU: number;
  targetConcurrency: number;
}

export type Layout = Record<string, CompositionInfo>; // key: node name

export interface AppResult {
  timestamp: string;
  event: unknown; // event can be any JSON value
}
