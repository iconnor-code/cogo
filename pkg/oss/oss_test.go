package oss

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
)

func TestPresignedPutURL(t *testing.T) {
	client, err := NewClient(&cogoconfig.Config{Config: core.Config{OSS: core.OSSConfig{
		Endpoint:        "minio.local:9001",
		AccessKeyID:     "access",
		AccessKeySecret: "secret",
		BucketName:      "mysite",
	}}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	signed, err := client.PresignedPutURL("avatars/user 1.png", "image/png", time.Minute)
	if err != nil {
		t.Fatalf("presign put: %v", err)
	}
	parsed, err := url.Parse(signed.URL)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	if signed.Method != "PUT" {
		t.Fatalf("method = %q, want PUT", signed.Method)
	}
	if parsed.Host != "minio.local:9001" {
		t.Fatalf("host = %q, want minio.local:9001", parsed.Host)
	}
	if parsed.Path != "/mysite/avatars/user 1.png" {
		t.Fatalf("path = %q, want object path", parsed.Path)
	}
	if parsed.Query().Get("X-Amz-Signature") == "" {
		t.Fatalf("signature missing")
	}
	if parsed.Query().Get("X-Amz-SignedHeaders") != "content-type;host" {
		t.Fatalf("signed headers = %q", parsed.Query().Get("X-Amz-SignedHeaders"))
	}
}

func TestPublicURLKeepsAbsoluteURL(t *testing.T) {
	client := &Client{conf: core.OSSConfig{BaseURL: "http://minio.local:9001/mysite"}}
	got := client.PublicURL("https://example.com/a.png")
	if got != "https://example.com/a.png" {
		t.Fatalf("public url = %q", got)
	}
	got = client.PublicURL("/article-covers/a b.png")
	if !strings.HasPrefix(got, "http://minio.local:9001/mysite/article-covers/a%20b.png") {
		t.Fatalf("public url = %q", got)
	}
}

func TestObjectKeyConvertsBucketURL(t *testing.T) {
	client := &Client{conf: core.OSSConfig{
		Endpoint:   "minio.local:9001",
		BucketName: "mysite",
		BaseURL:    "http://minio.local:9001/mysite",
	}}
	got := client.ObjectKey("http://minio.local:9001/mysite/article-covers/a%20b.png?X-Amz-Signature=abc")
	if got != "article-covers/a b.png" {
		t.Fatalf("object key = %q", got)
	}
	got = client.ObjectKey("https://example.com/a.png")
	if got != "https://example.com/a.png" {
		t.Fatalf("external url = %q", got)
	}
}
