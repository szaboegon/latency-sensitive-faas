import axios from 'axios';
import type { FunctionApp } from '../models/models';

const FunctionAppService = {
    async fetchFunctionApps(): Promise<FunctionApp[]> {
        const response = await axios.get('/api/function-apps'); // Adjust the endpoint as needed
        return response.data;
    }
};

export default FunctionAppService;