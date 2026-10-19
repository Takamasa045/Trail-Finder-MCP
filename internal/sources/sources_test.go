package sources

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func doc(raw string) Document {
	return Document{URL: raw, RetrievedAt: "2026-10-19T03:00:00Z", Excerpts: []Excerpt{}, CheckLinks: []string{raw}, DateMentions: []string{}}
}
func TestScopeAndOldDates(t *testing.T) {
	body := `<html><title>Mountain hiking</title><meta property="article:published_time" content="2019-10-17T12:00:00+09:00"><meta property="article:modified_time" content="2026-10-19T00:00:00+09:00"><main><h2>別の山</h2><p>距離 99km 所要時間99時間 紅葉が見頃</p><h2>対象山</h2><p>距離 8km（往復）</p><p>所要時間7時間30分</p><p>2025年10月20日 通行止め</p><p>紅葉は例年10月</p><h2>隣の山</h2><p>距離999km</p></main><footer>2026 copyright</footer></html>`
	asOf, _ := time.Parse("2006-01-02", "2026-10-19")
	d := parse(doc("https://example.org/hike"), body, "対象山", asOf)
	if d.Freshness != "older_than_14_days" || d.Published.Value != "2019-10-17" || d.Modified.Value != "2026-10-19" {
		t.Fatalf("dates=%+v", d)
	}
	for _, e := range d.Excerpts {
		if strings.Contains(e.Text, "99km") || strings.Contains(e.Text, "999") {
			t.Fatalf("cross-mountain contamination: %+v", e)
		}
	}
	if len(d.Excerpts) == 0 || len(d.DateMentions) != 1 || d.DateMentions[0] != "2025年10月20日" {
		t.Fatalf("evidence=%+v", d)
	}
}

func TestUndatedAndURLDates(t *testing.T) {
	asOf, _ := time.Parse("2006-01-02", "2026-10-19")
	for _, tc := range []struct{ url, body, expect string }{
		{"https://example.org/mountain", `<title>対象山</title><main><p>紅葉の季節に人気</p></main><footer>2026-10-19</footer>`, "undated"},
		{"https://example.org/2019/10/17/old/", `<title>対象山</title><p>台風による通行止め</p>`, "older_than_14_days"},
		{"https://example.org/2026/10/20/new/", `<title>対象山</title><p>紅葉</p>`, "future_document"},
		{"https://example.org/news", `<title>対象山</title><script type="application/ld+json">{"datePublished":"2026-10-18T10:00:00+09:00"}</script><p>紅葉の様子</p>`, "within_14_days"},
	} {
		d := parse(doc(tc.url), tc.body, "対象山", asOf)
		if d.Freshness != tc.expect {
			t.Fatalf("%s got %+v", tc.url, d)
		}
	}
	d := parse(doc("https://example.org/other"), `<title>別の山</title><p>距離10km</p>`, "対象山", asOf)
	if len(d.Excerpts) != 0 || d.Scope != "subject_not_located" {
		t.Fatalf("unrelated page=%+v", d)
	}
}

func TestPublicURLValidation(t *testing.T) {
	for _, s := range []string{"file:///etc/passwd", "http://localhost/x", "http://127.0.0.1/", "http://10.1.2.3/", "https://user:pass@example.com/", "https://example.org:8080/", "http://100.64.0.1/", "http://[::1]/", "https://host.internal/"} {
		if ValidateURL(s) == nil {
			t.Errorf("accepted %s", s)
		}
	}
	for _, s := range []string{"https://example.org/", "http://example.org/hike", "https://example.org:443/hike#map"} {
		if err := ValidateURL(s); err != nil {
			t.Errorf("%s: %v", s, err)
		}
	}
	for _, s := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "100.64.0.1", "::1", "fc00::1"} {
		if publicIP(net.ParseIP(s)) {
			t.Errorf("private address accepted: %s", s)
		}
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestFetchFailuresAndBounds(t *testing.T) {
	old := sourceHTTP
	t.Cleanup(func() { sourceHTTP = old })
	for _, tc := range []struct {
		status              int
		ctype, body, expect string
	}{
		{503, "text/html", "busy", "HTTP 503"}, {200, "application/pdf", "%PDF", "unsupported"},
		{200, "text/html", strings.Repeat("x", maxBody+1), "exceeds"}, {200, "text/html", "\xff", "UTF-8"},
	} {
		sourceHTTP = &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": []string{tc.ctype}}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}, nil
		})}
		d := Fetch(context.Background(), "https://example.org/", "対象山", time.Now())
		if d.Status != "unavailable" || !strings.Contains(d.Error, tc.expect) {
			t.Fatalf("%+v", d)
		}
	}
}

func TestPublicationTimezoneAndItineraryContext(t *testing.T) {
	t.Setenv("DEFAULT_TZ", "Asia/Tokyo")
	body := `<title>対象山</title><meta property="article:published_time" content="2026-10-07T15:00:00Z"><main><h2>対象山</h2><p>都心から50km</p><p>ブナが分布する</p><!-- <p>99時間 紅葉が見頃</p> --><div>登山口</div><div>約120分</div><div>約100分</div><div>中間点</div><div>約80分</div><div>約70分</div><div>最後の区間</div><div>約40分</div><div>約40分</div></main>`
	asOf, _ := time.Parse("2006-01-02", "2026-10-19")
	d := parse(doc("https://example.org/2026/10/08/post/"), body, "対象山", asOf)
	if d.Published.Value != "2026-10-08" || d.Published.Raw != "2026-10-07T15:00:00Z" {
		t.Fatalf("publication=%+v", d.Published)
	}
	var times []string
	for _, e := range d.Excerpts {
		if e.Topic == "distance" || e.Topic == "foliage" {
			t.Errorf("irrelevant/comment evidence: %+v", e)
		}
		if e.Topic == "duration" {
			times = append(times, e.Text)
		}
	}
	if len(times) != 3 || !strings.Contains(times[0], "約120分 | 約100分") || !strings.Contains(times[2], "約40分 | 約40分") {
		t.Fatalf("lost itinerary context: %+v", times)
	}
}

func TestRelatedDatesNeverBecomeArticleDateOrEvidence(t *testing.T) {
	body := `<title>対象山の情報</title><article><div class="entry-content"><div><p>紅葉は例年秋です。</p></div></div><section><h2>関連記事</h2><time datetime="2026-10-19">2026年10月19日</time><p>現在通行止め、紅葉ピーク</p></section></article>`
	asOf, _ := time.Parse("2006-01-02", "2026-10-19")
	d := parse(doc("https://example.org/mountain"), body, "対象山", asOf)
	if d.Published != nil || d.Freshness != "undated" || len(d.DateMentions) != 0 {
		t.Fatalf("related date used: %+v", d)
	}
	for _, e := range d.Excerpts {
		if strings.Contains(e.Text, "通行止め") || strings.Contains(e.Text, "ピーク") {
			t.Fatalf("related claim used: %+v", e)
		}
	}
}

func TestLiteralHeadingMarkersAndMalformedContainer(t *testing.T) {
	asOf, _ := time.Parse("2006-01-02", "2026-10-19")
	d := parse(doc("https://example.org/mountain"), `<title>対象山</title><h2>対象山</h2><p>@@H</p><p>所要時間2時間</p>`, "対象山", asOf)
	if len(d.Excerpts) == 0 {
		t.Fatal("valid article evidence lost")
	}
	if _, ok := contentElement(strings.Repeat(`<div class="entry-content">`, 1000)); ok {
		t.Fatal("unterminated container accepted")
	}
}
