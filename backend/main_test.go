package main

import (
"net/http/httptest"
"strings"
"testing"
)

func TestDeterministicVariantAssignment(t *testing.T) {
v1 := assignVariant("del-lupo-350gsm", "visitor-123", false, "TIE", 0)
for i := 0; i < 50; i++ {
if got := assignVariant("del-lupo-350gsm", "visitor-123", false, "TIE", 0); got != v1 {
t.Fatalf("Varyant tutarsızlığı: beklenen %s, gelen %s", v1, got)
}
}
}

func TestStatisticalConfidenceCalculation(t *testing.T) {
a := VariantMetrics{Views: 500, Purchases: 25}
b := VariantMetrics{Views: 500, Purchases: 60}
winner, conf := calculateABStats(a, b)
if winner != "B" || conf < 95.0 {
t.Fatalf("İstatistiksel test başarısız: winner=%s, conf=%.2f", winner, conf)
}
}

func TestSub50msFunnelEndpoint(t *testing.T) {
store := NewStore()
app := setupApp(store)

req := httptest.NewRequest("GET", "/api/f/del-lupo-350gsm", nil)
req.Header.Set("X-Visitor-ID", "test-user-99")
resp, err := app.Test(req)
if err != nil || resp.StatusCode != 200 {
t.Fatalf("Funnel endpoint hatası: %v, status=%d", err, resp.StatusCode)
}

payReq := httptest.NewRequest("POST", "/api/webhooks/payment", strings.NewReader(`{"slug":"del-lupo-350gsm","variant":"B","utmSource":"tiktok","revenueCents":8900}`))
payReq.Header.Set("Content-Type", "application/json")
payResp, _ := app.Test(payReq)
if payResp.StatusCode != 200 {
t.Fatalf("Webhook ödeme hatası: status=%d", payResp.StatusCode)
}
}