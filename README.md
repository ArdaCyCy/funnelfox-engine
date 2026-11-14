# FunnelFox V2 — Sub-50ms AI Web Funnel & Real-Time A/B Analytics Engine

A high-throughput, low-latency **Web2App & E-Commerce Monetization Funnel Engine** built with **Go (Fiber v2)**, **Svelte 5 (SvelteKit Runes)**, **PostgreSQL**, and **Redis**. Designed to generate AI-powered multi-step funnels, serve visitor experiences in **<50ms**, execute deterministic A/B traffic splitting, and compute real-time statistical significance (Two-Proportion Z-Test) with automated winner routing.

## System Architecture
- **Edge SSR & Client (SvelteKit 5):** Issues 30-day `HttpOnly` visitor cookies (`ff_vid`) and fires zero-blocking `navigator.sendBeacon()` telemetry.
- **High-Throughput API (Go Fiber v2):** Serves cached funnel configs in microseconds via Cache-Aside pattern, hashes `slug + visitorID` via FNV-1a for flicker-free A/B routing, and processes analytics via buffered Goroutine worker channels.
- **Persistence & Cache:** Hybrid adapter supporting Neon PostgreSQL (`JSONB`) + Upstash Redis in production with automatic In-Memory fallback locally.

## Core Engineering Highlights
- **Sub-50ms Cache-Aside Delivery:** Microsecond reads with automatic cache invalidation on live variant edits (`PUT /api/funnels/:slug`).
- **Two-Proportion Z-Test & Auto-Winner:** Calculates real-time statistical confidence (`Erf(z / sqrt(2))`) and shifts 90% of traffic to the winning variant once `>= 95%` confidence is reached.
- **Conditional Quiz Branching & Web2App Checkout:** Dynamically routes visitors to personalized offers based on quiz answers and records checkout revenue + `utm_source` attribution.

## Local Quickstart
1. **Backend:** `cd backend && go mod tidy && go test -v ./... && go run main.go`
2. **Frontend:** `cd frontend && npm install && npm run dev`

## Author
Built by **Arda** ([@LUPOFACE](https://github.com/LUPOFACE))