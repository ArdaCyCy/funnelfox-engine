import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import type { PageServerLoad } from './$types';
import type { FunnelRenderResponse } from '$lib/types/funnel';

export const load: PageServerLoad = async ({ params, fetch, url, cookies }) => {
  let visitorId = cookies.get('ff_vid');
  if (!visitorId) {
    visitorId = crypto.randomUUID();
    cookies.set('ff_vid', visitorId, { path: '/', httpOnly: true, sameSite: 'lax', maxAge: 60 * 60 * 24 * 30 });
  }

  const backendUrl = env.PUBLIC_API_URL || 'https://funnelfox-v2-api.onrender.com';
  const forcedVariant = url.searchParams.get('variant');
  const utmSource = url.searchParams.get('utm_source') || 'direct';
  const query = forcedVariant ? `?variant=${encodeURIComponent(forcedVariant)}` : '';

  const res = await fetch(`${backendUrl}/api/f/${encodeURIComponent(params.slug)}${query}`, {
    headers: { 'X-Visitor-ID': visitorId }
  });

  if (!res.ok) throw error(res.status, 'Satış hunisi bulunamadı.');
  const funnel = (await res.json()) as FunnelRenderResponse;
  return { funnel, utmSource };
};