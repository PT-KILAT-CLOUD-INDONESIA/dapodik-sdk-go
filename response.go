package dapodik

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// Response adalah wrapper respons WebService Dapodik, port dari
// DapodikResponse di Python. Rows berisi baris data hasil normalisasi,
// Raw berisi respons JSON mentah apa adanya (map kosong jika respons
// berupa list atau respons kosong).
//
// Berbeda dengan Python, tidak ada padanan to_dataframe() (pandas) di Go.
// Sebagai gantinya tersedia ToJSON dan ToCSV untuk ekspor.
type Response struct {
	// Rows adalah baris data hasil normalisasi respons.
	Rows []map[string]any
	// Raw adalah respons JSON mentah (map kosong jika respons bukan dict).
	Raw map[string]any
}

// Len mengembalikan jumlah baris. Sama seperti len() di Python.
func (r Response) Len() int {
	return len(r.Rows)
}

// Count mengembalikan jumlah baris. Sama seperti properti .count di Python.
func (r Response) Count() int {
	return len(r.Rows)
}

// First mengembalikan baris pertama, atau nil jika tidak ada baris.
// Sama seperti properti .first di Python.
func (r Response) First() map[string]any {
	if len(r.Rows) == 0 {
		return nil
	}
	return r.Rows[0]
}

// Filter mengembalikan Response baru yang hanya berisi baris yang semua
// kondisinya cocok persis (exact equality), seperti filter() di Python.
//
// Perbedaan kecil dari Python: angka JSON di-decode Go sebagai float64,
// sehingga filter numerik membandingkan secara numerik — misalnya
// Filter(map[string]any{"tingkat": 10}) cocok dengan nilai JSON 10
// (float64(10)). Nilai non-numerik dibandingkan dengan reflect.DeepEqual.
func (r Response) Filter(conditions map[string]any) Response {
	if len(conditions) == 0 {
		return Response{Rows: append([]map[string]any(nil), r.Rows...), Raw: r.Raw}
	}
	filtered := make([]map[string]any, 0, len(r.Rows))
	for _, row := range r.Rows {
		match := true
		for k, v := range conditions {
			got, ok := row[k]
			if !ok || !equalValue(got, v) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, row)
		}
	}
	return Response{Rows: filtered, Raw: r.Raw}
}

// equalValue membandingkan dua nilai hasil decode JSON. Nilai numerik
// (int, float64, dsb.) dibandingkan sebagai float64 agar filter dengan
// literal int di Go tetap cocok dengan angka JSON.
func equalValue(got, want any) bool {
	if gf, ok := toFloat64(got); ok {
		if wf, ok := toFloat64(want); ok {
			return gf == wf
		}
		return false
	}
	if _, ok := toFloat64(want); ok {
		return false
	}
	return reflect.DeepEqual(got, want)
}

// toFloat64 mengubah nilai numerik menjadi float64, atau mengembalikan
// ok=false jika nilainya bukan angka.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// Pluck mengembalikan nilai satu field dari seluruh baris yang memiliki
// field tersebut, seperti pluck() di Python. Baris yang tidak memiliki
// key dilewati.
func (r Response) Pluck(key string) []any {
	out := make([]any, 0, len(r.Rows))
	for _, row := range r.Rows {
		if v, ok := row[key]; ok {
			out = append(out, v)
		}
	}
	return out
}

// ToList mengembalikan salinan slice baris, seperti to_list() di Python.
func (r Response) ToList() []map[string]any {
	return append([]map[string]any(nil), r.Rows...)
}

// ToJSON mengembalikan representasi JSON dari Rows.
// Ini adalah pengganti to_dataframe() dari Python untuk kebutuhan ekspor
// di Go (bersama ToCSV).
func (r Response) ToJSON() ([]byte, error) {
	return json.Marshal(r.Rows)
}

// ToCSV menulis Rows sebagai CSV ke w. Header adalah gabungan (union)
// semua key di seluruh baris, dalam urutan pertama kali key tersebut
// muncul. Nilai skalar ditulis dengan fmt.Sprint, nilai map/slice
// ditulis sebagai JSON, nilai nil ditulis kosong. Baris tanpa key yang
// ada di header ditulis kosong pada kolom tersebut.
func (r Response) ToCSV(w io.Writer) error {
	cw := csv.NewWriter(w)
	// Kumpulkan header: union key dalam urutan pertama muncul.
	var header []string
	seen := make(map[string]bool)
	for _, row := range r.Rows {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				header = append(header, k)
			}
		}
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, row := range r.Rows {
		record := make([]string, len(header))
		for i, k := range header {
			record[i] = csvValue(row[k])
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// csvValue mengubah satu nilai sel menjadi string untuk CSV.
func csvValue(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case map[string]any, []any:
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(t); err != nil {
			return fmt.Sprint(v)
		}
		return string(bytes.TrimRight(buf.Bytes(), "\n"))
	default:
		return fmt.Sprint(v)
	}
}
