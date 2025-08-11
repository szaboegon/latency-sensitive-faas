import React from 'react';
import { useParams } from 'react-router';
import { useFunctionApps } from '../hooks/useFunctionApps';
import type { FunctionComposition, Component } from '../models/models';
import { Container, Typography, List, ListItem, ListItemText, Paper, Box } from '@mui/material';

const FunctionAppDetails: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const { data: functionApps, isLoading, error } = useFunctionApps();

    if (isLoading) {
        return (
            <Container>
                <Typography variant="h6" align="center">
                    Loading...
                </Typography>
            </Container>
        );
    }

    if (error) {
        return (
            <Container>
                <Typography variant="h6" align="center" color="error">
                    Error: {error.message}
                </Typography>
            </Container>
        );
    }

    const app = functionApps?.find((app) => app.id === id);

    if (!app) {
        return (
            <Container>
                <Typography variant="h6" align="center" color="textSecondary">
                    Function App not found.
                </Typography>
            </Container>
        );
    }

    return (
        <Container>
            <Box my={4}>
                <Typography variant="h4" gutterBottom>
                    {app.name} Details
                </Typography>
                <Paper elevation={3} sx={{ padding: 2, marginBottom: 4 }}>
                    <Typography variant="h5" gutterBottom>
                        Components
                    </Typography>
                    <List>
                        {app.components?.map((component: Component) => (
                            <ListItem key={component}>
                                <ListItemText primary={component} />
                            </ListItem>
                        ))}
                    </List>
                </Paper>
                <Paper elevation={3} sx={{ padding: 2 }}>
                    <Typography variant="h5" gutterBottom>
                        Function Compositions
                    </Typography>
                    <List>
                        {(Object.values(app.compositions || {}) as FunctionComposition[]).map((composition) => (
                            <ListItem key={composition.id} sx={{ flexDirection: 'column', alignItems: 'flex-start' }}>
                                <Typography variant="h6">{composition.node}</Typography>
                                <Typography variant="body2">Namespace: {composition.namespace}</Typography>
                                <Typography variant="body2">Runtime: {composition.runtime}</Typography>
                                <Typography variant="body2">Files:</Typography>
                                <List>
                                    {composition.files.map((file) => (
                                        <ListItem key={file}>
                                            <ListItemText primary={file} />
                                        </ListItem>
                                    ))}
                                </List>
                            </ListItem>
                        ))}
                    </List>
                </Paper>
            </Box>
        </Container>
    );
};

export default FunctionAppDetails;
