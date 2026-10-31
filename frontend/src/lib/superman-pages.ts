export const blankSupermanPages: Record<string, string> = {
	'affiliates/audit': 'Affiliate Audit',
	'affiliates/commission': 'Commission Management',
	'affiliates/dashboard': 'Affiliate Dashboard',
	'affiliates/list': 'Affiliate List',
	'affiliates/reports': 'Affiliate Reports',
	'affiliates/withdrawals': 'Withdrawal Requests',
	'alerts/configuration': 'Alert Configuration',
	'alerts/dashboard': 'Alert Dashboard',
	'alerts/provider-down': 'Provider Down Alerts',
	'api-docs': 'API Documentation',
	'api-management/accounts': 'API Accounts',
	'api-management/keys': 'API Keys',
	'api-management/logs': 'API Logs',
	'api-management/security': 'API Security',
	'approvals/customers': 'Pending Customers',
	'approvals/overview': 'Approval Overview',
	'approvals/resellers': 'Pending Resellers',
	'approvals/senders': 'Pending Senders',
	'audit-logs/employees': 'Employee Activity',
	'audit-logs/global': 'Global Audit Log',
	'audit-logs/security': 'Security Activity',
	'bd-network/dnd': 'DND Management',
	'bd-network/mno': 'MNO Management',
	'bd-network/mnp': 'MNP Management',
	'bd-network/number-validation': 'Number Validation',
	'bd-network/prefix': 'Prefix Management',
	'campaigns/create': 'Create Campaign',
	'campaigns/list': 'Campaign List',
	'campaigns/reports': 'Campaign Reports',
	'dlr/check': 'Delivery Check',
	'dlr/live': 'Live DLR',
	'dlr/logs': 'Delivery Logs',
	'dlr/queue': 'DLR Queue',
	'dlr/retry': 'DLR Retry',
	'employees/list': 'Employee List',
	'employees/sessions': 'Session Management',
	'mo/logs': 'MO Logs',
	'mo/messages': 'MO Messages',
	'mo/reports': 'MO Reports',
	'mo/routing': 'MO Routing',
	'monitoring/api-health': 'API Health',
	'monitoring/live': 'Live Monitoring',
	'monitoring/provider-health': 'Provider Health',
	'monitoring/queue-health': 'Queue Health',
	'monitoring/tps': 'TPS Monitor',
	'payment-gateways/reconciliation': 'Reconciliation',
	'payment-gateways/reports': 'Payment Reports',
	'payment-gateways/transactions': 'Payment Transactions',
	'profit/alerts': 'Profit Alerts',
	'profit/customer': 'Customer Profit',
	'profit/global': 'Global Profit',
	'profit/reseller': 'Reseller Profit',
	'profit/trend': 'Profit Trend',
	'providers/api-logs': 'Provider API Logs',
	'providers/dlr-mapping': 'DLR Mapping',
	'providers/error-mapping': 'Error Mapping',
	'reports/reseller': 'Reseller Reports',
	'resellers/api': 'Reseller API',
	'resellers/audit': 'Reseller Audit',
	'resellers/custom-domain': 'Custom Domain',
	'resellers/customers': 'Reseller Customers',
	'resellers/kyc': 'Reseller KYC',
	'resellers/list': 'Reseller List',
	'resellers/rates': 'Reseller Rates',
	'resellers/routing': 'Reseller Routing',
	'resellers/sender': 'Reseller Sender',
	'resellers/wallet': 'Reseller Wallet',
	'resellers/webhook': 'Reseller Webhook',
	'resellers/white-label': 'White Label',
	'roles-permissions/permissions': 'Permissions',
	'roles-permissions/roles': 'Roles',
	'routing/dashboard': 'Routing Dashboard',
	'routing/health': 'Route Health',
	'routing/load-balancing': 'Load Balancing',
	'routing/routes': 'Route List',
	'routing/testing': 'Route Testing',
	'security/dashboard': 'Security Dashboard',
	'security/fraud-rules': 'Fraud Rules',
	'security/logs': 'Security Logs',
	'sms/all': 'All SMS',
	'sms/bulk': 'Bulk SMS',
	'sms/failed': 'Failed SMS',
	'sms/inquiry': 'SMS Inquiry',
	'sms/logs': 'SMS Logs',
	'sms/message-search': 'Message Search',
	'sms/pending': 'Pending SMS',
	'sms/retry-queue': 'Retry Queue',
	'sms/scheduled': 'Scheduled SMS',
	'sms/send': 'Send SMS',
	'system-logs/all': 'System Logs',
	'system-settings/dlr': 'DLR Settings',
	'system-settings/error-codes': 'Error Code Management',
	'system-settings/maintenance': 'Maintenance Mode',
	'system-settings/notifications': 'Notification Settings',
	'system-settings/queue': 'Queue Settings',
	'webhooks/configuration': 'Webhook Configuration',
	'webhooks/customer': 'Customer Webhooks',
	'webhooks/failed': 'Failed Webhooks',
	'webhooks/logs': 'Webhook Logs',
	'webhooks/reseller': 'Reseller Webhooks'
};

export interface SupermanTablePage {
	title: string;
	description: string;
	endpoint: string;
	columns: Array<{
		key: string;
		label: string;
		format?: 'currency' | 'date' | 'boolean';
	}>;
}

export interface SupermanSettingsField {
	key: string;
	label: string;
	type: 'boolean' | 'number' | 'text' | 'select';
	options?: Array<{ value: string; label: string }>;
}

export interface SupermanSettingsPage {
	title: string;
	description: string;
	group: 'general' | 'sms' | 'security';
	fields: SupermanSettingsField[];
}

const messageColumns: SupermanTablePage['columns'] = [
	{ key: 'customer', label: 'Customer' },
	{ key: 'recipient', label: 'Recipient' },
	{ key: 'body', label: 'Message' },
	{ key: 'status', label: 'Status' },
	{ key: 'cost', label: 'Charged', format: 'currency' },
	{ key: 'operator_cost', label: 'Operator Cost', format: 'currency' },
	{ key: 'primary_dipping_cost', label: 'Gateway Cost', format: 'currency' },
	{ key: 'created_at', label: 'Created', format: 'date' }
];

export const tableSupermanPages: Record<string, SupermanTablePage> = {
	'finance/transactions': {
		title: 'Transactions',
		description: 'Customer account debit and credit transactions.',
		endpoint: '/api/v1/superman/transactions?per_page=100',
		columns: [
			{ key: 'id', label: 'ID' },
			{ key: 'customer', label: 'Customer' },
			{ key: 'reference', label: 'Reference' },
			{ key: 'direction', label: 'Direction' },
			{ key: 'amount', label: 'Amount', format: 'currency' },
			{ key: 'balance_before', label: 'Before', format: 'currency' },
			{ key: 'balance_after', label: 'After', format: 'currency' },
			{ key: 'created_at', label: 'Created', format: 'date' }
		]
	},
	messages: {
		title: 'Messages',
		description: 'All SMS messages submitted through the platform.',
		endpoint: '/api/v1/superman/messages?per_page=100',
		columns: messageColumns
	},
	batches: {
		title: 'Batches',
		description: 'Bulk message processing batches and progress.',
		endpoint: '/api/v1/superman/batches?per_page=100',
		columns: [
			{ key: 'customer', label: 'Customer' },
			{ key: 'sender_id', label: 'Sender ID' },
			{ key: 'source', label: 'Source' },
			{ key: 'status', label: 'Status' },
			{ key: 'approval', label: 'Approval' },
			{ key: 'total_recipients', label: 'Total' },
			{ key: 'remaining', label: 'Remaining' },
			{ key: 'progress_percent', label: 'Progress' },
			{ key: 'created_at', label: 'Created', format: 'date' }
		]
	},
	gateways: {
		title: 'Gateways / Providers',
		description: 'Configured SMS provider routes and their operating state.',
		endpoint: '/api/v1/superman/gateways',
		columns: [
			{ key: 'name', label: 'Gateway' },
			{ key: 'driver', label: 'Driver' },
			{ key: 'priority', label: 'OTP Priority' },
			{ key: 'otp_enabled', label: 'OTP Route', format: 'boolean' },
			{ key: 'enabled', label: 'Status', format: 'boolean' },
			{ key: 'requests_per_second', label: 'RPS' },
			{ key: 'health_status', label: 'Health' }
		]
	},
	'reports/sms': {
		title: 'SMS Reports',
		description: 'Historical SMS delivery and billing report.',
		endpoint: '/api/v1/superman/messages?per_page=100',
		columns: messageColumns
	},
	'sms/all': {
		title: 'All SMS',
		description: 'All messages across every delivery state.',
		endpoint: '/api/v1/superman/messages?per_page=100',
		columns: messageColumns
	},
	'sms/logs': {
		title: 'SMS Logs',
		description: 'Complete record of submitted SMS messages.',
		endpoint: '/api/v1/superman/messages?per_page=100',
		columns: messageColumns
	},
	'sms/pending': {
		title: 'Pending SMS',
		description: 'Messages waiting for SMS dispatch.',
		endpoint: '/api/v1/superman/messages?per_page=100&status=pending',
		columns: messageColumns
	},
	'sms/failed': {
		title: 'Failed SMS',
		description: 'Messages that failed during sending or delivery.',
		endpoint: '/api/v1/superman/messages?per_page=100&status=failed',
		columns: messageColumns
	},
	'dlr/logs': {
		title: 'Delivery Logs',
		description: 'Delivery status records for processed messages.',
		endpoint: '/api/v1/superman/messages?per_page=100',
		columns: messageColumns
	},
	'reports/customer': {
		title: 'Customer Reports',
		description: 'Customer account and balance overview.',
		endpoint: '/api/v1/superman/customers?per_page=100',
		columns: [
			{ key: 'name', label: 'Customer' }, { key: 'email', label: 'Email' }, { key: 'balance', label: 'Balance', format: 'currency' },
			{ key: 'kyc_status', label: 'KYC Status' }, { key: 'created_at', label: 'Joined', format: 'date' }
		]
	},
	'reports/provider': {
		title: 'Provider Reports',
		description: 'Configured SMS gateway operating overview.',
		endpoint: '/api/v1/superman/gateways',
		columns: [
			{ key: 'name', label: 'Provider' }, { key: 'driver', label: 'Driver' }, { key: 'health_status', label: 'Health' },
			{ key: 'enabled', label: 'Enabled', format: 'boolean' }, { key: 'requests_per_second', label: 'RPS' }
		]
	},
	'tariff-plans': {
		title: 'Tariff Plans',
		description: 'Configured billing tariff plans and assigned rates.',
		endpoint: '/api/v1/superman/tariff-plans',
		columns: [
			{ key: 'name', label: 'Plan Name' },
			{ key: 'slug', label: 'Slug' },
			{ key: 'type', label: 'Type' },
			{ key: 'currency', label: 'Currency' },
			{ key: 'rates_count', label: 'Configured Rates' },
			{ key: 'status', label: 'Status' },
			{ key: 'created_at', label: 'Created', format: 'date' }
		]
	},
	rates: {
		title: 'Coverage Rates',
		description: 'SMS pricing rates across operators and tariff plans.',
		endpoint: '/api/v1/superman/rates',
		columns: [
			{ key: 'plan_name', label: 'Tariff Plan' },
			{ key: 'operator_name', label: 'Operator' },
			{ key: 'masking_price', label: 'Masking Rate', format: 'currency' },
			{ key: 'non_masking_price', label: 'Non-Masking Rate', format: 'currency' },
			{ key: 'effective_from', label: 'Effective From', format: 'date' }
		]
	},
	countries: {
		title: 'Countries',
		description: 'Supported destinations and international dial prefixes.',
		endpoint: '/api/v1/superman/countries',
		columns: [
			{ key: 'name', label: 'Country' },
			{ key: 'code', label: 'Dial Code' },
			{ key: 'iso2', label: 'ISO Alpha-2' },
			{ key: 'is_active', label: 'Status', format: 'boolean' }
		]
	},
	operators: {
		title: 'Operators',
		description: 'Mobile Network Operators and routing configurations.',
		endpoint: '/api/v1/superman/operators',
		columns: [
			{ key: 'name', label: 'Operator' },
			{ key: 'country_name', label: 'Country' },
			{ key: 'codes', label: 'Prefix Codes' },
			{ key: 'masking_price', label: 'Base Masking', format: 'currency' },
			{ key: 'non_masking_price', label: 'Base Non-Masking', format: 'currency' },
			{ key: 'is_active', label: 'Active', format: 'boolean' }
		]
	},
	'approvals/customers': {
		title: 'Pending Customers',
		description: 'Client accounts waiting for administrative approval.',
		endpoint: '/api/v1/superman/customers?status=pending',
		columns: [
			{ key: 'name', label: 'Customer' },
			{ key: 'email', label: 'Email' },
			{ key: 'phone', label: 'Phone' },
			{ key: 'business_name', label: 'Company' },
			{ key: 'status', label: 'Status' },
			{ key: 'created_at', label: 'Registered', format: 'date' }
		]
	},
	'approvals/resellers': {
		title: 'Pending Resellers',
		description: 'Reseller applications waiting for review.',
		endpoint: '/api/v1/superman/sps?status=pending',
		columns: [
			{ key: 'name', label: 'Reseller' },
			{ key: 'email', label: 'Email' },
			{ key: 'phone', label: 'Phone' },
			{ key: 'business_name', label: 'Business' },
			{ key: 'status', label: 'Status' },
			{ key: 'created_at', label: 'Applied', format: 'date' }
		]
	},
	'approvals/senders': {
		title: 'Pending Senders',
		description: 'Alphanumeric sender IDs submitted for approval.',
		endpoint: '/api/v1/superman/senders?status=pending',
		columns: [
			{ key: 'sender_id', label: 'Sender ID' },
			{ key: 'type', label: 'Type' },
			{ key: 'requested_by_name', label: 'Requested By' },
			{ key: 'approval_status', label: 'Status' },
			{ key: 'created_at', label: 'Submitted', format: 'date' }
		]
	},
	'sender-cli/masking': {
		title: 'Masking Senders',
		description: 'Approved and registered masking sender IDs.',
		endpoint: '/api/v1/superman/senders?type=masking',
		columns: [
			{ key: 'sender_id', label: 'Sender ID' },
			{ key: 'requested_by_name', label: 'Account' },
			{ key: 'approval_status', label: 'Approval' },
			{ key: 'is_enabled', label: 'Enabled', format: 'boolean' },
			{ key: 'created_at', label: 'Created', format: 'date' }
		]
	},
	'sender-cli/non-masking': {
		title: 'Non-Masking Senders',
		description: 'Registered numeric and non-masking sender routes.',
		endpoint: '/api/v1/superman/senders?type=non_masking',
		columns: [
			{ key: 'sender_id', label: 'Sender ID' },
			{ key: 'requested_by_name', label: 'Account' },
			{ key: 'approval_status', label: 'Approval' },
			{ key: 'is_enabled', label: 'Enabled', format: 'boolean' },
			{ key: 'created_at', label: 'Created', format: 'date' }
		]
	},
	'resellers/list': {
		title: 'Reseller List',
		description: 'All registered service providers and credit accounts.',
		endpoint: '/api/v1/superman/sps',
		columns: [
			{ key: 'name', label: 'Reseller' },
			{ key: 'email', label: 'Email' },
			{ key: 'phone', label: 'Phone' },
			{ key: 'business_name', label: 'Business' },
			{ key: 'balance', label: 'Balance', format: 'currency' },
			{ key: 'customer_count', label: 'Clients' },
			{ key: 'status', label: 'Status' }
		]
	}
};

export const settingsSupermanPages: Record<string, SupermanSettingsPage> = {
	'settings/general': {
		title: 'Company Info',
		description: 'Manage the platform name and logo file path.',
		group: 'general',
		fields: [
			{ key: 'appName', label: 'App Name', type: 'text' },
			{ key: 'logoUrl', label: 'Logo File Path', type: 'text' }
		]
	},
	'system-settings/sms': {
		title: 'SMS Technical Settings',
		description: 'Configure global SMS dispatch, encoding, limits, validation, and message logging.',
		group: 'sms',
		fields: [
			{ key: 'sms_enabled', label: 'SMS Sending Enabled', type: 'boolean' },
			{
				key: 'default_sms_type', label: 'Default SMS Type', type: 'select', options: [
					{ value: 'plain_text', label: 'Plain Text' }, { value: 'masking', label: 'Masking SMS' },
					{ value: 'non_masking', label: 'Non-Masking SMS' }, { value: 'transactional', label: 'Transactional' },
					{ value: 'promotional', label: 'Promotional' }
				]
			},
			{
				key: 'default_encoding', label: 'Default Character Encoding', type: 'select', options: [
					{ value: 'GSM-7', label: 'GSM 7-bit' }, { value: 'UCS-2', label: 'UCS-2 / Unicode' }, { value: 'UTF-8', label: 'UTF-8 Automatic' }
				]
			},
			{ key: 'auto_detect_unicode', label: 'Auto-Detect Unicode', type: 'boolean' },
			{ key: 'force_unicode', label: 'Force Unicode for All Messages', type: 'boolean' },
			{ key: 'allow_long_sms', label: 'Allow Multi-Part SMS', type: 'boolean' },
			{ key: 'max_sms_parts', label: 'Maximum SMS Parts', type: 'number' },
			{ key: 'duplicate_protection_enabled', label: 'Enable Duplicate SMS Protection', type: 'boolean' },
			{ key: 'duplicate_window_seconds', label: 'Deduplication Window (Seconds)', type: 'number' },
			{ key: 'default_country_code', label: 'Default Country Code', type: 'text' },
			{ key: 'log_payloads', label: 'Record Request and Response Payloads', type: 'boolean' },
			{ key: 'log_retention_days', label: 'Log Retention (Days)', type: 'number' },
			{ key: 'message_id_prefix', label: 'Message ID Prefix', type: 'text' }
		]
	},
	'security/ip-whitelist': {
		title: 'API IP Whitelist & Security Policy',
		description: 'Configure direct API IP-whitelist enforcement and portal bypass controls.',
		group: 'security',
		fields: [
			{ key: 'ip_whitelist_enforced', label: 'Enforce IP Whitelist for Direct API', type: 'boolean' },
			{ key: 'allow_all_when_whitelist_empty', label: 'Allow All IPs When Whitelist Is Empty', type: 'boolean' },
			{ key: 'auto_approve_api_requests', label: 'Auto-Approve IP Whitelists Submitted via API', type: 'boolean' },
			{ key: 'web_portal_bypass_enabled', label: 'Bypass IP Whitelist for Web Portal Backend', type: 'boolean' },
			{ key: 'trusted_portal_ips', label: 'Trusted Web Portal Server IPs', type: 'text' },
			{ key: 'blocked_message', label: 'Access Denied Error Message', type: 'text' }
		]
	}
};
