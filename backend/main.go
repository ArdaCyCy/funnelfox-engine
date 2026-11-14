package main

import (
"bytes"
"context"
"database/sql"
"encoding/json"
"fmt"
"hash/fnv"
"io"
"log"
"math"
"net/http"
"os"
"regexp"
"strings"
"sync"
"time"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/cors"
"github.com/gofiber/fiber/v2/middleware/limiter"
"github.com/gofiber/fiber/v2/middleware/logger"
_ "github.com/lib/pq"
"github.com/redis/go-redis/v9"
)

type QuizOption struct {
Label     string `json:"label"`
OfferKey  string `json:"offerKey"`
}

type OfferDetail struct {
Title       string   `json:"title"`
Price       string   `json:"price"`
PriceCents  int64    `json:"priceCents"`
Description string   `json:"description"`
Features    []string `json:"features"`
CtaText     string   `json:"ctaText"`
}

type FunnelVariant struct {
Hook struct {
Badge    string `json:"badge"`
Headline string `json:"headline"`
Subhead  string `json:"subhead"`
CtaText  string `json:"ctaText"`
} `json:"hook"`
Quiz struct {
Question string       `json:"question"`
Options  []QuizOption `json:"options"`
} `json:"quiz"`
Offers map[string]OfferDetail `json:"offers"`
}

type Funnel struct {
Slug           string                   `json:"slug"`
ProductName    string                   `json:"productName"`
TargetAudience string                   `json:"targetAudience"`
AutoWinner     bool                     `json:"autoWinner"`
Variants       map[string]FunnelVariant `json:"variants"`
CreatedAt      time.Time                `json:"createdAt"`
}

type VariantMetrics struct {
Views         int64   `json:"views"`
QuizCompleted int64   `json:"quizCompleted"`
CheckoutViews int64   `json:"checkoutViews"`
Purchases     int64   `json:"purchases"`
RevenueUSD    float64 `json:"revenueUsd"`
ConversionPct float64 `json:"conversionPct"`
}

type AnalyticsResponse struct {
Slug          string                    `json:"slug"`
AutoWinner    bool                      `json:"autoWinner"`
Variants      map[string]VariantMetrics `json:"variants"`
UtmBreakdown  map[string]int64          `json:"utmBreakdown"`
Winner        string                    `json:"winner"`
ConfidencePct float64                   `json:"confidencePct"`
CacheHitRate  float64                   `json:"cacheHitRate"`
}

type Store struct {
mu           sync.RWMutex
funnels      map[string]Funnel
cache        map[string]Funnel
metrics      map[string]map[string]*VariantMetrics
utm          map[string]map[string]int64
cacheHits    int64
cacheMisses  int64
db           *sql.DB
rdb          *redis.Client
eventChan    chan trackPayload
}

type trackPayload struct {
Slug         string `json:"slug"`
Variant      string `json:"variant"`
Event        string `json:"event"`
UtmSource    string `json:"utmSource"`
RevenueCents int64  `json:"revenueCents"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9-]{2,80}$`)

func NewStore() *Store {
s := &Store{
funnels:   make(map[string]Funnel),
cache:     make(map[string]Funnel),
metrics:   make(map[string]map[string]*VariantMetrics),
utm:       make(map[string]map[string]int64),
eventChan: make(chan trackPayload, 2048),
}

if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
if db, err := sql.Open("postgres", dbURL); err == nil && db.Ping() == nil {
s.db = db
if schemaBytes, err := os.ReadFile("schema.sql"); err == nil {
_, _ = s.db.Exec(string(schemaBytes))
}
log.Println("PostgreSQL bağlantısı aktif.")
}
}

if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
if opt, err := redis.ParseURL(redisURL); err == nil {
rdb := redis.NewClient(opt)
if rdb.Ping(context.Background()).Err() == nil {
s.rdb = rdb
log.Println("Redis önbellek bağlantısı aktif.")
}
}
}

if s.db == nil && s.rdb == nil {
log.Println("Yerel mod: Yüksek hızlı In-Memory + Cache-Aside motoru aktif.")
}

s.seedDefaultFunnel()
go s.telemetryWorker()
return s
}

func (s *Store) seedDefaultFunnel() {
f := generateFallbackFunnel("del-lupo-350gsm", "DEL LUPO 350 GSM Heavyweight Hoodie", "Sokak modası tutkunları, sınırlı sayıda ilk drop")
_ = s.SaveFunnel(f)
}

func (s *Store) telemetryWorker() {
for ev := range s.eventChan {
s.applyEvent(ev)
}
}

func (s *Store) applyEvent(ev trackPayload) {
if ev.Variant != "A" && ev.Variant != "B" {
ev.Variant = "A"
}
if ev.UtmSource == "" {
ev.UtmSource = "direct"
}

s.mu.Lock()
if _, ok := s.metrics[ev.Slug]; !ok {
s.metrics[ev.Slug] = map[string]*VariantMetrics{
"A": {},
"B": {},
}
}
if _, ok := s.utm[ev.Slug]; !ok {
s.utm[ev.Slug] = make(map[string]int64)
}

m := s.metrics[ev.Slug][ev.Variant]
switch ev.Event {
case "view":
m.Views++
s.utm[ev.Slug][ev.UtmSource]++
case "step_complete":
m.QuizCompleted++
case "checkout_click":
m.CheckoutViews++
case "purchase":
m.Purchases++
m.RevenueUSD += float64(ev.RevenueCents) / 100.0
}
if m.Views > 0 {
m.ConversionPct = math.Round((float64(m.Purchases)/float64(m.Views))*10000) / 100
}
s.mu.Unlock()

if s.db != nil {
_, _ = s.db.Exec(
`INSERT INTO funnel_events (slug, variant, event_type, utm_source, revenue_cents) VALUES ($1, $2, $3, $4, $5)`,
ev.Slug, ev.Variant, ev.Event, ev.UtmSource, ev.RevenueCents,
)
}
}

func (s *Store) SaveFunnel(f Funnel) error {
raw, err := json.Marshal(f.Variants)
if err != nil {
return err
}

if s.db != nil {
_, err = s.db.Exec(`
INSERT INTO funnels (slug, product_name, target_audience, auto_winner, variants_json)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (slug) DO UPDATE SET
product_name = EXCLUDED.product_name,
target_audience = EXCLUDED.target_audience,
auto_winner = EXCLUDED.auto_winner,
variants_json = EXCLUDED.variants_json
`, f.Slug, f.ProductName, f.TargetAudience, f.AutoWinner, string(raw))
if err != nil {
return err
}
}

s.mu.Lock()
s.funnels[f.Slug] = f
delete(s.cache, f.Slug)
if _, ok := s.metrics[f.Slug]; !ok {
s.metrics[f.Slug] = map[string]*VariantMetrics{"A": {}, "B": {}}
}
s.mu.Unlock()

if s.rdb != nil {
_ = s.rdb.Del(context.Background(), "funnel:"+f.Slug).Err()
}
return nil
}

func (s *Store) GetFunnel(slug string) (Funnel, bool) {
if s.rdb != nil {
if val, err := s.rdb.Get(context.Background(), "funnel:"+slug).Result(); err == nil {
var f Funnel
if json.Unmarshal([]byte(val), &f) == nil {
s.mu.Lock()
s.cacheHits++
s.mu.Unlock()
return f, true
}
}
}

s.mu.RLock()
if f, ok := s.cache[slug]; ok {
s.mu.RUnlock()
s.mu.Lock()
s.cacheHits++
s.mu.Unlock()
return f, true
}
s.mu.RUnlock()

s.mu.Lock()
s.cacheMisses++
s.mu.Unlock()

if s.db != nil {
var f Funnel
var raw []byte
err := s.db.QueryRow(`SELECT slug, product_name, target_audience, auto_winner, variants_json, created_at FROM funnels WHERE slug = $1`, slug).
Scan(&f.Slug, &f.ProductName, &f.TargetAudience, &f.AutoWinner, &raw, &f.CreatedAt)
if err == nil && json.Unmarshal(raw, &f.Variants) == nil {
s.setCache(f)
return f, true
}
}

s.mu.RLock()
f, ok := s.funnels[slug]
s.mu.RUnlock()
if ok {
s.setCache(f)
return f, true
}
return Funnel{}, false
}

func (s *Store) setCache(f Funnel) {
s.mu.Lock()
s.cache[f.Slug] = f
s.mu.Unlock()
if s.rdb != nil {
if raw, err := json.Marshal(f); err == nil {
_ = s.rdb.Set(context.Background(), "funnel:"+f.Slug, string(raw), 10*time.Minute).Err()
}
}
}

func (s *Store) ListFunnels() []Funnel {
s.mu.RLock()
defer s.mu.RUnlock()
list := make([]Funnel, 0, len(s.funnels))
for _, f := range s.funnels {
list = append(list, f)
}
return list
}

func assignVariant(slug, visitorID string, autoWinner bool, winner string, confidence float64) string {
h := fnv.New32a()
_, _ = h.Write([]byte(slug + ":" + visitorID))
bucket := h.Sum32() % 100

if autoWinner && confidence >= 95.0 && (winner == "A" || winner == "B") {
if bucket < 90 {
return winner
}
if winner == "A" {
return "B"
}
return "A"
}

if bucket < 50 {
return "A"
}
return "B"
}

func calculateABStats(a, b VariantMetrics) (string, float64) {
if a.Views < 5 || b.Views < 5 {
if a.ConversionPct > b.ConversionPct {
return "A", 50.0
} else if b.ConversionPct > a.ConversionPct {
return "B", 50.0
}
return "TIE", 0.0
}

p1 := float64(a.Purchases) / float64(a.Views)
p2 := float64(b.Purchases) / float64(b.Views)
pPool := float64(a.Purchases+b.Purchases) / float64(a.Views+b.Views)
if pPool <= 0 || pPool >= 1 {
return "TIE", 0.0
}

se := math.Sqrt(pPool * (1 - pPool) * (1.0/float64(a.Views) + 1.0/float64(b.Views)))
if se == 0 {
return "TIE", 0.0
}

z := math.Abs(p1-p2) / se
conf := math.Erf(z/math.Sqrt2) * 100.0
conf = math.Round(conf*10) / 10

winner := "TIE"
if p1 > p2 {
winner = "A"
} else if p2 > p1 {
winner = "B"
}
return winner, conf
}

func (s *Store) GetAnalytics(slug string) AnalyticsResponse {
s.mu.RLock()
defer s.mu.RUnlock()

mMap, ok := s.metrics[slug]
if !ok {
mMap = map[string]*VariantMetrics{"A": {}, "B": {}}
}
a := *mMap["A"]
b := *mMap["B"]
winner, conf := calculateABStats(a, b)

utmCopy := make(map[string]int64)
for k, v := range s.utm[slug] {
utmCopy[k] = v
}

totalCache := s.cacheHits + s.cacheMisses
hitRate := 100.0
if totalCache > 0 {
hitRate = math.Round((float64(s.cacheHits)/float64(totalCache))*1000) / 10
}

autoWin := false
if f, exists := s.funnels[slug]; exists {
autoWin = f.AutoWinner
}

return AnalyticsResponse{
Slug:          slug,
AutoWinner:    autoWin,
Variants:      map[string]VariantMetrics{"A": a, "B": b},
UtmBreakdown:  utmCopy,
Winner:        winner,
ConfidencePct: conf,
CacheHitRate:  hitRate,
}
}

func generateFallbackFunnel(slug, product, audience string) Funnel {
varA := FunnelVariant{}
varA.Hook.Badge = "SINIRLI DROP • ÖZEL ERİŞİM"
varA.Hook.Headline = fmt.Sprintf("%s ile Standartları Yeniden Tanımla", product)
varA.Hook.Subhead = fmt.Sprintf("%s için özel geliştirilen yüksek performanslı deneyim. Sana özel paketi 30 saniyede belirleyelim.", audience)
varA.Hook.CtaText = "Sana Özel Seçimi Başlat →"
varA.Quiz.Question = "Senin için en kritik öncelik hangisi?"
varA.Quiz.Options = []QuizOption{
{Label: "Maksimum Kalite & Premium Detaylar", OfferKey: "premium"},
{Label: "Fiyat / Performans & Hızlı Başlangıç", OfferKey: "starter"},
}
varA.Offers = map[string]OfferDetail{
"premium": {
Title:       product + " — Collector's Edition",
Price:       "$89.00",
PriceCents:  8900,
Description: "En yüksek kalite standartlarında, öncelikli kargo ve özel seri numarası ile.",
Features:    []string{"Ücretsiz Ekspres Teslimat", "30 Gün Koşulsuz İade Garantisi", "VIP Drop Erken Erişim Hakkı"},
CtaText:     "Hemen Siparişi Tamamla",
},
"starter": {
Title:       product + " — Core Drop Paketi",
Price:       "$59.00",
PriceCents:  5900,
Description: "İlk kez deneyimleyecekler için özel lansman indirimli giriş paketi.",
Features:    []string{"Lansmana Özel %25 İndirim", "Aynı Gün Kargoya Teslim", "7/24 Öncelikli Destek"},
CtaText:     "İndirimli Paketi Al",
},
}

varB := FunnelVariant{}
varB.Hook.Badge = "SON 48 SAAT • %25 LANSMAN FIRSATI"
varB.Hook.Headline = fmt.Sprintf("Neden Herkes %s Konuşuyor?", product)
varB.Hook.Subhead = fmt.Sprintf("Sıradan seçeneklerle vakit kaybetme. %s için tasarlanan en hızlı çözümü hemen keşfet.", audience)
varB.Hook.CtaText = "30 Saniyede Test Et →"
varB.Quiz.Question = "Bu ürünü hangi sıklıkla kullanmayı planlıyorsun?"
varB.Quiz.Options = []QuizOption{
{Label: "Her Gün / Yoğun Kullanım", OfferKey: "premium"},
{Label: "Özel Günler / İlk Deneyim", OfferKey: "starter"},
}
varB.Offers = varA.Offers

return Funnel{
Slug:           slug,
ProductName:    product,
TargetAudience: audience,
AutoWinner:     true,
Variants:       map[string]FunnelVariant{"A": varA, "B": varB},
CreatedAt:      time.Now(),
}
}

func generateWithOpenAI(slug, product, audience string) Funnel {
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
return generateFallbackFunnel(slug, product, audience)
}

prompt := fmt.Sprintf(`Generate a high-converting 2-variant (A and B) web funnel JSON for product "%s" targeting "%s".
Return ONLY valid JSON matching this exact structure:
{
  "A": {
    "hook": {"badge": "...", "headline": "...", "subhead": "...", "ctaText": "..."},
    "quiz": {"question": "...", "options": [{"label": "...", "offerKey": "premium"}, {"label": "...", "offerKey": "starter"}]},
    "offers": {
      "premium": {"title": "...", "price": "$89.00", "priceCents": 8900, "description": "...", "features": ["...", "...", "..."], "ctaText": "..."},
      "starter": {"title": "...", "price": "$59.00", "priceCents": 5900, "description": "...", "features": ["...", "...", "..."], "ctaText": "..."}
    }
  },
  "B": { ...same structure with different angle... }
}`, product, audience)

reqBody, _ := json.Marshal(map[string]any{
"model": "gpt-4o-mini",
"messages": []map[string]string{
{"role": "system", "content": "You are a senior conversion rate optimization copywriter. Output strict JSON only."},
{"role": "user", "content": prompt},
},
"response_format": map[string]string{"type": "json_object"},
})

req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBody))
req.Header.Set("Authorization", "Bearer "+apiKey)
req.Header.Set("Content-Type", "application/json")

client := &http.Client{Timeout: 12 * time.Second}
resp, err := client.Do(req)
if err != nil {
return generateFallbackFunnel(slug, product, audience)
}
defer resp.Body.Close()

bodyBytes, _ := io.ReadAll(resp.Body)
var aiResp struct {
Choices []struct {
Message struct {
Content string `json:"content"`
} `json:"message"`
} `json:"choices"`
}
if json.Unmarshal(bodyBytes, &aiResp) != nil || len(aiResp.Choices) == 0 {
return generateFallbackFunnel(slug, product, audience)
}

var variants map[string]FunnelVariant
if json.Unmarshal([]byte(aiResp.Choices[0].Message.Content), &variants) != nil || variants["A"].Hook.Headline == "" {
return generateFallbackFunnel(slug, product, audience)
}

return Funnel{
Slug:           slug,
ProductName:    product,
TargetAudience: audience,
AutoWinner:     true,
Variants:       variants,
CreatedAt:      time.Now(),
}
}

func slugify(s string) string {
s = strings.ToLower(strings.TrimSpace(s))
reg := regexp.MustCompile(`[^a-z0-9]+`)
s = reg.ReplaceAllString(s, "-")
return strings.Trim(s, "-")
}

func setupApp(store *Store) *fiber.App {
app := fiber.New(fiber.Config{
AppName: "FunnelFox V2 Enterprise Engine",
})

app.Use(logger.New())
app.Use(cors.New(cors.Config{
AllowOrigins: "*",
AllowHeaders: "Origin, Content-Type, Accept, X-Visitor-ID",
AllowMethods: "GET, POST, PUT, PATCH, OPTIONS",
}))
app.Use(limiter.New(limiter.Config{
Max:        300,
Expiration: 1 * time.Minute,
}))

app.Get("/health", func(c *fiber.Ctx) error {
return c.JSON(fiber.Map{"status": "ok", "engine": "FunnelFox V2", "timestamp": time.Now()})
})

app.Get("/api/funnels", func(c *fiber.Ctx) error {
return c.JSON(store.ListFunnels())
})

app.Post("/api/funnels/generate", func(c *fiber.Ctx) error {
var body struct {
ProductName    string `json:"productName"`
TargetAudience string `json:"targetAudience"`
Slug           string `json:"slug"`
}
if err := c.BodyParser(&body); err != nil {
return c.Status(400).JSON(fiber.Map{"error": "Geçersiz JSON verisi"})
}
if strings.TrimSpace(body.ProductName) == "" {
return c.Status(400).JSON(fiber.Map{"error": "Ürün adı zorunludur"})
}
slug := slugify(body.Slug)
if slug == "" {
slug = slugify(body.ProductName)
}
if !slugRegex.MatchString(slug) {
slug = fmt.Sprintf("funnel-%d", time.Now().Unix()%100000)
}

funnel := generateWithOpenAI(slug, body.ProductName, body.TargetAudience)
if err := store.SaveFunnel(funnel); err != nil {
return c.Status(500).JSON(fiber.Map{"error": "Kaydetme hatası"})
}
return c.Status(201).JSON(funnel)
})

app.Put("/api/funnels/:slug", func(c *fiber.Ctx) error {
slug := c.Params("slug")
existing, ok := store.GetFunnel(slug)
if !ok {
return c.Status(404).JSON(fiber.Map{"error": "Funnel bulunamadı"})
}
var body struct {
AutoWinner *bool                    `json:"autoWinner"`
Variants   map[string]FunnelVariant `json:"variants"`
}
if err := c.BodyParser(&body); err != nil {
return c.Status(400).JSON(fiber.Map{"error": "Geçersiz veri"})
}
if body.AutoWinner != nil {
existing.AutoWinner = *body.AutoWinner
}
if len(body.Variants) > 0 {
existing.Variants = body.Variants
}
if err := store.SaveFunnel(existing); err != nil {
return c.Status(500).JSON(fiber.Map{"error": "Güncelleme hatası"})
}
return c.JSON(existing)
})

app.Get("/api/f/:slug", func(c *fiber.Ctx) error {
start := time.Now()
slug := c.Params("slug")
funnel, ok := store.GetFunnel(slug)
if !ok {
return c.Status(404).JSON(fiber.Map{"error": "Funnel bulunamadı"})
}

visitorID := c.Get("X-Visitor-ID")
if visitorID == "" {
visitorID = c.IP()
}

forced := strings.ToUpper(c.Query("variant"))
analytics := store.GetAnalytics(slug)
variantKey := assignVariant(slug, visitorID, funnel.AutoWinner, analytics.Winner, analytics.ConfidencePct)
if forced == "A" || forced == "B" {
variantKey = forced
}

elapsedUs := time.Since(start).Microseconds()
c.Set("X-Response-Time-Us", fmt.Sprintf("%d", elapsedUs))

return c.JSON(fiber.Map{
"slug":        funnel.Slug,
"productName": funnel.ProductName,
"variant":     variantKey,
"content":     funnel.Variants[variantKey],
"latencyUs":   elapsedUs,
})
})

app.Post("/api/track", func(c *fiber.Ctx) error {
var payload trackPayload
if err := json.Unmarshal(c.Body(), &payload); err != nil {
return c.Status(400).JSON(fiber.Map{"error": "Geçersiz telemetri"})
}
select {
case store.eventChan <- payload:
default:
store.applyEvent(payload)
}
return c.SendStatus(202)
})

app.Post("/api/webhooks/payment", func(c *fiber.Ctx) error {
var payload struct {
Slug         string `json:"slug"`
Variant      string `json:"variant"`
UtmSource    string `json:"utmSource"`
RevenueCents int64  `json:"revenueCents"`
}
if err := c.BodyParser(&payload); err != nil {
return c.Status(400).JSON(fiber.Map{"error": "Geçersiz ödeme webhook verisi"})
}
if payload.RevenueCents <= 0 {
payload.RevenueCents = 8900
}
store.applyEvent(trackPayload{
Slug:         payload.Slug,
Variant:      payload.Variant,
Event:        "purchase",
UtmSource:    payload.UtmSource,
RevenueCents: payload.RevenueCents,
})
return c.JSON(fiber.Map{"status": "recorded", "revenueCents": payload.RevenueCents})
})

app.Get("/api/analytics/:slug", func(c *fiber.Ctx) error {
return c.JSON(store.GetAnalytics(c.Params("slug")))
})

return app
}

func main() {
store := NewStore()
app := setupApp(store)
port := os.Getenv("PORT")
if port == "" {
port = "8080"
}
log.Printf("FunnelFox V2 Enterprise Engine http://localhost:%s üzerinde hazır!", port)
log.Fatal(app.Listen(":" + port))
}