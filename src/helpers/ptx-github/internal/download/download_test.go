/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDownloadBasic(t *testing.T) {
	body := []byte("hello world")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.bin")
	res, err := Download(Options{URL: srv.URL, Dest: dest})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if res.Size != int64(len(body)) {
		t.Fatalf("size: got %d want %d", res.Size, len(body))
	}
	got, _ := os.ReadFile(dest)
	if string(got) != string(body) {
		t.Fatalf("content mismatch: %q", got)
	}
}

func TestDownloadChecksumOK(t *testing.T) {
	body := []byte("checksum-me")
	sum := sha256.Sum256(body)
	expected := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.bin")
	res, err := Download(Options{URL: srv.URL, Dest: dest, ExpectedSHA: expected})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !strings.EqualFold(res.SHA256, expected) {
		t.Fatalf("hash: got %s want %s", res.SHA256, expected)
	}
}

func TestDownloadChecksumMismatch(t *testing.T) {
	body := []byte("data")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.bin")
	_, err := Download(Options{
		URL: srv.URL, Dest: dest,
		ExpectedSHA: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	if err == nil {
		t.Fatalf("expected checksum mismatch error")
	}
	if _, statErr := os.Stat(dest); statErr == nil {
		t.Fatalf("destination should be removed after checksum mismatch")
	}
}

func TestDownloadResume(t *testing.T) {
	full := []byte("0123456789ABCDEF")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHdr := r.Header.Get("Range")
		if rangeHdr != "" {
			// Expect "bytes=N-"
			parts := strings.SplitN(strings.TrimPrefix(rangeHdr, "bytes="), "-", 2)
			start, _ := strconv.Atoi(parts[0])
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(full)-1, len(full)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(full)-start))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(full[start:])
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(full)))
		w.WriteHeader(http.StatusOK)
		w.Write(full)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.bin")
	if err := os.WriteFile(dest, full[:6], 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Download(Options{URL: srv.URL, Dest: dest, Resume: true})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !res.Resumed {
		t.Fatalf("expected Resumed=true, got %+v", res)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != string(full) {
		t.Fatalf("content mismatch: got %q want %q", got, full)
	}
}

func TestDownloadHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := Download(Options{
		URL:  srv.URL,
		Dest: filepath.Join(t.TempDir(), "out"),
	})
	if err == nil {
		t.Fatalf("expected error on 403")
	}
}

func TestVerifyFile(t *testing.T) {
	body := []byte("verify me")
	dest := filepath.Join(t.TempDir(), "v.bin")
	if err := os.WriteFile(dest, body, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	expected := hex.EncodeToString(sum[:])
	got, err := VerifyFile(dest, expected)
	if err != nil {
		t.Fatalf("VerifyFile: %v", err)
	}
	if got != expected {
		t.Fatalf("hash: got %s want %s", got, expected)
	}
	if _, err := VerifyFile(dest, "deadbeef"); err == nil {
		t.Fatalf("expected mismatch")
	}
}
