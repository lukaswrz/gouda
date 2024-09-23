// Package godl provides utilities for serving files over HTTP, with MIME type
// inference and support for setting Content-Disposition headers to control
// whether files are served inline or as attachments.
package godl

import (
	"mime"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

// Infer returns the MIME type of the file specified by the given path. It
// first attempts to determine the MIME type using InferByMagic, and if
// unsuccessful, it falls back to InferByExtension.
func Infer(path string) string {
	m := InferByMagic(path)
	if m == "" {
		return InferByExtension(path)
	}
	return m
}

// InferByExtension returns the MIME type of the file specified by the given
// path using the file extension, or an empty string if no match is found.
func InferByExtension(path string) string {
	return mime.TypeByExtension(filepath.Ext(path))
}

// InferByMagic returns the MIME type of the file specified by the given
// path using the mimetype module, or an empty string if no match is found.
func InferByMagic(path string) string {
	if m, err := mimetype.DetectFile(path); err == nil {
		return m.String()
	}
	return ""
}

// SetContentType sets the Content-Type header for the file specified by the
// given path, inferred using the provided infer function.
func SetContentType(w http.ResponseWriter, path string, infer func(string) string) {
	m := infer(path)
	if m == "" {
		m = "application/octet-stream"
	}
	w.Header().Set("Content-Type", m)
}

// SetAttachment sets the Content-Disposition header to inform the client
// that the file is an attachment, specifying the name of the file.
func SetAttachment(w http.ResponseWriter, name string) {
	w.Header().Set(
		"Content-Disposition",
		"attachment; filename*=UTF-8''"+url.QueryEscape(name),
	)
}

// ServeAttachment serves a file with the specified name and path, setting
// the Content-Type header using the provided infer function and marking it as
// an attachment by setting the Content-Disposition header.
func ServeAttachment(w http.ResponseWriter, r *http.Request, path string, name string, infer func(string) string) {
	SetContentType(w, path, infer)
	SetAttachment(w, name)
	http.ServeFile(w, r, path)
}

// ServeDownload serves a file with the specified name and path, setting the
// Content-Type header using the provided infer function and determining
// whether to show the file inline based on the list of inline types. If the
// list is empty, all content types are treated as inline. Additionally, it
// sets the Content-Disposition header accordingly.
func ServeDownload(w http.ResponseWriter, r *http.Request, path string, name string, inlineTypes []string, infer func(string) string) {
	SetContentType(w, path, infer)

	inline := len(inlineTypes) == 0
	for _, it := range inlineTypes {
		if it == w.Header().Get("Content-Type") {
			inline = true
			break
		}
	}

	if !inline {
		SetAttachment(w, name)
	}

	http.ServeFile(w, r, path)
}
