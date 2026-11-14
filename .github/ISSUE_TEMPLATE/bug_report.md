---
name: Bug Report
about: Laporkan bug atau kendala pada dapodik-sdk-go
title: '[BUG] <deskripsi singkat bug>'
labels: 'bug'
assignees: ''
---

**Deskripsi Singkat Bug**
Jelaskan kendala atau bug yang Anda temukan.

**Langkah Reproduksi**
```go
package main

import (
	"context"

	"github.com/kilat-cloud/dapodik-sdk-go"
)

func main() {
	client, _ := dapodik.NewClient("...", "...")
	client.GetSekolah(context.Background(), "")
}
```

**Perilaku yang Diharapkan**
Jelaskan apa yang seharusnya terjadi.

**Informasi Lingkungan:**
- Go Version: [misal Go 1.25]
- OS: [misal macOS, Ubuntu, Windows]
- Versi Dapodik: [misal 2027/2027.b]
- Versi dapodik-sdk-go: [misal v1.1.0]
