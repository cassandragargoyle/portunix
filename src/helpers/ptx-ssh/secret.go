/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
)

// SecretString wraps a credential value that must never be logged, printed, or
// serialized in cleartext. Its String, GoString, MarshalJSON, MarshalYAML, and
// MarshalText methods all return a masked token, so values accidentally passed
// to fmt.Print / log / json.Marshal never leak the underlying credential.
//
// Callers must use Reveal() to obtain the plaintext and call Zero() when done.
type SecretString struct {
	value []byte
}

// NewSecret creates a SecretString from a plaintext string. The input string
// remains in memory managed by the Go runtime; callers that need the strongest
// guarantees should assemble the credential as []byte and use NewSecretBytes.
func NewSecret(value string) SecretString {
	if value == "" {
		return SecretString{}
	}
	b := make([]byte, len(value))
	copy(b, value)
	return SecretString{value: b}
}

// NewSecretBytes wraps a byte slice. The slice is retained by reference so the
// caller should not modify it afterwards. Zero() will overwrite the underlying
// bytes in place.
func NewSecretBytes(b []byte) SecretString {
	return SecretString{value: b}
}

// Reveal returns the underlying credential value. Use only at the point it is
// passed to the SSH auth callback; never to print or log.
func (s SecretString) Reveal() string {
	return string(s.value)
}

// IsEmpty reports whether the secret has no value set.
func (s SecretString) IsEmpty() bool {
	return len(s.value) == 0
}

// Zero overwrites the underlying buffer with zeros. After calling Zero, Reveal
// returns empty string.
func (s *SecretString) Zero() {
	for i := range s.value {
		s.value[i] = 0
	}
	s.value = nil
}

// maskToken is the value printed whenever a SecretString is rendered through
// any standard formatting or serialization path.
const maskToken = "***"

// String satisfies fmt.Stringer and blocks accidental %s / %v leaks.
func (s SecretString) String() string {
	if len(s.value) == 0 {
		return ""
	}
	return maskToken
}

// GoString satisfies fmt.GoStringer and blocks accidental %#v leaks.
func (s SecretString) GoString() string {
	if len(s.value) == 0 {
		return `SecretString{}`
	}
	return `SecretString{value:"***"}`
}

// MarshalJSON ensures json.Marshal produces the mask, never the value.
func (s SecretString) MarshalJSON() ([]byte, error) {
	if len(s.value) == 0 {
		return []byte(`""`), nil
	}
	return json.Marshal(maskToken)
}

// MarshalYAML ensures yaml.Marshal produces the mask, never the value.
func (s SecretString) MarshalYAML() (interface{}, error) {
	if len(s.value) == 0 {
		return "", nil
	}
	return maskToken, nil
}

// MarshalText ensures encoding.TextMarshaler usage produces the mask.
func (s SecretString) MarshalText() ([]byte, error) {
	if len(s.value) == 0 {
		return []byte{}, nil
	}
	return []byte(maskToken), nil
}
