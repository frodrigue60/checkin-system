import { env } from '$env/dynamic/public';
import { authState } from './auth.svelte';

const BASE_URL = env.PUBLIC_API_URL;

export interface TypedResponse<T> extends Response {
	json(): Promise<T>;
}

/**
 * Standardized API Fetch with Type Safety and Error Auditing
 */
export async function apiFetch<T = any>(path: string, config: RequestInit = {}): Promise<TypedResponse<T>> {
	const url = `${BASE_URL}${path}`;
	const headers = new Headers(config.headers || {});
	
	if (!headers.has('Content-Type') && !(config.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}

	if (authState.token) {
		headers.set('Authorization', `Bearer ${authState.token}`);
	}

	try {
		const response = (await fetch(url, { ...config, headers })) as TypedResponse<T>;

		// Handle Unauthorized
		if (response.status === 401) {
			console.warn('🚨 Unauthorized request. Redirecting to login...');
			authState.logout();
			if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
				window.location.href = '/login';
			}
		}

		// Centralized Audit/Logging for Frontend Errors
		if (!response.ok) {
			const errorClone = response.clone();
			errorClone.json().then((errData) => {
				console.group(`🚨 API Error [${response.status}]: ${path}`);
				console.error('Payload:', errData);
				console.error('Config:', config);
				console.groupEnd();
			}).catch(() => {
				console.error(`🚨 API Error [${response.status}] (No JSON): ${path}`);
			});
		}

		return response;
	} catch (error) {
		console.error(`🚨 Network/Fetch Fatal Error: ${path}`, error);
		throw error;
	}
}
