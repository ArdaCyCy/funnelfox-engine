<script lang="ts">
  const features = [
    {
      title: 'Sub-50ms Cache-Aside Engine',
      desc: 'Go Fiber + Redis/In-Memory mimarisi ile ziyaretçi sayfalarını veritabanına dokunmadan mikrosaniyeler içinde sunar.'
    },
    {
      title: 'Deterministik FNV-1a A/B Testi',
      desc: 'HttpOnly çerezler ve FNV-1a hash algoritması ile aynı ziyaretçiye her yenilemede tutarlı varyant gösterir.'
    },
    {
      title: 'Two-Proportion Z-Test & Auto-Winner',
      desc: '%95 istatistiksel güven eşiğini geçen varyantı matematiksel olarak tespit eder ve trafiğin %90ını otomatik yönlendirir.'
    },
    {
      title: 'Koşullu Dallanma & Web2App Checkout',
      desc: 'Quiz yanıtına göre kişiselleştirilmiş teklif sunar, sendBeacon + Goroutine hattı ile sıfır gecikmeli gelir takibi yapar.'
    }
  ];
</script>

<div class="min-h-screen bg-slate-950 text-white">
  <header class="mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
    <div class="flex items-center gap-2.5">
      <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-600 font-black">FF</div>
      <span class="text-lg font-extrabold tracking-tight">FunnelFox <span class="text-indigo-400">V2 Engine</span></span>
    </div>
    <div class="flex items-center gap-4">
      <a href="/f/del-lupo-350gsm?utm_source=showcase" class="text-sm font-semibold text-slate-300 hover:text-white">
        Canlı Demo Huni
      </a>
      <a href="/builder" class="rounded-xl bg-indigo-600 px-4 py-2.5 text-sm font-bold text-white hover:bg-indigo-500">
        Funnel Studio'yu Aç →
      </a>
    </div>
  </header>

  <main class="mx-auto max-w-6xl px-6 py-16">
    <div class="max-w-3xl space-y-6">
      <span class="inline-flex items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/10 px-4 py-1.5 text-xs font-bold text-indigo-300">
        GO FIBER V2 • SVELTE 5 RUNES • REDIS CACHE-ASIDE • Z-TEST ANALYTICS
      </span>
      <h1 class="text-5xl leading-tight font-black tracking-tight">
        50ms Altında Yanıt Veren <span class="text-indigo-400">AI Web Funnel & A/B</span> Motoru.
      </h1>
      <p class="text-lg leading-relaxed text-slate-400">
        Yüksek trafikli Web2App ve e-ticaret kampanyaları için tasarlandı. Yapay zeka ile saniyeler içinde dallanmalı satış hunileri üretin, FNV-1a ile trafiği bölün ve Z-Score istatistik motoruyla kazanan varyantı otomatik ölçekleyin.
      </p>
      <div class="flex flex-wrap gap-4 pt-2">
        <a href="/builder" class="rounded-2xl bg-indigo-600 px-7 py-4 text-base font-bold text-white shadow-xl shadow-indigo-950 hover:bg-indigo-500">
          Yeni AI Funnel Oluştur
        </a>
        <a href="/f/del-lupo-350gsm?utm_source=tiktok" class="rounded-2xl border border-slate-700 bg-slate-900 px-7 py-4 text-base font-bold text-slate-200 hover:border-slate-500">
          Örnek Canlı Huniyi Test Et
        </a>
      </div>
    </div>

    <div class="mt-16 grid gap-6 md:grid-cols-2">
      {#each features as f}
        <div class="rounded-3xl border border-slate-800 bg-slate-900/60 p-7">
          <h3 class="text-lg font-bold text-white">{f.title}</h3>
          <p class="mt-2 text-sm leading-relaxed text-slate-400">{f.desc}</p>
        </div>
      {/each}
    </div>
  </main>
</div>