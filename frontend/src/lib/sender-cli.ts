import { supermanApi } from '$lib/superman-api.js';

export interface SenderUserRef {
	id: number;
	name: string;
	email: string | null;
}

export interface SenderRecord {
	id: number;
	sender_id: string;
	type: string;
	masking_category: string;
	is_enabled: boolean;
	approval_status: string;
	rejection_reason: string | null;
	gateway_id: number | null;
	gateway_name: string | null;
	requested_by_id: number | null;
	requested_by_name: string | null;
	reviewed_at: string | null;
	created_at: string;
	updated_at: string | null;
	users: SenderUserRef[];
	users_count: number;
	messages_count: number;
}

export interface PageMeta {
	page: number;
	per_page: number;
	total: number;
	total_pages: number;
}

export interface SenderListResponse {
	data: SenderRecord[];
	meta: PageMeta;
}

export interface GatewayOption {
	id: number;
	name: string;
}

export interface SenderOptionRecord {
	id: number;
	sender_id: string;
	type: string;
	masking_category: string;
}

export interface SenderOptions {
	gateways: GatewayOption[];
	customers: SenderUserRef[];
	senders: SenderOptionRecord[];
}

const OPTIONS_CACHE = new Map<string, SenderOptions>();

export function loadSenderOptions(category = ''): Promise<SenderOptions> {
	const key = category;
	const cached = OPTIONS_CACHE.get(key);
	if (cached) {
		return Promise.resolve(cached);
	}

	const query = category ? `?type=${encodeURIComponent(category)}` : '';

	return supermanApi
		.get<{ data: SenderOptions }>(`/api/v1/superman/sender-cli/options${query}`)
		.then((response) => {
			OPTIONS_CACHE.set(key, response.data);

			return response.data;
		});
}

export function invalidateSenderOptions(): void {
	OPTIONS_CACHE.clear();
}

export interface SenderListQuery {
	page: number;
	perPage: number;
	search?: string;
	category?: string;
	status?: string;
	gatewayId?: number;
	assignment?: string;
}

export function loadSenders(query: SenderListQuery): Promise<SenderListResponse> {
	const params = new URLSearchParams({
		page: String(query.page),
		per_page: String(query.perPage)
	});
	if (query.search) params.set('search', query.search);
	if (query.category) params.set('type', query.category);
	if (query.status) params.set('status', query.status);
	if (query.gatewayId) params.set('gateway_id', String(query.gatewayId));
	if (query.assignment) params.set('assignment', query.assignment);

	return supermanApi.get<SenderListResponse>(`/api/v1/superman/senders?${params.toString()}`);
}

export function typeBadgeClass(category: string): string {
	return category === 'masking'
		? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300'
		: 'border-sky-200 bg-sky-50 text-sky-700 dark:border-sky-800 dark:bg-sky-950/40 dark:text-sky-300';
}

export function typeLabel(category: string): string {
	return category === 'masking' ? 'Masking' : 'Non-masking';
}

export function approvalBadgeClass(status: string): string {
	if (status === 'approved') {
		return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300';
	}
	if (status === 'rejected') {
		return 'border-rose-200 bg-rose-50 text-rose-700 dark:border-rose-800 dark:bg-rose-950/40 dark:text-rose-300';
	}

	return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300';
}

export function formatDateTime(value: string | null | undefined): string {
	if (!value) {
		return '—';
	}
	const date = new Date(value);
	if (isNaN(date.getTime())) {
		return value;
	}

	return (
		date.toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' }) +
		' ' +
		date.toTimeString().slice(0, 8)
	);
}

export function formatDate(value: string | null | undefined): string {
	if (!value) {
		return '—';
	}
	const date = new Date(value);
	if (isNaN(date.getTime())) {
		return value;
	}

	return date.toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' });
}

export function currency(value: number | null | undefined): string {
	return `৳${Number(value ?? 0).toFixed(4)}`;
}
