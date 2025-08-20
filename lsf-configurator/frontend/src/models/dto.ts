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
}

export interface FunctionCompositionBulkCreateDto {
    components: string[];
    files: string[];
    deployments: DeploymentCreateDto[];
}

export interface DeploymentBulkCreateDto {
    node: string;
    namespace: string;
    routingTable: Record<string, Array<{ to: string; function: string }>>;
}
