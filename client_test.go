package dapodik

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	testNPSN  = "20300001"
	testToken = "secret-token-12345"
)

func mustClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(testNPSN, testToken, WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func jsonHandler(t *testing.T, status int, payload any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name  string
		npsn  string
		token string
	}{
		{"empty npsn", "", "tok"},
		{"whitespace npsn", "   ", "tok"},
		{"empty token", testNPSN, ""},
		{"whitespace token", testNPSN, "   "},
		{"newline npsn", "2030\n0001", "tok"},
		{"carriage return npsn", "2030\r0001", "tok"},
		{"newline token", testNPSN, "tok\nx"},
		{"carriage return token", testNPSN, "tok\rx"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.npsn, tc.token)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var derr *DapodikError
			if !errors.As(err, &derr) {
				t.Fatalf("expected *DapodikError, got %T: %v", err, err)
			}
		})
	}
}

func TestBaseURLConstruction(t *testing.T) {
	c, err := NewClient(testNPSN, testToken, WithHost("192.168.1.50"), WithPort(5774))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.BaseURL != "http://192.168.1.50:5774/WebService" {
		t.Fatalf("BaseURL = %q", c.BaseURL)
	}
	if c.NPSN != testNPSN || c.Token != testToken {
		t.Fatalf("NPSN/Token not stored: %q %q", c.NPSN, c.Token)
	}
	// Host yang sudah ber-skema tidak ditambahi http:// lagi.
	c2, err := NewClient(testNPSN, testToken, WithHost("https://dapodik.local"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c2.BaseURL != "https://dapodik.local:5774/WebService" {
		t.Fatalf("BaseURL = %q", c2.BaseURL)
	}
	// BaseURL override: trailing "/" dihapus.
	c3, err := NewClient(testNPSN, testToken, WithBaseURL("http://example.com/api/"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c3.BaseURL != "http://example.com/api" {
		t.Fatalf("BaseURL = %q", c3.BaseURL)
	}
}

func TestGetSekolahSuccess(t *testing.T) {
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{
		"status": "success",
		"rows":   []any{map[string]any{"sekolah_id": "s1", "nama": "SMA Negeri 1 Gedeg", "npsn": testNPSN}},
	}))
	defer ts.Close()

	res, err := mustClient(t, ts.URL).GetSekolah(context.Background(), "")
	if err != nil {
		t.Fatalf("GetSekolah: %v", err)
	}
	if res.Len() != 1 || res.Count() != 1 {
		t.Fatalf("Len/Count = %d/%d, want 1/1", res.Len(), res.Count())
	}
	if res.First()["nama"] != "SMA Negeri 1 Gedeg" {
		t.Fatalf("First nama = %v", res.First()["nama"])
	}
	if res.Raw["status"] != "success" {
		t.Fatalf("Raw status = %v", res.Raw["status"])
	}
}

func TestGetSekolahSingleObjectNormalization(t *testing.T) {
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{
		"status": "success",
		"rows":   map[string]any{"sekolah_id": "s1", "nama": "SMA Negeri 1 Gedeg"},
	}))
	defer ts.Close()

	res, err := mustClient(t, ts.URL).Sekolah(context.Background(), "")
	if err != nil {
		t.Fatalf("Sekolah: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("len = %d, want 1", len(res.Rows))
	}
	if res.First()["nama"] != "SMA Negeri 1 Gedeg" {
		t.Fatalf("First nama = %v", res.First()["nama"])
	}
}

func TestDataKeyNormalizationAndListResponse(t *testing.T) {
	// dict dengan "data" list -> rows=data
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{
		"data": []any{map[string]any{"nama": "A"}},
	}))
	defer ts.Close()
	res, err := mustClient(t, ts.URL).GetPengguna(context.Background())
	if err != nil {
		t.Fatalf("GetPengguna: %v", err)
	}
	if len(res.Rows) != 1 || res.Rows[0]["nama"] != "A" {
		t.Fatalf("rows = %v", res.Rows)
	}

	// JSON list langsung -> rows=list, Raw kosong
	ts2 := httptest.NewServer(jsonHandler(t, 200, []any{map[string]any{"nama": "B"}}))
	defer ts2.Close()
	res2, err := mustClient(t, ts2.URL).GetPengguna(context.Background())
	if err != nil {
		t.Fatalf("GetPengguna list: %v", err)
	}
	if len(res2.Rows) != 1 || res2.Rows[0]["nama"] != "B" {
		t.Fatalf("rows = %v", res2.Rows)
	}
	if len(res2.Raw) != 0 {
		t.Fatalf("Raw should be empty for list response, got %v", res2.Raw)
	}
}

func TestPDAliasFilterPluck(t *testing.T) {
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{
		"status": "success",
		"rows": []any{
			map[string]any{"peserta_didik_id": "p1", "nama": "Budi", "jenis_kelamin": "L"},
			map[string]any{"peserta_didik_id": "p2", "nama": "Siti", "jenis_kelamin": "P"},
		},
	}))
	defer ts.Close()

	res, err := mustClient(t, ts.URL).PD(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("PD: %v", err)
	}
	if len(res.Rows) != 2 {
		t.Fatalf("len = %d, want 2", len(res.Rows))
	}
	if got := res.Pluck("nama"); len(got) != 2 || got[0] != "Budi" || got[1] != "Siti" {
		t.Fatalf("Pluck = %v", got)
	}
	filtered := res.Filter(map[string]any{"jenis_kelamin": "L"})
	if len(filtered.Rows) != 1 || filtered.First()["nama"] != "Budi" {
		t.Fatalf("Filter = %v", filtered.Rows)
	}
	// Pluck key yang tidak ada di baris manapun -> kosong.
	if got := res.Pluck("tidak_ada"); len(got) != 0 {
		t.Fatalf("Pluck missing key = %v", got)
	}
	// First pada Response kosong -> nil.
	if (Response{}).First() != nil {
		t.Fatal("First on empty Response should be nil")
	}
}

func TestRequestHeadersAndQuery(t *testing.T) {
	var gotAuth, gotAccept, gotUA, gotNPSN, gotPage string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotUA = r.Header.Get("User-Agent")
		gotNPSN = r.URL.Query().Get("npsn")
		gotPage = r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer ts.Close()

	_, err := mustClient(t, ts.URL).GetGTK(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("GetGTK: %v", err)
	}
	if gotAuth != "Bearer "+testToken {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotAccept != "application/json, text/plain, */*" {
		t.Fatalf("Accept = %q", gotAccept)
	}
	if gotUA != "dapodik-sdk-go/1.1.0" {
		t.Fatalf("User-Agent = %q", gotUA)
	}
	if gotNPSN != testNPSN {
		t.Fatalf("npsn query = %q", gotNPSN)
	}
	if gotPage != "1" {
		t.Fatalf("page query = %q, want 1", gotPage)
	}
}

func TestAuthErrorOn401And403(t *testing.T) {
	for _, code := range []int{401, 403} {
		ts := httptest.NewServer(jsonHandler(t, code, map[string]any{"message": "unauthorized"}))
		_, err := mustClient(t, ts.URL).GetSekolah(context.Background(), "")
		ts.Close()
		if err == nil {
			t.Fatalf("code %d: expected error", code)
		}
		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Fatalf("code %d: expected *AuthError, got %T: %v", code, err, err)
		}
		if authErr.StatusCode != code {
			t.Fatalf("code %d: StatusCode = %d", code, authErr.StatusCode)
		}
		var derr *DapodikError
		if !errors.As(err, &derr) {
			t.Fatalf("code %d: AuthError should unwrap to *DapodikError", code)
		}
	}
}

func TestHTTPErrorOn500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("boom"))
	}))
	defer ts.Close()

	_, err := mustClient(t, ts.URL).GetGTK(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != 500 || httpErr.Endpoint != "getGtk" || httpErr.RawBody != "boom" {
		t.Fatalf("HTTPError = %+v", httpErr)
	}
	want := "Dapodik WebService HTTP 500 (Internal Server Error) saat memanggil endpoint 'getGtk'"
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestPostNilaiTableParam(t *testing.T) {
	var gotTable, gotSemester, gotBody string
	var gotCT string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTable = r.URL.Query().Get("table")
		gotSemester = r.URL.Query().Get("semester_id")
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","message":"Tersimpan"}`))
	}))
	defer ts.Close()

	// Default table = "rapor", semester_id dari params.
	res, err := mustClient(t, ts.URL).PostNilai(context.Background(),
		[]any{map[string]any{"nilai": 90}}, "",
		map[string]any{"semester_id": "20241"})
	if err != nil {
		t.Fatalf("PostNilai: %v", err)
	}
	if gotTable != "rapor" {
		t.Fatalf("table = %q, want rapor", gotTable)
	}
	if gotSemester != "20241" {
		t.Fatalf("semester_id = %q", gotSemester)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type = %q", gotCT)
	}
	if !strings.Contains(gotBody, `"nilai":90`) {
		t.Fatalf("body = %s", gotBody)
	}
	if res.First()["message"] != "Tersimpan" {
		t.Fatalf("First message = %v", res.First()["message"])
	}

	// Table kustom mengesampingkan default.
	_, err = mustClient(t, ts.URL).PostNilai(context.Background(),
		[]any{map[string]any{"nilai": 80}}, "nilai_akhir")
	if err != nil {
		t.Fatalf("PostNilai custom table: %v", err)
	}
	if gotTable != "nilai_akhir" {
		t.Fatalf("table = %q, want nilai_akhir", gotTable)
	}
}

func TestPostMatevRapor(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer ts.Close()
	_, err := mustClient(t, ts.URL).PostMatevRapor(context.Background(), map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("PostMatevRapor: %v", err)
	}
	if gotPath != "/postMatevRapor" {
		t.Fatalf("path = %q", gotPath)
	}
}

// pagedHandler mengembalikan rows berdasarkan query page/limit.
func pagedHandler(pages map[string][]map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		rows := pages[page]
		if rows == nil {
			rows = []map[string]any{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "rows": rows})
	}
}

func TestFetchAllPesertaDidik(t *testing.T) {
	ts := httptest.NewServer(pagedHandler(map[string][]map[string]any{
		"1": {{"nama": "A"}, {"nama": "B"}},
		"2": {{"nama": "C"}},
	}))
	defer ts.Close()

	var progress [][3]int
	all, err := mustClient(t, ts.URL).FetchAllPesertaDidik(context.Background(), 2, 0,
		func(page, count, total int) { progress = append(progress, [3]int{page, count, total}) })
	if err != nil {
		t.Fatalf("FetchAllPesertaDidik: %v", err)
	}
	if len(all.Rows) != 3 {
		t.Fatalf("len = %d, want 3", len(all.Rows))
	}
	if got := all.Pluck("nama"); fmt.Sprint(got) != "[A B C]" {
		t.Fatalf("Pluck = %v", got)
	}
	if len(progress) != 2 || progress[0] != [3]int{1, 2, 2} || progress[1] != [3]int{2, 1, 3} {
		t.Fatalf("progress = %v", progress)
	}
}

func TestFetchAllGTK(t *testing.T) {
	ts := httptest.NewServer(pagedHandler(map[string][]map[string]any{
		"1": {{"nama": "Guru A"}, {"nama": "Guru B"}},
		"2": {},
	}))
	defer ts.Close()

	all, err := mustClient(t, ts.URL).FetchAllGTK(context.Background(), 2, 0, nil)
	if err != nil {
		t.Fatalf("FetchAllGTK: %v", err)
	}
	if len(all.Rows) != 2 {
		t.Fatalf("len = %d, want 2", len(all.Rows))
	}
}

func TestFetchAllPrasaranaStopsOnShortPage(t *testing.T) {
	ts := httptest.NewServer(pagedHandler(map[string][]map[string]any{
		"1": {{"nama": "Tanah"}},
	}))
	defer ts.Close()
	all, err := mustClient(t, ts.URL).FetchAllPrasarana(context.Background(), 100, 0, nil)
	if err != nil {
		t.Fatalf("FetchAllPrasarana: %v", err)
	}
	if len(all.Rows) != 1 {
		t.Fatalf("len = %d, want 1", len(all.Rows))
	}
}

func TestIteratePesertaDidikAndGTK(t *testing.T) {
	ts := httptest.NewServer(pagedHandler(map[string][]map[string]any{
		"1": {{"nama": "A"}, {"nama": "B"}},
		"2": {{"nama": "C"}},
	}))
	defer ts.Close()
	c := mustClient(t, ts.URL)

	var batches [][]map[string]any
	for batch, err := range c.IteratePesertaDidik(context.Background(), 2) {
		if err != nil {
			t.Fatalf("IteratePesertaDidik: %v", err)
		}
		batches = append(batches, batch)
	}
	if len(batches) != 2 || len(batches[0]) != 2 || len(batches[1]) != 1 {
		t.Fatalf("batches = %v", batches)
	}

	// IterateGTK (celah Python yang diperbaiki di Go): pola yang sama.
	var gtkBatches [][]map[string]any
	for batch, err := range c.IterateGTK(context.Background(), 2) {
		if err != nil {
			t.Fatalf("IterateGTK: %v", err)
		}
		gtkBatches = append(gtkBatches, batch)
	}
	if len(gtkBatches) != 2 {
		t.Fatalf("gtk batches = %v", gtkBatches)
	}

	// Early break: consumer berhenti setelah batch pertama.
	n := 0
	for _, err := range c.IteratePrasarana(context.Background(), 2) {
		if err != nil {
			t.Fatalf("IteratePrasarana: %v", err)
		}
		n++
		break
	}
	if n != 1 {
		t.Fatalf("early break batches = %d", n)
	}
}

func TestGetPrasaranaAndAlias(t *testing.T) {
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{
		"status":  "success",
		"results": 1,
		"rows": []any{
			map[string]any{"id_tanah": "t-1", "nama": "Tanah Sekolah",
				"bangunan": []any{map[string]any{"id_bangunan": "b-1", "nama": "Gedung A"}}},
		},
	}))
	defer ts.Close()
	c := mustClient(t, ts.URL)

	res, err := c.GetPrasarana(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("GetPrasarana: %v", err)
	}
	if len(res.Rows) != 1 || res.First()["nama"] != "Tanah Sekolah" {
		t.Fatalf("rows = %v", res.Rows)
	}
	alias, err := c.Prasarana(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("Prasarana alias: %v", err)
	}
	if alias.First()["id_tanah"] != "t-1" {
		t.Fatalf("alias first = %v", alias.First())
	}
}

func TestPathTraversalRejection(t *testing.T) {
	c := mustClient(t, "http://127.0.0.1:5774")
	for _, ep := range []string{"../etc/passwd", "..\\windows", "getSekolah/../../x", "/..%2f"} {
		_, err := c.Request(context.Background(), "GET", ep, nil, nil)
		if err == nil {
			t.Fatalf("endpoint %q: expected error", ep)
		}
		var derr *DapodikError
		if !errors.As(err, &derr) {
			t.Fatalf("endpoint %q: expected *DapodikError, got %T", ep, err)
		}
	}
	// Endpoint dengan leading "/" tetap boleh (di-trim seperti Python).
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{"rows": []any{}}))
	defer ts.Close()
	if _, err := mustClient(t, ts.URL).Request(context.Background(), "GET", "/getSekolah", nil, nil); err != nil {
		t.Fatalf("leading slash endpoint should work: %v", err)
	}
}

func TestEmptyBodyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()
	res, err := mustClient(t, ts.URL).GetSekolah(context.Background(), "")
	if err != nil {
		t.Fatalf("GetSekolah empty body: %v", err)
	}
	if len(res.Rows) != 0 || len(res.Raw) != 0 {
		t.Fatalf("empty response = %+v", res)
	}
}

func TestInvalidJSONError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>bukan json</html>"))
	}))
	defer ts.Close()
	_, err := mustClient(t, ts.URL).GetSekolah(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	var derr *DapodikError
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DapodikError, got %T: %v", err, err)
	}
}

func TestConnectionError(t *testing.T) {
	ts := httptest.NewServer(jsonHandler(t, 200, map[string]any{"rows": []any{}}))
	u := ts.URL
	ts.Close() // server mati -> koneksi gagal
	_, err := mustClient(t, u).GetSekolah(context.Background(), "")
	if err == nil {
		t.Fatal("expected connection error")
	}
	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *ConnectionError, got %T: %v", err, err)
	}
	if connErr.Unwrap() == nil {
		t.Fatal("ConnectionError.Unwrap should return the original error")
	}
}

func TestFactory(t *testing.T) {
	d := NewDapodik("127.0.0.1", 5774, 0)
	if d.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %v, want %v", d.Timeout, DefaultTimeout)
	}
	client, err := d.API("tok", testNPSN)
	if err != nil {
		t.Fatalf("API: %v", err)
	}
	if client.NPSN != testNPSN {
		t.Fatalf("NPSN = %q", client.NPSN)
	}
	if client.BaseURL != "http://127.0.0.1:5774/WebService" {
		t.Fatalf("BaseURL = %q", client.BaseURL)
	}
	if _, err := d.API("", testNPSN); err == nil {
		t.Fatal("API with empty token should error")
	}
}

func TestResponseToJSONAndCSV(t *testing.T) {
	res := Response{Rows: []map[string]any{
		{"nama": "Budi", "jenis_kelamin": "L"},
		{"nama": "Siti", "jenis_kelamin": "P", "umur": float64(17)},
	}}
	b, err := res.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil || len(decoded) != 2 {
		t.Fatalf("ToJSON round-trip failed: %s / %v", b, err)
	}
	var buf bytes.Buffer
	if err := res.ToCSV(&buf); err != nil {
		t.Fatalf("ToCSV: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("CSV lines = %q", buf.String())
	}
	if lines[0] != "nama,jenis_kelamin,umur" {
		t.Fatalf("CSV header = %q", lines[0])
	}
	if lines[1] != "Budi,L," {
		t.Fatalf("CSV row 1 = %q", lines[1])
	}
	if lines[2] != "Siti,P,17" {
		t.Fatalf("CSV row 2 = %q", lines[2])
	}
	// ToList mengembalikan salinan.
	list := res.ToList()
	if len(list) != 2 {
		t.Fatalf("ToList len = %d", len(list))
	}
}

func TestSemesterAndPaginationParams(t *testing.T) {
	var gotQ url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer ts.Close()
	c := mustClient(t, ts.URL)

	// semester_id hanya dikirim bila non-empty.
	if _, err := c.GetSekolah(context.Background(), ""); err != nil {
		t.Fatalf("GetSekolah: %v", err)
	}
	if gotQ.Get("semester_id") != "" {
		t.Fatalf("semester_id should be omitted, got %q", gotQ.Get("semester_id"))
	}
	if _, err := c.GetSekolah(context.Background(), "20241"); err != nil {
		t.Fatalf("GetSekolah: %v", err)
	}
	if gotQ.Get("semester_id") != "20241" {
		t.Fatalf("semester_id = %q", gotQ.Get("semester_id"))
	}
	// page <= 0 -> 1, limit <= 0 -> tidak dikirim.
	if _, err := c.GetGTK(context.Background(), 0, 0); err != nil {
		t.Fatalf("GetGTK: %v", err)
	}
	if gotQ.Get("page") != "1" || gotQ.Get("limit") != "" {
		t.Fatalf("page/limit = %q/%q", gotQ.Get("page"), gotQ.Get("limit"))
	}
	// Context timeout -> ConnectionError.
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	_, err := c.GetSekolah(ctx, "")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *ConnectionError, got %T: %v", err, err)
	}
}
