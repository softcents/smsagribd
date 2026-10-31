<svelte:head>
	<title>System Error Logs — Superman Administrator</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-gray-950 dark:text-white">System Error Logs</h1>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
				Captured application exceptions, fatal warnings, and backend panic recovery events.
			</p>
		</div>
	</div>

	<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xs dark:border-gray-800 dark:bg-gray-900">
		<div class="py-16 text-center text-xs text-gray-500 dark:text-gray-400">
			<svg class="mx-auto h-8 w-8 text-emerald-500 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			No active critical errors logged. All Go Fiber services operating normally.
		</div>
	</div>
</div>

