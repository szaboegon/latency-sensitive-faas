export type Component = string;

export interface FunctionApp {
    id?: string;
    name: string;
    runtime: string;
    sourcePath?: string;
    compositions?: FunctionComposition[];
    components?: Component[];
    files?: string[];
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

export type RoutingTable = Record<Component, Route[]>;

export interface Build {
    image?: string;
    timestamp?: string;
}