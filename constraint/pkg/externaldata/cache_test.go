package externaldata

import (
	"testing"

	"github.com/open-policy-agent/frameworks/constraint/pkg/apis/externaldata/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	validCABundle   = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUIwekNDQVgyZ0F3SUJBZ0lKQUkvTTdCWWp3Qit1TUEwR0NTcUdTSWIzRFFFQkJRVUFNRVV4Q3pBSkJnTlYKQkFZVEFrRlZNUk13RVFZRFZRUUlEQXBUYjIxbExWTjBZWFJsTVNFd0h3WURWUVFLREJoSmJuUmxjbTVsZENCWAphV1JuYVhSeklGQjBlU0JNZEdRd0hoY05NVEl3T1RFeU1qRTFNakF5V2hjTk1UVXdPVEV5TWpFMU1qQXlXakJGCk1Rc3dDUVlEVlFRR0V3SkJWVEVUTUJFR0ExVUVDQXdLVTI5dFpTMVRkR0YwWlRFaE1COEdBMVVFQ2d3WVNXNTAKWlhKdVpYUWdWMmxrWjJsMGN5QlFkSGtnVEhSa01Gd3dEUVlKS29aSWh2Y05BUUVCQlFBRFN3QXdTQUpCQU5MSgpoUEhoSVRxUWJQa2xHM2liQ1Z4d0dNUmZwL3Y0WHFoZmRRSGRjVmZIYXA2TlE1V29rLzR4SUErdWkzNS9NbU5hCnJ0TnVDK0JkWjF0TXVWQ1BGWmNDQXdFQUFhTlFNRTR3SFFZRFZSME9CQllFRkp2S3M4UmZKYVhUSDA4VytTR3YKelF5S24wSDhNQjhHQTFVZEl3UVlNQmFBRkp2S3M4UmZKYVhUSDA4VytTR3Z6UXlLbjBIOE1Bd0dBMVVkRXdRRgpNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUZCUUFEUVFCSmxmZkpIeWJqREd4Uk1xYVJtRGhYMCs2djAyVFVLWnNXCnI1UXVWYnBRaEg2dSswVWdjVzBqcDlRd3B4b1BUTFRXR1hFV0JCQnVyeEZ3aUNCaGtRK1YKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	badBase64String = "!"
	badCABundle     = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCmhlbGxvCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
)

type cacheTestCase struct {
	Name          string
	Provider      *v1alpha1.Provider
	ErrorExpected bool
}

func createProvider(name string, url string, timeout int, caBundle string, insecureTLSSkipVerify bool) *v1alpha1.Provider {
	return &v1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ProviderSpec{
			URL:                   url,
			Timeout:               timeout,
			CABundle:              caBundle,
			InsecureTLSSkipVerify: insecureTLSSkipVerify,
		},
	}
}

func TestUpsert(t *testing.T) {
	tc := []cacheTestCase{
		{
			Name:          "valid http provider",
			Provider:      createProvider("test", "http://test", 1, "", true),
			ErrorExpected: false,
		},
		{
			Name:          "http provider with caBundle and insecureTLSSkipVerify",
			Provider:      createProvider("test", "http://test", 1, validCABundle, true),
			ErrorExpected: false,
		},
		{
			Name:          "http provider with bad caBundle and insecureTLSSkipVerify",
			Provider:      createProvider("test", "http://test", 1, badCABundle, true),
			ErrorExpected: true,
		},
		{
			Name:          "http provider without insecure tls skip verify",
			Provider:      createProvider("test", "http://test", 1, "", false),
			ErrorExpected: true,
		},
		{
			Name:          "valid https provider",
			Provider:      createProvider("test", "https://test", 1, validCABundle, false),
			ErrorExpected: false,
		},
		{
			Name:          "https provider with no caBundle",
			Provider:      createProvider("test", "https://test", 1, "", false),
			ErrorExpected: true,
		},
		{
			Name:          "https provider with bad base64 caBundle",
			Provider:      createProvider("test", "https://test", 1, badBase64String, false),
			ErrorExpected: true,
		},
		{
			Name:          "https provider with bad caBundle",
			Provider:      createProvider("test", "https://test", 1, badCABundle, false),
			ErrorExpected: true,
		},
		{
			Name:          "empty name",
			Provider:      createProvider("", "http://test", 1, "", true),
			ErrorExpected: true,
		},
		{
			Name:          "empty url",
			Provider:      createProvider("test", "", 1, "", true),
			ErrorExpected: true,
		},
		{
			Name:          "url with invalid scheme",
			Provider:      createProvider("test", "gopher://test", 1, "", true),
			ErrorExpected: true,
		},
		{
			Name:          "invalid url",
			Provider:      createProvider("test", " http://foo.com", 1, "", true),
			ErrorExpected: true,
		},
		{
			Name:          "invalid timeout",
			Provider:      createProvider("test", "http://test", -1, "", true),
			ErrorExpected: true,
		},
		{
			Name:          "empty provider",
			Provider:      &v1alpha1.Provider{},
			ErrorExpected: true,
		},
	}
	for _, tt := range tc {
		cache := NewCache()
		t.Run(tt.Name, func(t *testing.T) {
			err := cache.Upsert(tt.Provider)

			if (err == nil) && tt.ErrorExpected {
				t.Fatalf("err = nil; want non-nil")
			}
			if (err != nil) && !tt.ErrorExpected {
				t.Fatalf("err = \"%s\"; want nil", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tc := []cacheTestCase{
		{
			Name:          "valid provider",
			Provider:      createProvider("test", "http://test", 1, "", true),
			ErrorExpected: false,
		},
		{
			Name:          "invalid provider",
			Provider:      createProvider("", "http://test", 1, "", true),
			ErrorExpected: true,
		},
	}
	for _, tt := range tc {
		cache := NewCache()
		t.Run(tt.Name, func(t *testing.T) {
			_ = cache.Upsert(tt.Provider)
			_, err := cache.Get(tt.Provider.Name)

			if (err == nil) && tt.ErrorExpected {
				t.Fatalf("err = nil; want non-nil")
			}
			if (err != nil) && !tt.ErrorExpected {
				t.Fatalf("err = \"%s\"; want nil", err)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tc := []cacheTestCase{
		{
			Name:          "valid provider",
			Provider:      createProvider("test", "http://test", 1, "", true),
			ErrorExpected: false,
		},
	}
	for _, tt := range tc {
		cache := NewCache()
		t.Run(tt.Name, func(t *testing.T) {
			_ = cache.Upsert(tt.Provider)
			cache.Remove(tt.Provider.Name)

			if (cache != nil) && tt.ErrorExpected {
				t.Fatalf("cache = \"%v\"; want nil", cache)
			}
		})
	}
}
