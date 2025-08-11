export type Component = string;

export interface FunctionApp {
    id?: string;
    name: string;
    compositions?: FunctionComposition[];
    components?: Component[];
    files?: string[];
}

export interface FunctionComposition {
    id?: string;
    functionAppId?: string;
    node?: string;
    components: RoutingTable;
    namespace: string;
    sourcePath?: string;
    runtime: string;
    files: string[];
    build: Build;
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
