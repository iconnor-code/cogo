// Package oss provides S3-compatible object storage helpers.
package oss

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
)

const (
	defaultPresignExpire = 15 * time.Minute
	emptyPayloadSHA256   = "UNSIGNED-PAYLOAD"
	objectHTTPTimeout    = 10 * time.Second
)

type Client struct {
	conf core.OSSConfig
}

type PresignedURL struct {
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	ObjectKey string    `json:"object_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewClient(conf core.IConfig) (*Client, error) {
	if conf == nil {
		return nil, cerrs.New("oss config not found")
	}
	ossConf := conf.GetOSS()
	if strings.TrimSpace(ossConf.Endpoint) == "" {
		return nil, cerrs.New("oss endpoint is required")
	}
	if strings.TrimSpace(ossConf.BucketName) == "" {
		return nil, cerrs.New("oss bucket name is required")
	}
	return &Client{conf: ossConf}, nil
}

func (c *Client) PublicURL(objectKey string) string {
	key := NormalizeObjectKey(objectKey)
	if key == "" {
		return ""
	}
	if isAbsoluteURL(key) {
		return key
	}
	baseURL := strings.TrimRight(c.conf.BaseURL, "/")
	if baseURL == "" {
		endpoint := c.endpointURL()
		baseURL = strings.TrimRight(endpoint.String(), "/") + "/" + c.conf.BucketName
	}
	return baseURL + "/" + pathEscape(key)
}

func (c *Client) ObjectKey(value string) string {
	key := NormalizeObjectKey(value)
	if key == "" || !isAbsoluteURL(key) {
		return key
	}
	parsed, err := url.Parse(key)
	if err != nil {
		return key
	}
	baseURL := strings.TrimRight(c.conf.BaseURL, "/")
	if baseURL != "" {
		if parsedBase, err := url.Parse(baseURL); err == nil && parsedBase.Scheme == parsed.Scheme && parsedBase.Host == parsed.Host {
			basePath := strings.TrimRight(parsedBase.Path, "/") + "/"
			if strings.HasPrefix(parsed.Path, basePath) {
				if unescaped, err := url.PathUnescape(strings.TrimPrefix(parsed.Path, basePath)); err == nil {
					return NormalizeObjectKey(unescaped)
				}
				return NormalizeObjectKey(strings.TrimPrefix(parsed.Path, basePath))
			}
		}
	}
	prefix := "/" + c.conf.BucketName + "/"
	if strings.HasPrefix(parsed.Path, prefix) {
		if unescaped, err := url.PathUnescape(strings.TrimPrefix(parsed.Path, prefix)); err == nil {
			return NormalizeObjectKey(unescaped)
		}
		return NormalizeObjectKey(strings.TrimPrefix(parsed.Path, prefix))
	}
	return key
}

func (c *Client) PresignedGetURL(objectKey string, expires time.Duration) (*PresignedURL, error) {
	return c.presign(http.MethodGet, objectKey, "", expires)
}

func (c *Client) PresignedHeadURL(objectKey string, expires time.Duration) (*PresignedURL, error) {
	return c.presign(http.MethodHead, objectKey, "", expires)
}

func (c *Client) PresignedPutURL(objectKey, contentType string, expires time.Duration) (*PresignedURL, error) {
	return c.presign(http.MethodPut, objectKey, contentType, expires)
}

func (c *Client) ObjectExists(objectKey string) (bool, error) {
	return c.ObjectExistsContext(context.Background(), objectKey)
}

func (c *Client) ObjectExistsContext(ctx context.Context, objectKey string) (bool, error) {
	signed, err := c.PresignedHeadURL(objectKey, 0)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, signed.URL, nil)
	if err != nil {
		return false, cerrs.Wrap(err, "oss create head object request failed")
	}
	resp, err := objectHTTPClient.Do(req)
	if err != nil {
		return false, cerrs.Wrap(err, "oss head object failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, cerrs.New("oss head object returned status " + resp.Status)
	}
	return true, nil
}

func (c *Client) GetObject(objectKey string) ([]byte, string, error) {
	return c.GetObjectContext(context.Background(), objectKey)
}

func (c *Client) GetObjectContext(ctx context.Context, objectKey string) ([]byte, string, error) {
	signed, err := c.PresignedGetURL(objectKey, 0)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signed.URL, nil)
	if err != nil {
		return nil, "", cerrs.Wrap(err, "oss create get object request failed")
	}
	resp, err := objectHTTPClient.Do(req)
	if err != nil {
		return nil, "", cerrs.Wrap(err, "oss get object failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, "", cerrs.New("oss get object returned status " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", cerrs.Wrap(err, "oss read object failed")
	}
	return body, resp.Header.Get("Content-Type"), nil
}

func (c *Client) PutObject(objectKey, contentType string, data []byte) error {
	return c.PutObjectContext(context.Background(), objectKey, contentType, data)
}

func (c *Client) PutObjectContext(ctx context.Context, objectKey, contentType string, data []byte) error {
	signed, err := c.PresignedPutURL(objectKey, contentType, 0)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signed.URL, bytes.NewReader(data))
	if err != nil {
		return cerrs.Wrap(err, "oss create put object request failed")
	}
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", strings.TrimSpace(contentType))
	}
	resp, err := objectHTTPClient.Do(req)
	if err != nil {
		return cerrs.Wrap(err, "oss put object failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return cerrs.New("oss put object returned status " + resp.Status + ": " + responseBodyString(resp.Body))
	}
	return nil
}

var objectHTTPClient = &http.Client{Timeout: objectHTTPTimeout}

func responseBodyString(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (c *Client) presign(method, objectKey, contentType string, expires time.Duration) (*PresignedURL, error) {
	if err := c.validateCredentials(); err != nil {
		return nil, err
	}
	key := NormalizeObjectKey(objectKey)
	if key == "" {
		return nil, cerrs.New("oss object key is required")
	}
	if isAbsoluteURL(key) {
		return &PresignedURL{
			Method:    method,
			URL:       key,
			ObjectKey: key,
			ExpiresAt: time.Now().Add(c.expireDuration(expires)),
		}, nil
	}

	now := time.Now().UTC()
	expires = c.expireDuration(expires)
	endpoint := c.presignURL(key)
	endpoint.RawQuery = ""

	credentialDate := now.Format("20060102")
	scope := credentialDate + "/us-east-1/s3/aws4_request"
	amzDate := now.Format("20060102T150405Z")
	query := endpoint.Query()
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", c.conf.AccessKeyID+"/"+scope)
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", strconv.Itoa(int(expires.Seconds())))

	signedHeaders := "host"
	headers := "host:" + endpoint.Host + "\n"
	if contentType != "" {
		signedHeaders = "content-type;host"
		headers = "content-type:" + strings.TrimSpace(contentType) + "\n" + headers
	}
	query.Set("X-Amz-SignedHeaders", signedHeaders)

	canonicalURI := "/" + pathEscape(strings.TrimPrefix(endpoint.Path, "/"))
	canonicalQuery := canonicalQueryString(query)
	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		canonicalQuery,
		headers,
		signedHeaders,
		emptyPayloadSHA256,
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hexSHA256(canonicalRequest),
	}, "\n")
	signingKey := awsSigningKey(c.conf.AccessKeySecret, credentialDate)
	signature := hexHMAC(signingKey, stringToSign)
	query.Set("X-Amz-Signature", signature)
	endpoint.RawQuery = query.Encode()

	return &PresignedURL{
		Method:    method,
		URL:       endpoint.String(),
		ObjectKey: key,
		ExpiresAt: now.Add(expires),
	}, nil
}

func (c *Client) validateCredentials() error {
	accessKeyID := strings.TrimSpace(c.conf.AccessKeyID)
	accessKeySecret := strings.TrimSpace(c.conf.AccessKeySecret)
	if accessKeyID == "" || accessKeySecret == "" {
		return cerrs.New("oss access key is required")
	}
	if strings.HasPrefix(accessKeyID, "replace-me-") || strings.HasPrefix(accessKeySecret, "replace-me-") {
		return cerrs.New("oss access key is not configured")
	}
	return nil
}

func (c *Client) expireDuration(value time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	if c.conf.PresignExpire > 0 {
		return time.Duration(c.conf.PresignExpire) * time.Second
	}
	return defaultPresignExpire
}

func (c *Client) endpointURL() url.URL {
	scheme := "http"
	if c.conf.UseSSL {
		scheme = "https"
	}
	endpoint := strings.TrimSpace(c.conf.Endpoint)
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return *parsed
	}
	return url.URL{Scheme: scheme, Host: endpoint}
}

func (c *Client) presignURL(objectKey string) url.URL {
	baseURL := strings.TrimRight(c.conf.BaseURL, "/")
	if baseURL != "" {
		if parsed, err := url.Parse(baseURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			parsed.Path = path.Join(parsed.Path, objectKey)
			return *parsed
		}
	}
	endpoint := c.endpointURL()
	endpoint.Path = "/" + path.Join(c.conf.BucketName, objectKey)
	return endpoint
}

func NormalizeObjectKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || isAbsoluteURL(value) {
		return value
	}
	value = strings.TrimPrefix(value, "/")
	return path.Clean(value)
}

func NewObjectKey(prefix, filename string) string {
	ext := strings.ToLower(path.Ext(strings.TrimSpace(filename)))
	if len(ext) > 32 {
		ext = ""
	}
	key := uuid.NewString() + ext
	prefix = NormalizeObjectKey(prefix)
	if prefix == "" || prefix == "." {
		return key
	}
	return path.Join(prefix, key)
}

func isAbsoluteURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func canonicalQueryString(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0)
	for _, key := range keys {
		items := append([]string(nil), values[key]...)
		sort.Strings(items)
		for _, value := range items {
			parts = append(parts, uriEncode(key)+"="+uriEncode(value))
		}
	}
	return strings.Join(parts, "&")
}

func pathEscape(value string) string {
	segments := strings.Split(value, "/")
	for i, segment := range segments {
		segments[i] = uriEncode(segment)
	}
	return strings.Join(segments, "/")
}

func uriEncode(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}

func hexSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func awsSigningKey(secret, date string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), date)
	regionKey := hmacSHA256(dateKey, "us-east-1")
	serviceKey := hmacSHA256(regionKey, "s3")
	return hmacSHA256(serviceKey, "aws4_request")
}

func hexHMAC(key []byte, value string) string {
	return hex.EncodeToString(hmacSHA256(key, value))
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
