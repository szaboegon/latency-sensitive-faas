import type { FunctionApp } from '../models/models';

export const functionAppsMock: FunctionApp[] = [
    {
        id: '1',
        name: 'Function App 1',
        components: ['Component A1', 'Component B1'],
        compositions: [
            {
                id: 'composition1',
                functionAppId: '1',
                node: 'Node 1',
                namespace: 'Namespace 1',
                runtime: 'Node.js',
                files: ['file1.js', 'file2.js'],
                components: {
                    'Component A1': [{ to: 'Component B1', function: 'func1' }],
                },
                build: {
                    image: 'image1',
                    timestamp: '2023-01-01T00:00:00Z',
                },
            },
        ],
    },
    {
        id: '2',
        name: 'Function App 2',
        components: ['Component X2', 'Component Y2'],
        compositions: [
            {
                id: 'composition2',
                functionAppId: '2',
                node: 'Node 2',
                namespace: 'Namespace 2',
                runtime: 'Python',
                files: ['file3.py', 'file4.py'],
                components: {
                    'Component X2': [{ to: 'Component Y2', function: 'func2' }],
                },
                build: {
                    image: 'image2',
                    timestamp: '2023-02-01T00:00:00Z',
                },
            },
        ],
    },
    {
        id: '3',
        name: 'Function App 3',
        components: ['Component M3', 'Component N3', 'Component O3'],
        compositions: [
            {
                id: 'composition3a',
                functionAppId: '3',
                node: 'Node 3',
                namespace: 'Namespace 3',
                runtime: 'Go',
                files: ['file5.go', 'file6.go'],
                components: {
                    'Component M3': [{ to: 'Component N3', function: 'func3a' }],
                },
                build: {
                    image: 'image3a',
                    timestamp: '2023-03-01T00:00:00Z',
                },
            },
            {
                id: 'composition3b',
                functionAppId: '3',
                node: 'Node 3',
                namespace: 'Namespace 3',
                runtime: 'Go',
                files: ['file7.go', 'file8.go'],
                components: {
                    'Component N3': [{ to: 'Component O3', function: 'func3b' }],
                },
                build: {
                    image: 'image3b',
                    timestamp: '2023-03-02T00:00:00Z',
                },
            },
        ],
    },
    {
        id: '4',
        name: 'Function App 4',
        components: ['Component P4', 'Component Q4'],
        compositions: [
            {
                id: 'composition4a',
                functionAppId: '4',
                node: 'Node 4',
                namespace: 'Namespace 4',
                runtime: 'Ruby',
                files: ['file9.rb', 'file10.rb'],
                components: {
                    'Component P4': [{ to: 'Component Q4', function: 'func4a' }],
                },
                build: {
                    image: 'image4a',
                    timestamp: '2023-04-01T00:00:00Z',
                },
            },
            {
                id: 'composition4b',
                functionAppId: '4',
                node: 'Node 4',
                namespace: 'Namespace 4',
                runtime: 'Ruby',
                files: ['file11.rb', 'file12.rb'],
                components: {
                    'Component Q4': [{ to: 'Component P4', function: 'func4b' }],
                },
                build: {
                    image: 'image4b',
                    timestamp: '2023-04-02T00:00:00Z',
                },
            },
            {
                id: 'composition4c',
                functionAppId: '4',
                node: 'Node 4',
                namespace: 'Namespace 4',
                runtime: 'Ruby',
                files: ['file13.rb', 'file14.rb'],
                components: {
                    'Component P4': [{ to: 'Component Q4', function: 'func4c' }],
                },
                build: {
                    image: 'image4c',
                    timestamp: '2023-04-03T00:00:00Z',
                },
            },
        ],
    },
    {
        id: '5',
        name: 'Function App 5',
        components: ['Component R5', 'Component S5', 'Component T5', 'Component U5'],
        compositions: [
            {
                id: 'composition5a',
                functionAppId: '5',
                node: 'Node 5',
                namespace: 'Namespace 5',
                runtime: 'Java',
                files: ['file15.java', 'file16.java'],
                components: {
                    'Component R5': [{ to: 'Component S5', function: 'func5a' }],
                },
                build: {
                    image: 'image5a',
                    timestamp: '2023-05-01T00:00:00Z',
                },
            },
            {
                id: 'composition5b',
                functionAppId: '5',
                node: 'Node 5',
                namespace: 'Namespace 5',
                runtime: 'Java',
                files: ['file17.java', 'file18.java'],
                components: {
                    'Component S5': [{ to: 'Component T5', function: 'func5b' }],
                },
                build: {
                    image: 'image5b',
                    timestamp: '2023-05-02T00:00:00Z',
                },
            },
            {
                id: 'composition5c',
                functionAppId: '5',
                node: 'Node 5',
                namespace: 'Namespace 5',
                runtime: 'Java',
                files: ['file19.java', 'file20.java'],
                components: {
                    'Component T5': [{ to: 'Component U5', function: 'func5c' }],
                },
                build: {
                    image: 'image5c',
                    timestamp: '2023-05-03T00:00:00Z',
                },
            },
            {
                id: 'composition5d',
                functionAppId: '5',
                node: 'Node 5',
                namespace: 'Namespace 5',
                runtime: 'Java',
                files: ['file21.java', 'file22.java'],
                components: {
                    'Component U5': [{ to: 'Component R5', function: 'func5d' }],
                },
                build: {
                    image: 'image5d',
                    timestamp: '2023-05-04T00:00:00Z',
                },
            },
        ],
    },
];
            