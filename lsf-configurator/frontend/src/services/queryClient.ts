import { Mutation, MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import {toast} from "react-toastify";

const handleError = (
    error: Error,
    _variables?: unknown,
    _context?: unknown,
    mutation?: Mutation<unknown, unknown, unknown, unknown>
) => {
    if (!(error instanceof AxiosError) || !error.response) {
        toast.error("An unknown error occurred");
        return;
    }

    const message = mutation?.meta?.errorMessage;
    if (message) {
        toast.error(message as string);
        return;
    }

    if (error.response.status === 401 || error.response.status === 403) {
        toast.error("You are not authorized to perform this action");
        return;
    }

    if (String(error.response.status).startsWith("5")) {
        toast.error("An internal server error occurred");
        return;
    }
};

const queryClient = new QueryClient({
    queryCache: new QueryCache({
        onError: (error) => handleError(error),
    }),
    mutationCache: new MutationCache({
        onSuccess: (_data, _variables, _context, mutation) => {
            const message = mutation.meta?.successMessage;
            if (message) {
                toast.success(message as string);
            }
        },
        onError: handleError,
    }),
});

export default queryClient;