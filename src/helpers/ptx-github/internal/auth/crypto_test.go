/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package auth

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	cs := NewCryptoService("test-seed-1234567890")
	cases := []string{
		"",
		"ghp_helloworld",
		"long token with spaces and unicode přílíš žluťoučký",
	}
	for _, plain := range cases {
		enc, err := cs.Encrypt(plain)
		if err != nil {
			t.Fatalf("encrypt %q: %v", plain, err)
		}
		got, err := cs.Decrypt(enc)
		if err != nil {
			t.Fatalf("decrypt %q: %v", plain, err)
		}
		if got != plain {
			t.Fatalf("round-trip mismatch: want %q got %q", plain, got)
		}
	}
}

func TestDecryptWrongKey(t *testing.T) {
	good := NewCryptoService("seed-A")
	bad := NewCryptoService("seed-B")
	enc, err := good.Encrypt("secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bad.Decrypt(enc); err == nil {
		t.Fatalf("expected decryption to fail with wrong key")
	}
}

func TestEncryptDifferentIVs(t *testing.T) {
	cs := NewCryptoService("same-seed")
	a, _ := cs.Encrypt("hello")
	b, _ := cs.Encrypt("hello")
	if a == b {
		t.Fatalf("expected distinct ciphertexts due to random IVs")
	}
}

func TestMachineSeed(t *testing.T) {
	s1, err := MachineSeed("")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := MachineSeed("")
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 {
		t.Fatalf("MachineSeed not deterministic: %q vs %q", s1, s2)
	}
	withPw, _ := MachineSeed("pw")
	if withPw == s1 {
		t.Fatalf("expected password to change the seed")
	}
}
