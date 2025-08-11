import type { FunctionApp } from '../models/models';

export const functionAppsMock: FunctionApp[] = [
    {
        id: '1',
        name: 'Function App 1',
        components: ['Component A', 'Component B'],
        compositions: {
            'composition1': {
                id: 'composition1',
                functionAppId: '1',
                node: 'Node 1',
                namespace: 'Namespace 1',
                runtime: 'Node.js',
                files: ['file1.js', 'file2.js'],
                components: {
                    'Component A': [{ to: 'Component B', function: 'func1' }],
                },
                build: {
                    image: 'image1',
                    timestamp: '2023-01-01T00:00:00Z',
                },
            },
        },
    },
    {
        id: '2',
        name: 'Function App 2',
        components: ['Component X', 'Component Y'],
        compositions: {
            'composition2': {
                id: 'composition2',
                functionAppId: '2',
                node: 'Node 2',
                namespace: 'Namespace 2',
                runtime: 'Python',
                files: ['file3.py', 'file4.py'],
                components: {
                    'Component X': [{ to: 'Component Y', function: 'func2' }],
                },
                build: {
                    image: 'image2',
                    timestamp: '2023-02-01T00:00:00Z',
                },
            },
        },
    },
];
