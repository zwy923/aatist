import { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/client';

/**
 * Hook for paginated queries
 * @param {Object} options
 * @param {string} options.endpoint - API endpoint
 * @param {Object} options.params - Additional query params
 * @param {number} options.pageSize - Items per page
 */
export const usePaginatedQuery = ({ endpoint, params = {}, pageSize = 10 }) => {
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [hasNext, setHasNext] = useState(false);

    const fetchData = useCallback(async (currentPage) => {
        setLoading(true);
        setError(null);
        try {
            const response = await apiClient.get(endpoint, {
                params: {
                    ...params,
                    page: currentPage,
                    pageSize,
                },
            });

            const { items, total: totalItems, page: respPage, hasNext: respHasNext } = response.data.data;

            setData(items || []);
            setTotal(totalItems || 0);
            setHasNext(respHasNext || false);
        } catch (err) {
            setError(err.message || 'Failed to fetch data');
            console.error(`Error fetching ${endpoint}:`, err);
        } finally {
            setLoading(false);
        }
    }, [endpoint, JSON.stringify(params), pageSize]);

    useEffect(() => {
        fetchData(page);
    }, [page, fetchData]);

    const refresh = () => fetchData(page);

    return {
        data,
        loading,
        error,
        page,
        total,
        hasNext,
        setPage,
        refresh,
    };
};

export default usePaginatedQuery;
