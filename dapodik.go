package dapodik

import "time"

// Dapodik adalah factory helper, port dari class Dapodik di Python.
// Simpan satu Dapodik per host WebService, lalu panggil API untuk
// membuat Client per (token, NPSN).
type Dapodik struct {
	// Host adalah host WebService Dapodik.
	Host string
	// Port adalah port WebService Dapodik.
	Port int
	// Timeout adalah timeout per request.
	Timeout time.Duration
}

// NewDapodik membuat factory Dapodik. Host kosong menjadi DefaultHost,
// port <= 0 menjadi DefaultPort, timeout <= 0 menjadi DefaultTimeout.
func NewDapodik(host string, port int, timeout time.Duration) *Dapodik {
	if host == "" {
		host = DefaultHost
	}
	if port <= 0 {
		port = DefaultPort
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Dapodik{Host: host, Port: port, Timeout: timeout}
}

// API membuat Client baru untuk token dan NPSN tertentu, seperti
// Dapodik.api(token, npsn) di Python. Validasi NPSN/token sama seperti
// NewClient; error validasi dikembalikan sebagai error.
func (d *Dapodik) API(token, npsn string) (*Client, error) {
	return NewClient(npsn, token,
		WithHost(d.Host),
		WithPort(d.Port),
		WithTimeout(d.Timeout),
	)
}
