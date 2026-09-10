package geo

import "testing"

func TestDistanceMetersTakao(t *testing.T) {
	// 高尾山口付近 → 高尾山頂付近（おおよそ 3–6km）
	d := DistanceMeters(35.6319, 139.2694, 35.6254, 139.2435)
	if d < 2000 || d > 8000 {
		t.Fatalf("unexpected distance %.1f m", d)
	}
}

func TestSampleEveryIncludesEnds(t *testing.T) {
	pts := []Point{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 0.002},
		{Lat: 0, Lon: 0.004},
		{Lat: 0, Lon: 0.006},
		{Lat: 0, Lon: 0.008},
	}
	got := SampleEvery(pts, 300, 8)
	if got[0] != pts[0] {
		t.Fatalf("first = %+v", got[0])
	}
	if got[len(got)-1] != pts[len(pts)-1] {
		t.Fatalf("last = %+v", got[len(got)-1])
	}
}

func TestSampleEveryDensifiesLongSegment(t *testing.T) {
	// ~11 km east-west at equator
	pts := []Point{{Lat: 0, Lon: 0}, {Lat: 0, Lon: 0.1}}
	got := SampleEvery(pts, 250, 16)
	if len(got) != 16 {
		t.Fatalf("len=%d", len(got))
	}
	mid := got[len(got)/2]
	if mid.Lon < 0.04 || mid.Lon > 0.06 {
		t.Fatalf("midpoint not sampled: %+v", mid)
	}
	if DistanceMeters(got[0].Lat, got[0].Lon, got[len(got)-1].Lat, got[len(got)-1].Lon) < 10000 {
		t.Fatal("expected ~11km span")
	}
}

func TestSampleEveryEmpty(t *testing.T) {
	if SampleEvery(nil, 100, 8) != nil {
		t.Fatal("expected nil")
	}
}

func TestElevationGainLoss(t *testing.T) {
	gain, loss := ElevationGainLoss([]float64{100, 150, 120, 180})
	if gain != 110 {
		t.Fatalf("gain=%v want 110", gain)
	}
	if loss != 30 {
		t.Fatalf("loss=%v want 30", loss)
	}
}

func TestMinDistanceMeters(t *testing.T) {
	pts := []Point{{Lat: 35.0, Lon: 139.0}, {Lat: 35.1, Lon: 139.1}}
	d := MinDistanceMeters(35.0, 139.0, pts)
	if d != 0 {
		t.Fatalf("distance to first point = %v", d)
	}
	if MinDistanceMeters(0, 0, nil) != 0 {
		t.Fatal("empty points should be 0")
	}
}
