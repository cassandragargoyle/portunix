/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const secretValue = "super-secret-password"

func TestSecretStringMaskOnStringer(t *testing.T) {
	s := NewSecret(secretValue)
	if got := s.String(); got != "***" {
		t.Errorf("String() = %q, want \"***\"", got)
	}
	formatted := fmt.Sprintf("password=%s", s)
	if strings.Contains(formatted, secretValue) {
		t.Errorf("fmt.Sprintf leaked secret: %q", formatted)
	}
}

func TestSecretStringMaskOnGoStringer(t *testing.T) {
	s := NewSecret(secretValue)
	formatted := fmt.Sprintf("%#v", s)
	if strings.Contains(formatted, secretValue) {
		t.Errorf("%%#v leaked secret: %q", formatted)
	}
}

func TestSecretStringMaskOnJSON(t *testing.T) {
	s := NewSecret(secretValue)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(data), secretValue) {
		t.Errorf("json leaked secret: %s", data)
	}
	if string(data) != `"***"` {
		t.Errorf("json = %s, want \"***\"", data)
	}
}

func TestSecretStringMaskInStruct(t *testing.T) {
	wrapper := struct {
		User     string       `json:"user"`
		Password SecretString `json:"password"`
	}{
		User:     "alice",
		Password: NewSecret(secretValue),
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(data), secretValue) {
		t.Errorf("struct json leaked secret: %s", data)
	}
}

func TestSecretStringRevealReturnsValue(t *testing.T) {
	s := NewSecret(secretValue)
	if got := s.Reveal(); got != secretValue {
		t.Errorf("Reveal() = %q, want %q", got, secretValue)
	}
}

func TestSecretStringZeroClears(t *testing.T) {
	s := NewSecret(secretValue)
	s.Zero()
	if got := s.Reveal(); got != "" {
		t.Errorf("Reveal after Zero = %q, want empty", got)
	}
	if !s.IsEmpty() {
		t.Errorf("IsEmpty after Zero = false, want true")
	}
}

func TestSecretStringEmpty(t *testing.T) {
	s := NewSecret("")
	if !s.IsEmpty() {
		t.Errorf("IsEmpty on empty secret = false")
	}
	if got := s.String(); got != "" {
		t.Errorf("String on empty = %q, want empty", got)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != `""` {
		t.Errorf("empty json = %s, want \"\"", data)
	}
}

func TestSecretStringMarshalText(t *testing.T) {
	s := NewSecret(secretValue)
	text, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(text) != "***" {
		t.Errorf("MarshalText = %q, want \"***\"", text)
	}
}
