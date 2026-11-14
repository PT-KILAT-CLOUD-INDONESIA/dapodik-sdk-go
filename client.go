package dapodik

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version adalah versi SDK Go ini, mengikuti versi Python yang di-port
// (pyproject.toml v1.1.0; __version__ Python masih 1.0.0 — lihat README).
const Version = "1.1.0"

// UserAgent adalah nilai header User-Agent yang dikirim setiap request.
const UserAgent = "dapodik-sdk-go/1.1.0"

// DefaultHost, DefaultPort dan DefaultTimeout adalah default koneksi
// WebService Dapodik, sama seperti Python.
const (
	DefaultHost    = "127.0.0.1"
	DefaultPort    = 5774
	DefaultTimeout = 30 * time.Second
)

// Client adalah client untuk WebService Dapodik Kemendikdasmen (port 5774).
// Port dari DapodikClient di Python; memakai stdlib net/http saja
// (zero external dependencies).
type Client struct {
	// NPSN adalah Nomor Pokok Sekolah Nasional.
	NPSN string
	// Token adalah token WebService Dapodik (dari menu Web Service di
	// aplikasi Dapodik).
	Token string
	// BaseURL adalah base URL WebService, mis.
	// "http://127.0.0.1:5774/WebService".
	BaseURL string
	// Timeout adalah timeout per request.
	Timeout time.Duration
	// HTTPClient adalah client HTTP yang dipakai; jika nil, dipakai
	// http.Client dengan Timeout di atas.
	HTTPClient *http.Client

	// host, port, dan baseURL adalah state internal konstruksi yang
	// dipakai NewClient sebelum BaseURL final dihitung.
	host    string
	port    int
	baseURL string
}

// Option adalah functional option untuk NewClient.
type Option func(*Client)

// WithHost mengatur host WebService (default "127.0.0.1").
func WithHost(host string) Option {
	return func(c *Client) { c.host = host }
}

// WithPort mengatur port WebService (default 5774).
func WithPort(port int) Option {
	return func(c *Client) { c.port = port }
}

// WithBaseURL mengatur base URL secara langsung, mengesampingkan
// host/port. Trailing "/" dihapus, seperti Python (rstrip("/")).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithTimeout mengatur timeout per request (default 30 detik).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.Timeout = d }
}

// WithHTTPClient mengatur *http.Client yang dipakai untuk request.
// Berguna untuk testing (httptest) atau transport kustom. Jika di-set,
// Timeout pada http.Client tersebut yang berlaku, kecuali Timeout
// client juga di-set dan HTTPClient.Timeout masih 0.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.HTTPClient = hc }
}

// NewClient membuat Client baru. Validasi sama persis seperti Python:
// NPSN wajib diisi (tidak boleh kosong atau hanya spasi) dan tidak boleh
// mengandung \r atau \n; token wajib diisi dan tidak boleh mengandung
// \r atau \n (pencegahan CRLF injection). Nilai yang disimpan sudah
// di-trim, seperti Python.
//
// BaseURL: jika diberikan lewat WithBaseURL, trailing "/" dihapus;
// jika tidak, dibangun "http://<host>:<port>/WebService" dengan
// "http://" ditambahkan bila host belum berawalan http:// atau https://.
func NewClient(npsn, token string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(npsn) == "" {
		return nil, &DapodikError{Message: "NPSN wajib diisi"}
	}
	if strings.ContainsAny(npsn, "\r\n") {
		return nil, &DapodikError{Message: "NPSN tidak boleh mengandung karakter newline"}
	}
	if strings.TrimSpace(token) == "" {
		return nil, &DapodikError{Message: "Token WebService Dapodik wajib diisi"}
	}
	if strings.ContainsAny(token, "\r\n") {
		return nil, &DapodikError{Message: "Token tidak boleh mengandung karakter newline (CRLF injection prevention)"}
	}

	c := &Client{
		NPSN:    strings.TrimSpace(npsn),
		Token:   strings.TrimSpace(token),
		Timeout: DefaultTimeout,
		host:    DefaultHost,
		port:    DefaultPort,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.baseURL != "" {
		c.BaseURL = strings.TrimRight(c.baseURL, "/")
	} else {
		host := strings.TrimSpace(c.host)
		if host == "" {
			host = DefaultHost
		}
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "http://" + host
		}
		port := c.port
		if port == 0 {
			port = DefaultPort
		}
		c.BaseURL = fmt.Sprintf("%s:%d/WebService", host, port)
	}
	return c, nil
}

// Request mengirim HTTP request ke endpoint WebService Dapodik.
//
//   - endpoint tidak boleh mengandung ".." atau "\\" (path traversal),
//     seperti Python.
//   - Query selalu berisi npsn=<npsn> ditambah params non-nil saja
//     (di-url-encode).
//   - Header: Authorization "Bearer <token>", Accept
//     "application/json, text/plain, */*", User-Agent "dapodik-sdk-go/1.1.0".
//   - Body non-nil di-encode JSON dengan Content-Type application/json.
//   - HTTP 401/403 -> *AuthError; non-2xx lain -> *HTTPError;
//     error jaringan/timeout -> *ConnectionError (Unwrap error asli).
//   - Body respons kosong -> Response kosong. JSON tidak valid ->
//     *DapodikError.
//   - Normalisasi rows sama seperti Python: JSON list -> rows=list;
//     dict dengan "rows" object -> [rows], list -> rows; dict dengan
//     "data" list -> rows=data; dict lain -> [parsed]. Raw = parsed dict
//     atau map kosong.
func (c *Client) Request(ctx context.Context, method, endpoint string, params map[string]any, body any) (Response, error) {
	cleanEndpoint := strings.TrimLeft(endpoint, "/")
	if strings.Contains(cleanEndpoint, "..") || strings.Contains(cleanEndpoint, "\\") {
		return Response{}, &DapodikError{Message: "Endpoint tidak valid (path traversal detected)"}
	}

	query := url.Values{}
	query.Set("npsn", c.NPSN)
	for k, v := range params {
		if v == nil {
			continue
		}
		query.Set(k, paramValue(v))
	}
	reqURL := c.BaseURL + "/" + cleanEndpoint + "?" + query.Encode()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return Response{}, &DapodikError{Message: fmt.Sprintf("Gagal meng-encode body request ke JSON: %v", err)}
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), reqURL, bodyReader)
	if err != nil {
		return Response{}, &ConnectionError{Msg: fmt.Sprintf("Request timeout / connection error: %v", err), Err: err}
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: c.Timeout}
	}

	resp, err := hc.Do(req)
	if err != nil {
		return Response{}, &ConnectionError{Msg: fmt.Sprintf("Gagal terhubung ke server Dapodik (%s): %v", c.BaseURL, err), Err: err}
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, &ConnectionError{Msg: fmt.Sprintf("Request timeout / connection error: %v", err), Err: err}
	}
	rawText := string(rawBytes)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return Response{}, NewAuthError(resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, &HTTPError{
			StatusCode: resp.StatusCode,
			// StatusText adalah frasa alasan saja ("Internal Server
			// Error"), seperti e.reason di Python — bukan resp.Status
			// yang mengulang kode ("500 Internal Server Error").
			StatusText: strings.TrimPrefix(resp.Status, fmt.Sprintf("%d ", resp.StatusCode)),
			Endpoint:   cleanEndpoint,
			RawBody:    rawText,
		}
	}

	if strings.TrimSpace(rawText) == "" {
		return Response{Rows: []map[string]any{}, Raw: map[string]any{}}, nil
	}

	var parsed any
	if err := json.Unmarshal(rawBytes, &parsed); err != nil {
		snip := rawText
		if len(snip) > 200 {
			snip = snip[:200]
		}
		return Response{}, &DapodikError{Message: fmt.Sprintf("Respons bukan JSON yang valid dari Dapodik: %s", snip)}
	}

	return normalizeResponse(parsed), nil
}

// normalizeResponse menormalisasi hasil parse JSON menjadi Response,
// mengikuti logika Python di request().
func normalizeResponse(parsed any) Response {
	rows := make([]map[string]any, 0)
	var raw map[string]any
	switch p := parsed.(type) {
	case []any:
		for _, item := range p {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, m)
			}
		}
		raw = map[string]any{}
	case map[string]any:
		raw = p
		if v, ok := p["rows"]; ok {
			switch r := v.(type) {
			case map[string]any:
				rows = append(rows, r)
			case []any:
				for _, item := range r {
					if m, ok := item.(map[string]any); ok {
						rows = append(rows, m)
					}
				}
			}
		} else if v, ok := p["data"]; ok {
			if list, ok := v.([]any); ok {
				for _, item := range list {
					if m, ok := item.(map[string]any); ok {
						rows = append(rows, m)
					}
				}
			}
		} else {
			rows = append(rows, p)
		}
	default:
		raw = map[string]any{}
	}
	return Response{Rows: rows, Raw: raw}
}

// paramValue mengubah nilai param query menjadi string seperti yang
// dilakukan Python (urllib.parse.urlencode memanggil str()): bool
// menjadi "True"/"False", sisanya fmt.Sprint.
func paramValue(v any) string {
	switch b := v.(type) {
	case bool:
		if b {
			return "True"
		}
		return "False"
	case string:
		return b
	default:
		return fmt.Sprint(v)
	}
}

// mergeParams menggabungkan beberapa map params (variadic) menjadi satu.
// Map yang disebut belakangan menang jika ada key yang sama.
// Sama seperti **params di Python.
func mergeParams(maps ...map[string]any) map[string]any {
	out := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// pageParam menormalisasi page seperti Python (default 1) dengan tambahan:
// page <= 0 dianggap 1.
func pageParam(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// pagedParams membangun map query untuk endpoint berpaginasi: page
// selalu dikirim, limit hanya dikirim bila > 0 (seperti Python yang
// hanya mengirim limit bila truthy).
func pagedParams(page, limit int, extra ...map[string]any) map[string]any {
	q := map[string]any{"page": pageParam(page)}
	if limit > 0 {
		q["limit"] = limit
	}
	for k, v := range mergeParams(extra...) {
		q[k] = v
	}
	return q
}

// semesterParams membangun map query untuk endpoint dengan semester_id
// opsional: semester_id hanya dikirim bila non-empty, seperti Python.
func semesterParams(semesterID string, extra ...map[string]any) map[string]any {
	q := mergeParams(extra...)
	if semesterID != "" {
		q["semester_id"] = semesterID
	}
	return q
}

// --- Endpoint GET ---

// GetSekolah menarik profil dan perizinan operasional sekolah.
func (c *Client) GetSekolah(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getSekolah", semesterParams(semesterID, params...), nil)
}

// GetPengguna menarik akun operator dan pengguna Dapodik.
func (c *Client) GetPengguna(ctx context.Context, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getPengguna", mergeParams(params...), nil)
}

// GetGTK menarik data Guru dan Tenaga Kependidikan. page <= 0 menjadi 1,
// limit <= 0 tidak dikirim (sama seperti limit=None di Python).
func (c *Client) GetGTK(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getGtk", pagedParams(page, limit, params...), nil)
}

// GetRombonganBelajar menarik data rombongan belajar (kelas) beserta
// anggota dan pembelajaran.
func (c *Client) GetRombonganBelajar(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getRombonganBelajar", semesterParams(semesterID, params...), nil)
}

// GetPesertaDidik menarik data siswa / peserta didik.
func (c *Client) GetPesertaDidik(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getPesertaDidik", pagedParams(page, limit, params...), nil)
}

// GetMataPelajaran menarik data referensi mata pelajaran nasional.
func (c *Client) GetMataPelajaran(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getMataPelajaran", semesterParams(semesterID, params...), nil)
}

// GetMatevNilai menarik referensi mata evaluasi nilai.
func (c *Client) GetMatevNilai(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getMatevNilai", semesterParams(semesterID, params...), nil)
}

// GetPrasarana menarik data prasarana (tanah, bangunan, dan ruang sekolah).
func (c *Client) GetPrasarana(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "GET", "getPrasarana", pagedParams(page, limit, params...), nil)
}

// --- Endpoint POST ---

// Post mengirim data via HTTP POST ke WebService Dapodik.
func (c *Client) Post(ctx context.Context, endpoint string, body any, params ...map[string]any) (Response, error) {
	return c.Request(ctx, "POST", endpoint, mergeParams(params...), body)
}

// PostMatevRapor mengirim data mata evaluasi rapor.
func (c *Client) PostMatevRapor(ctx context.Context, body any, params ...map[string]any) (Response, error) {
	return c.Post(ctx, "postMatevRapor", body, params...)
}

// PostNilai mengirim data nilai rapor. table default "rapor" bila kosong,
// dikirim sebagai query param table=<table>, seperti Python.
func (c *Client) PostNilai(ctx context.Context, body any, table string, params ...map[string]any) (Response, error) {
	if table == "" {
		table = "rapor"
	}
	q := mergeParams(params...)
	q["table"] = table
	return c.Post(ctx, "postNilai", body, q)
}

// --- Auto-pagination & streaming ---

// ProgressFunc dipanggil setiap halaman berhasil ditarik oleh FetchAll*:
// page = nomor halaman, count = jumlah baris halaman ini,
// total = total baris yang sudah terkumpul sejauh ini.
// Sama seperti on_progress(page, count, total) di Python.
type ProgressFunc func(page, count, total int)

// fetchAll adalah helper bersama untuk FetchAll*: menarik halaman demi
// halaman sampai halaman kosong atau jumlah baris < limit. delay
// diterapkan antar halaman (setelah halaman penuh), seperti Python.
func (c *Client) fetchAll(ctx context.Context, fetch func(ctx context.Context, page, limit int) (Response, error), limit int, delay time.Duration, onProgress ProgressFunc) (Response, error) {
	if limit <= 0 {
		limit = 100
	}
	allRows := make([]map[string]any, 0)
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return Response{}, &ConnectionError{Msg: fmt.Sprintf("Request timeout / connection error: %v", err), Err: err}
		}
		resp, err := fetch(ctx, page, limit)
		if err != nil {
			return Response{}, err
		}
		count := len(resp.Rows)
		if count == 0 {
			break
		}
		allRows = append(allRows, resp.Rows...)
		if onProgress != nil {
			onProgress(page, count, len(allRows))
		}
		if count < limit {
			break
		}
		page++
		if delay > 0 {
			select {
			case <-ctx.Done():
				return Response{}, &ConnectionError{Msg: fmt.Sprintf("Request timeout / connection error: %v", ctx.Err()), Err: ctx.Err()}
			case <-time.After(delay):
			}
		}
	}
	return Response{Rows: allRows, Raw: map[string]any{}}, nil
}

// FetchAllPesertaDidik menarik seluruh data siswa secara otomatis
// melintasi banyak halaman.
func (c *Client) FetchAllPesertaDidik(ctx context.Context, limit int, delay time.Duration, onProgress ProgressFunc, params ...map[string]any) (Response, error) {
	return c.fetchAll(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetPesertaDidik(ctx, page, limit, params...)
	}, limit, delay, onProgress)
}

// FetchAllGTK menarik seluruh data GTK secara otomatis melintasi banyak halaman.
func (c *Client) FetchAllGTK(ctx context.Context, limit int, delay time.Duration, onProgress ProgressFunc, params ...map[string]any) (Response, error) {
	return c.fetchAll(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetGTK(ctx, page, limit, params...)
	}, limit, delay, onProgress)
}

// FetchAllPrasarana menarik seluruh data prasarana secara otomatis
// melintasi banyak halaman.
func (c *Client) FetchAllPrasarana(ctx context.Context, limit int, delay time.Duration, onProgress ProgressFunc, params ...map[string]any) (Response, error) {
	return c.fetchAll(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetPrasarana(ctx, page, limit, params...)
	}, limit, delay, onProgress)
}

// iterate adalah helper bersama untuk Iterate*: mengembalikan iterator
// range-over-func (Go 1.23) yang menghasilkan batch per halaman.
// yield menerima (batch, err): err non-nil hanya bila fetch halaman
// gagal atau context dibatalkan; iterasi berhenti setelah itu.
//
// Dipilih pola range-over-func (bukan channel) karena: tidak ada goroutine
// bocor, context dibatalkan secara natural, dan pemakaian mengikuti
// idiom `for batch, err := range client.IteratePesertaDidik(...)`.
func (c *Client) iterate(ctx context.Context, fetch func(ctx context.Context, page, limit int) (Response, error), limit int) func(yield func([]map[string]any, error) bool) {
	if limit <= 0 {
		limit = 100
	}
	return func(yield func([]map[string]any, error) bool) {
		page := 1
		for {
			if err := ctx.Err(); err != nil {
				yield(nil, &ConnectionError{Msg: fmt.Sprintf("Request timeout / connection error: %v", err), Err: err})
				return
			}
			resp, err := fetch(ctx, page, limit)
			if err != nil {
				yield(nil, err)
				return
			}
			if len(resp.Rows) == 0 {
				return
			}
			if !yield(resp.Rows, nil) {
				return
			}
			if len(resp.Rows) < limit {
				return
			}
			page++
		}
	}
}

// IteratePesertaDidik adalah generator stream per-batch halaman siswa
// untuk efisiensi RAM, seperti iterate_peserta_didik di Python.
//
//	for batch, err := range client.IteratePesertaDidik(ctx, 100) {
//		if err != nil { ... }
//		// proses batch
//	}
func (c *Client) IteratePesertaDidik(ctx context.Context, limit int, params ...map[string]any) func(yield func([]map[string]any, error) bool) {
	return c.iterate(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetPesertaDidik(ctx, page, limit, params...)
	}, limit)
}

// IterateGTK adalah generator stream per-batch halaman GTK.
// CATATAN: Python CHANGELOG 1.0.0 mendokumentasikan iterate_gtk, tetapi
// client.py Python TIDAK mengimplementasikannya — ini adalah celah yang
// diperbaiki di port Go ini (lihat README/CHANGELOG).
func (c *Client) IterateGTK(ctx context.Context, limit int, params ...map[string]any) func(yield func([]map[string]any, error) bool) {
	return c.iterate(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetGTK(ctx, page, limit, params...)
	}, limit)
}

// IteratePrasarana adalah generator stream per-batch halaman prasarana,
// seperti iterate_prasarana di Python.
func (c *Client) IteratePrasarana(ctx context.Context, limit int, params ...map[string]any) func(yield func([]map[string]any, error) bool) {
	return c.iterate(ctx, func(ctx context.Context, page, limit int) (Response, error) {
		return c.GetPrasarana(ctx, page, limit, params...)
	}, limit)
}

// --- Aliases (nama sama seperti Python) ---

// Sekolah adalah alias GetSekolah.
func (c *Client) Sekolah(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.GetSekolah(ctx, semesterID, params...)
}

// Pengguna adalah alias GetPengguna.
func (c *Client) Pengguna(ctx context.Context, params ...map[string]any) (Response, error) {
	return c.GetPengguna(ctx, params...)
}

// GTK adalah alias GetGTK.
func (c *Client) GTK(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.GetGTK(ctx, page, limit, params...)
}

// Rombel adalah alias GetRombonganBelajar.
func (c *Client) Rombel(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.GetRombonganBelajar(ctx, semesterID, params...)
}

// PD adalah alias GetPesertaDidik.
func (c *Client) PD(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.GetPesertaDidik(ctx, page, limit, params...)
}

// Prasarana adalah alias GetPrasarana.
func (c *Client) Prasarana(ctx context.Context, page, limit int, params ...map[string]any) (Response, error) {
	return c.GetPrasarana(ctx, page, limit, params...)
}

// MataPelajaran adalah alias GetMataPelajaran.
func (c *Client) MataPelajaran(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.GetMataPelajaran(ctx, semesterID, params...)
}

// MatevNilai adalah alias GetMatevNilai.
func (c *Client) MatevNilai(ctx context.Context, semesterID string, params ...map[string]any) (Response, error) {
	return c.GetMatevNilai(ctx, semesterID, params...)
}
