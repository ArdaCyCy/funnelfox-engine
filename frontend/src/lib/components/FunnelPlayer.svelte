<script lang="ts">
  import { env } from '$env/dynamic/public';
  import type { FunnelVariant } from '$lib/types/funnel';

  let {
    slug,
    variantKey = 'A',
    variant,
    utmSource = 'direct',
    isPreview = false
  }: {
    slug: string;
    variantKey?: string;
    variant: FunnelVariant;
    utmSource?: string;
    isPreview?: boolean;
  } = $props();

  let step = $state(0);
  let selectedOfferKey = $state('premium');
  let selectedOptionLabel = $state('');
  let checkoutOpen = $state(false);
  let paymentSuccess = $state(false);
  let processing = $state(false);

  const backendUrl = env.PUBLIC_API_URL || 'http://localhost:8080';

  function track(event: string, revenueCents = 0) {
    if (isPreview) return;
    const payload = JSON.stringify({
      slug,
      variant: variantKey,
      event,
      utmSource,
      revenueCents
    });
    if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
      navigator.sendBeacon(`${backendUrl}/api/track`, new Blob([payload], { type: 'application/json' }));
    } else {
      fetch(`${backendUrl}/api/track`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
        keepalive: true
      }).catch(() => {});
    }
  }

  $effect(() => {
    track('view');
  });

  let activeOffer = $derived(
    variant.offers?.[selectedOfferKey] ||
      Object.values(variant.offers || {})[0] || {
        title: 'Özel Teklif Paketi',
        price: '$79.00',
        priceCents: 7900,
        description: 'Sınırlı süreli özel erişim paketi.',
        features: ['Hızlı Teslimat', '30 Gün İade Garantisi'],
        ctaText: 'Siparişi Tamamla'
      }
  );

  let progress = $derived(paymentSuccess ? 100 : step === 0 ? 33 : step === 1 ? 66 : 95);

  function startQuiz() {
    step = 1;
  }

  function chooseOption(option: { label: string; offerKey: string }) {
    selectedOptionLabel = option.label;
    selectedOfferKey = option.offerKey || 'premium';
    track('step_complete');
    step = 2;
  }

  function openCheckout() {
    track('checkout_click');
    checkoutOpen = true;
  }

  async function completePayment(e: Event) {
    e.preventDefault();
    processing = true;
    if (!isPreview) {
      await fetch(`${backendUrl}/api/webhooks/payment`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          slug,
          variant: variantKey,
          utmSource,
          revenueCents: activeOffer.priceCents || 8900
        })
      }).catch(() => {});
    }
    processing = false;
    checkoutOpen = false;
    paymentSuccess = true;
  }
</script>

<div class="w-full max-w-lg overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-2xl">
  <div class="h-1.5 w-full bg-slate-100">
    <div class="h-full bg-indigo-600 transition-all duration-300" style="width: {progress}%"></div>
  </div>

  <div class="p-8">
    {#if paymentSuccess}
      <div class="space-y-4 py-6 text-center">
        <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-100 text-3xl text-emerald-600">
          ✓
        </div>
        <h2 class="text-2xl font-extrabold text-slate-900">Siparişin Onaylandı!</h2>
        <p class="text-sm text-slate-600">
          <strong>{activeOffer.title}</strong> ({activeOffer.price}) ödemesi başarıyla alındı ve analitik motoruna anlık işlendi.
        </p>
        <button
          onclick={() => {
            paymentSuccess = false;
            step = 0;
          }}
          class="mt-4 rounded-xl border border-slate-300 px-5 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50"
        >
          Huniyi Baştan Test Et
        </button>
      </div>
    {:else if checkoutOpen}
      <form onsubmit={completePayment} class="space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <span class="text-xs font-bold tracking-wider text-indigo-600 uppercase">Web2App Instant Checkout</span>
          <span class="text-lg font-extrabold text-slate-900">{activeOffer.price}</span>
        </div>
        <p class="text-sm font-medium text-slate-700">{activeOffer.title}</p>
        <div>
          <label class="mb-1 block text-xs font-semibold text-slate-600">E-posta Adresi</label>
          <input type="email" required value="arda@dellupo.co" class="w-full rounded-xl border border-slate-300 px-3.5 py-2.5 text-sm" />
        </div>
        <div>
          <label class="mb-1 block text-xs font-semibold text-slate-600">Kart Bilgisi (Simülasyon)</label>
          <input type="text" required value="4242 •••• •••• 4242   12/28   888" class="w-full rounded-xl border border-slate-300 bg-slate-50 px-3.5 py-2.5 font-mono text-sm" />
        </div>
        <div class="flex gap-2 pt-2">
          <button
            type="button"
            onclick={() => (checkoutOpen = false)}
            class="rounded-xl border border-slate-300 px-4 py-3 text-sm font-semibold text-slate-600 hover:bg-slate-50"
          >
            ← Geri
          </button>
          <button
            type="submit"
            disabled={processing}
            class="flex-1 rounded-xl bg-emerald-600 py-3 text-sm font-bold text-white shadow-lg hover:bg-emerald-700"
          >
            {processing ? 'İşleniyor...' : `${activeOffer.price} Ödemeyi Tamamla`}
          </button>
        </div>
      </form>
    {:else if step === 0}
      <div class="space-y-5 text-center">
        <span class="inline-block rounded-full bg-indigo-50 px-3.5 py-1 text-xs font-bold tracking-wide text-indigo-700">
          {variant.hook.badge}
        </span>
        <h1 class="text-3xl leading-tight font-extrabold text-slate-900">
          {variant.hook.headline}
        </h1>
        <p class="text-base leading-relaxed text-slate-600">
          {variant.hook.subhead}
        </p>
        <button
          onclick={startQuiz}
          class="w-full rounded-2xl bg-indigo-600 px-6 py-4 text-base font-bold text-white shadow-lg shadow-indigo-200 transition hover:bg-indigo-700"
        >
          {variant.hook.ctaText}
        </button>
      </div>
    {:else if step === 1}
      <div class="space-y-5">
        <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
          <span>ADIM 2 / 3 • KİŞİSELLEŞTİRME</span>
          <button onclick={() => (step = 0)} class="text-slate-500 hover:text-slate-800">← Geri Dön</button>
        </div>
        <h2 class="text-2xl font-bold text-slate-900">{variant.quiz.question}</h2>
        <div class="space-y-3">
          {#each variant.quiz.options as opt}
            <button
              onclick={() => chooseOption(opt)}
              class="flex w-full items-center justify-between rounded-2xl border-2 border-slate-200 p-4 text-left font-semibold text-slate-800 transition hover:border-indigo-600 hover:bg-indigo-50/50"
            >
              <span>{opt.label}</span>
              <span class="text-indigo-600">→</span>
            </button>
          {/each}
        </div>
      </div>
    {:else}
      <div class="space-y-5">
        <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
          <span>SEÇİMİN: {selectedOptionLabel.toUpperCase()}</span>
          <button onclick={() => (step = 1)} class="text-slate-500 hover:text-slate-800">← Geri Dön</button>
        </div>
        <div class="rounded-2xl bg-slate-900 p-6 text-white">
          <div class="flex items-baseline justify-between">
            <h2 class="text-xl font-bold">{activeOffer.title}</h2>
            <span class="text-2xl font-extrabold text-emerald-400">{activeOffer.price}</span>
          </div>
          <p class="mt-2 text-sm text-slate-300">{activeOffer.description}</p>
          <ul class="mt-4 space-y-2 text-sm text-slate-200">
            {#each activeOffer.features as feat}
              <li class="flex items-center gap-2">
                <span class="text-emerald-400">✓</span> {feat}
              </li>
            {/each}
          </ul>
        </div>
        <button
          onclick={openCheckout}
          class="w-full rounded-2xl bg-emerald-600 px-6 py-4 text-base font-bold text-white shadow-lg shadow-emerald-200 transition hover:bg-emerald-700"
        >
          {activeOffer.ctaText}
        </button>
      </div>
    {/if}
  </div>
</div>