export interface QuizOption {
  label: string;
  offerKey: string;
}

export interface OfferDetail {
  title: string;
  price: string;
  priceCents: number;
  description: string;
  features: string[];
  ctaText: string;
}

export interface FunnelVariant {
  hook: {
    badge: string;
    headline: string;
    subhead: string;
    ctaText: string;
  };
  quiz: {
    question: string;
    options: QuizOption[];
  };
  offers: Record<string, OfferDetail>;
}

export interface Funnel {
  slug: string;
  productName: string;
  targetAudience: string;
  autoWinner: boolean;
  variants: Record<string, FunnelVariant>;
  createdAt: string;
}

export interface FunnelRenderResponse {
  slug: string;
  productName: string;
  variant: string;
  content: FunnelVariant;
  latencyUs: number;
}

export interface VariantMetrics {
  views: number;
  quizCompleted: number;
  checkoutViews: number;
  purchases: number;
  revenueUsd: number;
  conversionPct: number;
}

export interface AnalyticsResponse {
  slug: string;
  autoWinner: boolean;
  variants: Record<string, VariantMetrics>;
  utmBreakdown: Record<string, number>;
  winner: string;
  confidencePct: number;
  cacheHitRate: number;
}