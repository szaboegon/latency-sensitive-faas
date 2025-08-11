import React from 'react';
import { useFunctionApps } from '../hooks/useFunctionApps';
import { useNavigate } from 'react-router';
import type { FunctionApp } from '../models/models';
import { Container, Typography, List, ListItem, ListItemText, Button, Paper, Box } from '@mui/material';

const Home: React.FC = () => {
    const { data: functionApps, isLoading, error } = useFunctionApps();
    const navigate = useNavigate();

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

    return (
        <Container>
            <Box my={4}>
                <Typography variant="h4" gutterBottom>
                    Function Apps Overview
                </Typography>
                {functionApps?.length === 0 ? (
                    <Typography variant="body1" color="textSecondary">
                        No function apps found.
                    </Typography>
                ) : (
                    <Paper elevation={3} sx={{ padding: 2 }}>
                        <List>
                            {functionApps?.map((app: FunctionApp) => (
                                <ListItem key={app.id} sx={{ flexDirection: 'column', alignItems: 'flex-start' }}>
                                    <Typography variant="h5">{app.name}</Typography>
                                    <Typography variant="body2">Components:</Typography>
                                    <List>
                                        {app.components?.map((component) => (
                                            <ListItem key={component}>
                                                <ListItemText primary={component} />
                                            </ListItem>
                                        ))}
                                    </List>
                                    <Button
                                        variant="contained"
                                        color="primary"
                                        onClick={() => navigate(`/function-apps/${app.id}`)}
                                        sx={{ marginTop: 2 }}
                                    >
                                        View Details
                                    </Button>
                                </ListItem>
                            ))}
                        </List>
                    </Paper>
                )}
            </Box>
        </Container>
    );
};

export default Home;
