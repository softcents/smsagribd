<svelte:head>
	<title>Email Logs — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">Email & SMTP Notification Logs</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Audit log for transactional mail delivery, 2FA code emails, and admin alert dispatches.
			</p>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="py-16 text-center text-xs text-gray-500 dark:text-gray-400">
			No outbound email errors recorded. All mailer queues operational.
		</div>
	</div>
</div>

