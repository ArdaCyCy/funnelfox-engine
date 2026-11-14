<script lang="ts">
  import FunnelPlayer from '$lib/components/FunnelPlayer.svelte';
  import type { PageData } from './$types';
  let { data }: { data: PageData } = $props();
</script>

<svelte:head>
  <title>{data.funnel.content.hook.headline} | {data.funnel.productName}</title>
</svelte:head>

<main class="flex min-h-screen flex-col items-center justify-center bg-slate-100 px-4 py-8">
  <div class="mb-3 flex items-center gap-2 text-xs font-semibold text-slate-400">
    <span>Varyant {data.funnel.variant}</span>
    <span>•</span>
    <span>Yanıt Süresi: {(data.funnel.latencyUs / 1000).toFixed(2)} ms</span>
    <span>•</span>
    <span>Kaynak: {data.utmSource}</span>
  </div>
  <FunnelPlayer
    slug={data.funnel.slug}
    variantKey={data.funnel.variant}
    variant={data.funnel.content}
    utmSource={data.utmSource}
    isPreview={false}
  />
</main>