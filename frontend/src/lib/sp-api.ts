const baseURL = import.meta.env.VITE_SP_API_URL ?? '';

const storageKey = 'sms_sp_access_token';
const sessionStorageKey = 'sms_sp_session_token';

export class SpApiError extends Error {
	constructor(
		public status: number,
		public data: Record<string, unknown>
	) {
		super((data.message as string) ?? `HTTP ${status}`);
	}
}

function storedToken(): string | null {
	if (typeof window === 'undefined') {
		return null;
	}

	return localStorage.getItem(storageKey) ?? sessionStorage.getItem(sessionStorageKey);
}

function authorizationHeader(): Record<string, string> {
	const token = storedToken();

	return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const response = await fetch(`${baseURL}${path}`, {
		method,
		headers: {
			Accept: 'application/json',
			...authorizationHeader(),
			...(body === undefined ? {} : { 'Content-Type': 'application/json' })
		},
		body: body === undefined ? undefined : JSON.stringify(body)
	});

	if (!response.ok) {
		let data: Record<string, unknown> = {};
		try {
			data = await response.json();
		} catch {
			// A reverse proxy may return a non-JSON error before the API is available.
		}

		throw new SpApiError(response.status, data);
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return response.json() as Promise<T>;
}

export const spApi = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
	put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
	patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
	delete: <T>(path: string) => request<T>('DELETE', path),
	getToken(): string | null {
		return storedToken();
	},
	setToken(token: string, remember: boolean = true): void {
		if (typeof window === 'undefined') {
			return;
		}

		this.clearToken();
		if (remember) {
			localStorage.setItem(storageKey, token);
		} else {
			sessionStorage.setItem(sessionStorageKey, token);
		}
	},
	clearToken(): void {
		if (typeof window === 'undefined') {
			return;
		}

		localStorage.removeItem(storageKey);
		sessionStorage.removeItem(sessionStorageKey);
	}
};
