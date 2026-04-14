package web

import (
	"io/fs"
	"strings"
	"testing"
)

func TestFS_IndexHTML(t *testing.T) {
	f, err := FS.Open("index.html")
	if err != nil {
		t.Fatalf("FS.Open('index.html') failed: %v", err)
	}
	f.Close()
}

func TestFS_VendorPreact(t *testing.T) {
	f, err := FS.Open("vendor/preact.mjs")
	if err != nil {
		t.Fatalf("FS.Open('vendor/preact.mjs') failed: %v", err)
	}
	f.Close()
}

func TestFS_StylesMainCSS(t *testing.T) {
	f, err := FS.Open("styles/main.css")
	if err != nil {
		t.Fatalf("FS.Open('styles/main.css') failed: %v", err)
	}
	f.Close()
}

func TestFS_VendorHTMPreact(t *testing.T) {
	f, err := FS.Open("vendor/htm-preact.mjs")
	if err != nil {
		t.Fatalf("FS.Open('vendor/htm-preact.mjs') failed: %v", err)
	}
	f.Close()
}

func TestFS_VendorPreactHooks(t *testing.T) {
	f, err := FS.Open("vendor/preact-hooks.mjs")
	if err != nil {
		t.Fatalf("FS.Open('vendor/preact-hooks.mjs') failed: %v", err)
	}
	f.Close()
}

func TestFS_VendorSignals(t *testing.T) {
	f, err := FS.Open("vendor/signals.mjs")
	if err != nil {
		t.Fatalf("FS.Open('vendor/signals.mjs') failed: %v", err)
	}
	f.Close()
}

func TestFS_AppJS(t *testing.T) {
	f, err := FS.Open("app.js")
	if err != nil {
		t.Fatalf("FS.Open('app.js') failed: %v", err)
	}
	f.Close()
}

func TestFS_Components(t *testing.T) {
	files := []string{
		"components/search-bar.js",
		"components/job-table.js",
		"components/job-row.js",
		"components/confirm-dialog.js",
		"components/toast.js",
		"components/log-viewer.js",
		"components/diagnose-panel.js",
	}
	for _, name := range files {
		f, err := FS.Open(name)
		if err != nil {
			t.Errorf("FS.Open(%q) failed: %v", name, err)
			continue
		}
		f.Close()
	}
}

func TestFS_Lib(t *testing.T) {
	files := []string{
		"lib/state.js",
		"lib/sse.js",
		"lib/api.js",
	}
	for _, name := range files {
		f, err := FS.Open(name)
		if err != nil {
			t.Errorf("FS.Open(%q) failed: %v", name, err)
			continue
		}
		f.Close()
	}
}

func TestIndexHTML_ImportMap_NoExternalCDN(t *testing.T) {
	data, err := fs.ReadFile(FS, "index.html")
	if err != nil {
		t.Fatalf("ReadFile('index.html') failed: %v", err)
	}
	html := string(data)

	// Must contain import map with local vendor paths.
	if !strings.Contains(html, `"importmap"`) {
		t.Error("index.html does not contain an import map (type=\"importmap\")")
	}
	if !strings.Contains(html, `./vendor/preact.mjs`) {
		t.Error("import map missing local path for preact")
	}

	// No CDN URLs in production index.html.
	cdnPatterns := []string{
		"https://cdn.jsdelivr.net",
		"https://unpkg.com",
		"https://esm.sh",
		"https://cdn.skypack.dev",
	}
	for _, pat := range cdnPatterns {
		if strings.Contains(html, pat) {
			t.Errorf("index.html contains CDN URL %q — must use local vendor paths", pat)
		}
	}
}
