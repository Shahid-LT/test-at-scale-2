package logstream

import (
	"bytes"
	"testing"
)

func TestReplace(t *testing.T) {
	secrets := map[string]string{
		"cipher":  "lazy dog",
		"cipher2": "",
	}
	buf := &bytes.Buffer{}
	w := NewMasker(buf, secrets)
	w.Write([]byte("The quick brown fox jumps over the lazy dog")) // nolint:errcheck

	if got, want := buf.String(), "The quick brown fox jumps over the ****************"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

func TestReplaceMultiline(t *testing.T) {
	key := `
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
-----END RSA PRIVATE KEY-----`

	line := `> MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp`

	secrets := map[string]string{
		"cipher": key,
	}
	buf := &bytes.Buffer{}
	w := NewMasker(buf, secrets)
	w.Write([]byte(line)) // nolint:errcheck

	if got, want := buf.String(), "> ****************"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

func TestSkipSingleCharacterMask(t *testing.T) {
	secrets := map[string]string{
		"cipher": "l",
	}
	buf := &bytes.Buffer{}
	w := NewMasker(buf, secrets)
	w.Write([]byte("The quick brown fox jumps over the lazy dog")) // nolint:errcheck

	if got, want := buf.String(), "The quick brown fox jumps over the lazy dog"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

func TestReplaceMultilineJson(t *testing.T) {
	key := `{
  "token":"dXNlcm5hbWU6cGFzc3dvcmQ="
}`

	line := `{
  "token":"dXNlcm5hbWU6cGFzc3dvcmQ="
}`

	secrets := map[string]string{
		"cipher": key,
	}
	buf := &bytes.Buffer{}
	w := NewMasker(buf, secrets)
	w.Write([]byte(line)) // nolint:errcheck

	if got, want := buf.String(), "{\n  ****************\n}"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}
