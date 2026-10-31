import { supermanApi } from '$lib/superman-api.js';

export interface AdminUser {
	id: number;
	name: string;
	email: string;
	type: 'superman';
}

export interface TopbarStat {
	title: string;
	value: string;
	url: string;
	theme: string;
}

interface AdminState {
	user: AdminUser | null;
	loading: boolean;
	initialized: boolean;
	stats: Record<string, TopbarStat>;
}

function createAdminStore() {
	let state = $state<AdminState>({
		user: null,
		loading: false,
		initialized: false,
		stats: {}
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
		get stats() {
			return state.stats;
		},

		async init() {
			try {
				const response = await supermanApi.get<{ data: AdminUser }>('/api/v1/superman/auth/me');
				state.user = response.data;
			} catch {
				supermanApi.clearToken();
				state.user = null;
			} finally {
				state.initialized = true;
			}
		},

		async login(email: string, password: string, remember = true) {
			state.loading = true;
			try {
				const response = await supermanApi.post<{ data: { token: string; user: AdminUser } }>(
					'/api/v1/superman/auth/login',
					{ email, password, remember }
				);

				supermanApi.setToken(response.data.token, remember);
				state.user = response.data.user;

				return { success: true };
			} catch (error: unknown) {
				return {
					success: false,
					message: error instanceof Error ? error.message : 'Unable to sign in.'
				};
			} finally {
				state.loading = false;
			}
		},

		async logout() {
			try {
				await supermanApi.post('/api/v1/superman/auth/logout');
			} finally {
				supermanApi.clearToken();
				state.user = null;
				state.stats = {};
			}
		},

		setStats(stats: Record<string, TopbarStat>) {
			state.stats = stats;
		}
	};
}

export const adminAuth = createAdminStore();
