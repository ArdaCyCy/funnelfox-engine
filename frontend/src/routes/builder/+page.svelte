<script lang="ts">
  import { env } from '$env/dynamic/public';
  import FunnelPlayer from '$lib/components/FunnelPlayer.svelte';
  import type { AnalyticsResponse, Funnel } from '$lib/types/funnel';

  const backendUrl = env.PUBLIC_API_URL || 'https://funnelfox-v2-api.onrender.com';

  let productName = $state('DEL LUPO 350 GSM Heavyweight Hoodie');
  let targetAudience = $state('Sokak modasÄ± tutkunlarÄ±, sÄ±nÄ±rlÄ± sayÄ±da ilk drop, Ã¼cretsiz kargo');
  let activeFunnel = $state<Funnel | null>(null);
  let selectedVariant = $state<'A' | 'B'>('A');
  let analytics = $state<AnalyticsResponse | null>(null);
  let loading = $state(false);
  let saving = $state(false);

  async function loadInitial() {
    const res = await fetch(`${backendUrl}/api/funnels`).catch(() => null);
    if (res && res.ok) {
      const list: Funnel[] = await res.json();
      if (list.length > 0) {
        activeFunnel = list[0];
        await refreshAnalytics();
      }
    }
  }

  async function generateFunnel(e: Event) {
    e.preventDefault();
    loading = true;
    const res = await fetch(`${backendUrl}/api/funnels/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ productName, targetAudience })
    });
    if (res.ok) {
      activeFunnel = await res.json();
      selectedVariant = 'A';
      await refreshAnalytics();
    }
    loading = false;
  }

  async function saveEdits() {
    if (!activeFunnel) return;
    saving = true;
    const res = await fetch(`${backendUrl}/api/funnels/${activeFunnel.slug}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        autoWinner: activeFunnel.autoWinner,
        variants: activeFunnel.variants
      })
    });
    if (res.ok) {
      activeFunnel = await res.json();
    }
    saving = false;
  }

  async function refreshAnalytics() {
    if (!activeFunnel) return;
    const res = await fetch(`${backendUrl}/api/analytics/${activeFunnel.slug}`).catch(() => null);
    if (res && res.ok) {
      analytics = await res.json();
    }
  }

  $effect(() => {
    loadInitial();
  });
</script>

<div class="min-h-screen bg-slate-100 text-slate-900">
  <header class="border-b border-slate-200 bg-white px-6 py-4">
    <div class="mx-auto flex max-w-7xl items-center justify-between">
      <div class="flex items-center gap-3">
        <a href="/" class="rounded-lg bg-slate-900 px-2.5 py-1 text-xs font-black text-white">FF V2</a>
        <h1 class="text-lg font-extrabold">Funnel Studio & Real-Time A/B Engine</h1>
      </div>
      {#if activeFunnel}
        <div class="flex items-center gap-3">
          <a
            href="/f/{activeFunnel.slug}?utm_source=tiktok"
            target="_blank"
            class="rounded-xl bg-emerald-600 px-4 py-2 text-xs font-bold text-white hover:bg-emerald-700"
          >
            CanlÄ± Linki AÃ§ (/f/{activeFunnel.slug}) â†—
          </a>
        </div>
      {/if}
    </div>
  </header>

  <main class="mx-auto grid max-w-7xl gap-6 p-6 lg:grid-cols-12">
    <section class="space-y-6 lg:col-span-4">
      <form onsubmit={generateFunnel} class="space-y-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 class="text-sm font-extrabold tracking-wide text-slate-700 uppercase">1. AI Funnel Ãœretici</h2>
        <div>
          <label class="mb-1 block text-xs font-bold text-slate-600">ÃœrÃ¼n / Marka AdÄ±</label>
          <input bind:value={productName} required class="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label class="mb-1 block text-xs font-bold text-slate-600">Hedef Kitle & Teklif AÃ§Ä±sÄ±</label>
          <textarea bind:value={targetAudience} rows="2" required class="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"></textarea>
        </div>
        <button
          type="submit"
          disabled={loading}
          class="w-full rounded-xl bg-indigo-600 py-3 text-sm font-bold text-white hover:bg-indigo-700"
        >
          {loading ? 'AI Varyantlar Ãœretiliyor...' : 'AI ile A/B Funnel Ãœret'}
        </button>
      </form>

      {#if activeFunnel}
        <div class="space-y-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-extrabold tracking-wide text-slate-700 uppercase">2. CanlÄ± Varyant DÃ¼zenleyici</h2>
            <div class="flex rounded-lg bg-slate-100 p-1">
              <button
                onclick={() => (selectedVariant = 'A')}
                class="rounded-md px-3 py-1 text-xs font-bold {selectedVariant === 'A' ? 'bg-white shadow' : 'text-slate-500'}"
              >
                Varyant A
              </button>
              <button
                onclick={() => (selectedVariant = 'B')}
                class="rounded-md px-3 py-1 text-xs font-bold {selectedVariant === 'B' ? 'bg-white shadow' : 'text-slate-500'}"
              >
                Varyant B
              </button>
            </div>
          </div>

          <div>
            <label class="mb-1 block text-xs font-bold text-slate-600">Hook BaÅŸlÄ±ÄŸÄ± ({selectedVariant})</label>
            <input
              bind:value={activeFunnel.variants[selectedVariant].hook.headline}
              class="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs font-bold text-slate-600">Quiz Sorusu</label>
            <input
              bind:value={activeFunnel.variants[selectedVariant].quiz.question}
              class="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
            />
          </div>
          <label class="flex items-center gap-2 text-xs font-semibold text-slate-700">
            <input type="checkbox" bind:checked={activeFunnel.autoWinner} />
            Auto-Winner (%95 gÃ¼vende trafiÄŸin %90'Ä±nÄ± kazanana kaydÄ±r)
          </label>
          <button
            onclick={saveEdits}
            disabled={saving}
            class="w-full rounded-xl bg-slate-900 py-2.5 text-xs font-bold text-white hover:bg-slate-800"
          >
            {saving ? 'Kaydediliyor...' : 'DeÄŸiÅŸiklikleri Kaydet & Ã–nbelleÄŸi Yenile'}
          </button>
        </div>
      {/if}
    </section>

    <section class="flex flex-col items-center justify-start rounded-2xl border border-slate-200 bg-slate-200/60 p-6 lg:col-span-4">
      <div class="mb-4 flex items-center gap-2">
        <span class="rounded-full bg-white px-3 py-1 text-xs font-bold text-slate-700 shadow-sm">
          CanlÄ± Ã–nizleme â€¢ Varyant {selectedVariant}
        </span>
      </div>
      {#if activeFunnel}
        {#key `${activeFunnel.slug}-${selectedVariant}`}
          <FunnelPlayer
            slug={activeFunnel.slug}
            variantKey={selectedVariant}
            variant={activeFunnel.variants[selectedVariant]}
            isPreview={true}
          />
        {/key}
      {/if}
    </section>

    <section class="space-y-4 lg:col-span-4">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-sm font-extrabold tracking-wide text-slate-700 uppercase">3. CanlÄ± A/B & Z-Test AnalitiÄŸi</h2>
          <button onclick={refreshAnalytics} class="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold hover:bg-slate-200">
            Yenile
          </button>
        </div>

        {#if analytics}
          <div class="mb-4 grid grid-cols-2 gap-3">
            <div class="rounded-xl bg-indigo-50 p-3">
              <div class="text-xs font-semibold text-indigo-600">Ä°statistiksel GÃ¼ven</div>
              <div class="text-xl font-black text-indigo-950">%{analytics.confidencePct}</div>
              <div class="text-xs text-indigo-700">Lider: Varyant {analytics.winner}</div>
            </div>
            <div class="rounded-xl bg-emerald-50 p-3">
              <div class="text-xs font-semibold text-emerald-600">Cache Hit OranÄ±</div>
              <div class="text-xl font-black text-emerald-950">%{analytics.cacheHitRate}</div>
              <div class="text-xs text-emerald-700">Sub-50ms Motor</div>
            </div>
          </div>

          {#each ['A', 'B'] as v}
            {@const m = analytics.variants[v]}
            <div class="mb-3 rounded-xl border border-slate-200 p-4">
              <div class="flex items-center justify-between">
                <span class="font-extrabold text-slate-800">Varyant {v}</span>
                <span class="rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-bold text-emerald-800">
                  ${m.revenueUsd.toFixed(2)} Ciro
                </span>
              </div>
              <div class="mt-2 grid grid-cols-4 gap-2 text-center text-xs">
                <div class="rounded-lg bg-slate-50 p-2">
                  <div class="font-bold text-slate-900">{m.views}</div>
                  <div class="text-slate-500">GÃ¶rÃ¼ntÃ¼leme</div>
                </div>
                <div class="rounded-lg bg-slate-50 p-2">
                  <div class="font-bold text-slate-900">{m.quizCompleted}</div>
                  <div class="text-slate-500">Quiz</div>
                </div>
                <div class="rounded-lg bg-slate-50 p-2">
                  <div class="font-bold text-slate-900">{m.purchases}</div>
                  <div class="text-slate-500">SatÄ±ÅŸ</div>
                </div>
                <div class="rounded-lg bg-slate-50 p-2">
                  <div class="font-bold text-indigo-600">%{m.conversionPct}</div>
                  <div class="text-slate-500">DÃ¶nÃ¼ÅŸÃ¼m</div>
                </div>
              </div>
            </div>
          {/each}

          <div class="rounded-xl border border-slate-200 p-4">
            <div class="mb-2 text-xs font-bold text-slate-600 uppercase">UTM Trafik KaynaklarÄ±</div>
            {#if Object.keys(analytics.utmBreakdown).length === 0}
              <p class="text-xs text-slate-400">HenÃ¼z trafik kaydÄ± yok. CanlÄ± linki aÃ§Ä±p test edin.</p>
            {:else}
              <div class="space-y-1 text-xs">
                {#each Object.entries(analytics.utmBreakdown) as [src, count]}
                  <div class="flex justify-between font-medium">
                    <span>utm_source={src}</span>
                    <span class="font-bold">{count} ziyaret</span>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </section>
  </main>
</div>
