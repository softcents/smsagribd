/**
 * S P Reseller portal state store using Svelte 5 runes.
 * Backed by real Go Fiber API (/api/v1/sp).
 */

import { spApi, SpApiError } from '$lib/sp-api.js';

export interface SpUser {
	id: number;
	name: string;
	email: string;
	phone?: string;
	balance: number;
	type: string;
	status: string;
}

interface SpState {
	user: SpUser | null;
	loading: boolean;
	initialized: boolean;
}

function createSpStore() {
	let state = $state<SpState>({
		user: null,
		loading: false,
		initialized: false
	});

	return {
		get user() {
			return state.user;
		},
		get loading() {
			return state.loading;
		},
		get initialized() {
			return state.initialized;
		},
		get isLoggedIn() {
			return state.user !== null;
		},
		get balance() {
			return state.user?.balance ?? 0;
		},

		async init() {
			if (state.initialized) return;
			state.loading = true;

			const token = spApi.getToken();
			if (!token) {
				state.user = null;
				state.initialized = true;
				state.loading = false;
				return;
			}

			try {
				const res = await spApi.get<{ data: SpUser }>('/api/v1/sp/auth/me');
				state.user = res.data;
			} catch (e) {
				if (e instanceof SpApiError && e.status === 401) {
					spApi.clearToken();
					state.user = null;
				}
			} finally {
				state.loading = false;
				state.initialized = true;
			}
		},

		async login(email: string, password: string, remember: boolean = true) {
			state.loading = true;
			try {
				const res = await spApi.post<{ data: { token: string; user: SpUser } }>('/api/v1/sp/auth/login', {
					email,
					password,
					remember
				});

				if (res?.data?.token) {
					spApi.setToken(res.data.token, remember);
				}

				if (res?.data?.user) {
					state.user = res.data.user;
				} else {
					const meRes = await spApi.get<{ data: SpUser }>('/api/v1/sp/auth/me');
					state.user = meRes.data;
				}

				return { success: true };
			} catch (e) {
				spApi.clearToken();
				state.user = null;
				if (e instanceof SpApiError) {
					const message =
						(e.data?.message as string) ??
						((e.data?.errors as Record<string, string[]>)?.email?.[0] as string) ??
						'Invalid Reseller credentials.';
					return { success: false, message };
				}
				return { success: false, message: 'Network error. Please try again.' };
			} finally {
				state.loading = false;
			}
		},

		async logout() {
			try {
				await spApi.post('/api/v1/sp/auth/logout');
			} catch {
				// ignore
			} finally {
				spApi.clearToken();
				state.user = null;
				state.initialized = false;
			}
		},

		async refreshUser() {
			try {
				const res = await spApi.get<{ data: SpUser }>('/api/v1/sp/auth/me');
				state.user = res.data;
			} catch {
				// ignore
			}
		},

		updateBalance(newBalance: number) {
			if (state.user) {
				state.user = { ...state.user, balance: newBalance };
			}
		}
	};
}

export const spAuth = createSpStore();
