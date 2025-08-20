export interface FunctionAppCreateDto {
    name: string;
    runtime: string;
}

export interface FunctionCompositionCreateDto {
    functionAppId: string;
    components: string[];
    files: string[];
}

export interface DeploymentCreateDto {
    functionCompositionId: string;
    node: string;
    namespace: string;
    routingTable: Record<string, Array<{ to: string; function: string }>>;
}

export interface BulkCreateRequest {
    functionApp: FunctionAppCreateDto;
    functionCompositions: FunctionCompositionBulkCreateDto[];
    deployments: DeploymentBulkCreateDto[];
}

export interface FunctionCompositionBulkCreateDto {
    id: string; // here ID is a temporary ID, given by the user for finding matching objects in bulk requests
    components: string[];
    files: string[];
}

export interface DeploymentBulkCreateDto {
    id: string; // here ID is a temporary ID, given by the user for finding matching objects in bulk requests
    functionCompositionId: string;
    node: string;
    namespace: string;
    routingTable: Record<string, Array<{ to: string; function: string }>>;
}
