// license-e2e exercises the live v1.4 Backend -> Switch license path.
// It deliberately refuses remote hosts by default so this cannot accidentally
// create test licenses against production systems.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	licenseprotocol "horizon-core/internal/license"
)

type issueResponse struct {
	License licenseprotocol.SignedLicense `json:"license"`
}

type hwidResponse struct {
	HardwareID string `json:"hardware_id"`
	Error      string `json:"error"`
}

type checkResponse struct {
	Valid    bool   `json:"valid"`
	Verified bool   `json:"verified"`
	Error    string `json:"error"`
}

type tester struct {
	client     *http.Client
	backendURL string
	switchURL  string
	apiKey     string
	adminToken string
}

func main() {
	backend := flag.String("backend", "http://127.0.0.1:8080", "local Backend base URL")
	sw := flag.String("switch", "http://127.0.0.1:8081", "local Switch base URL")
	flag.Parse()

	if os.Getenv("HORIZON_E2E_ALLOW_REMOTE") != "YES" {
		if err := requireLocalURL(*backend); err != nil {
			fail(err)
		}
		if err := requireLocalURL(*sw); err != nil {
			fail(err)
		}
	}

	apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
	adminToken := strings.TrimSpace(os.Getenv("ADMIN_TOKEN"))
	if len(apiKey) < 32 || len(adminToken) < 32 {
		fail(errors.New("set API_KEY and ADMIN_TOKEN in the local shell; credentials are not read from or printed by this test"))
	}

	t := &tester{
		client: &http.Client{Timeout: 12 * time.Second},
		backendURL: strings.TrimRight(*backend, "/"),
		switchURL: strings.TrimRight(*sw, "/"),
		apiKey: apiKey, adminToken: adminToken,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if err := t.checkHealth(ctx); err != nil {
		fail(fmt.Errorf("health preflight: %w", err))
	}
	hardwareID, err := t.getHardwareID(ctx)
	if err != nil {
		fail(fmt.Errorf("read Switch HWID: %w", err))
	}
	signed, err := t.issue(ctx, hardwareID)
	if err != nil {
		fail(fmt.Errorf("issue test license: %w", err))
	}
	if strings.TrimSpace(signed.Statement.LicenseID) == "" {
		fail(errors.New("Backend returned an empty license_id"))
	}

	valid, status, err := t.check(ctx, signed)
	if err != nil {
		fail(fmt.Errorf("valid-license check: %w", err))
	}
	if status < 200 || status >= 300 || !valid.Valid || !valid.Verified {
		fail(fmt.Errorf("valid license was rejected: status=%d valid=%t verified=%t error=%s", status, valid.Valid, valid.Verified, valid.Error))
	}
	fmt.Printf("PASS valid signed license: %s\n", signed.Statement.LicenseID)

	tamperedSignature := signed
	tamperedSignature.Signature = flipLastHexDigit(tamperedSignature.Signature)
	if result, code, err := t.check(ctx, tamperedSignature); err != nil {
		fail(fmt.Errorf("tampered-signature check: %w", err))
	} else if code != http.StatusForbidden || result.Valid || result.Verified {
		fail(fmt.Errorf("tampered signature was not strictly rejected: status=%d valid=%t verified=%t error=%s", code, result.Valid, result.Verified, result.Error))
	}
	fmt.Println("PASS altered signature rejected")

	tamperedStatement := signed
	tamperedStatement.Statement.Volume++
	if result, code, err := t.check(ctx, tamperedStatement); err != nil {
		fail(fmt.Errorf("tampered-statement check: %w", err))
	} else if code != http.StatusForbidden || result.Valid || result.Verified {
		fail(fmt.Errorf("altered statement was not strictly rejected: status=%d valid=%t verified=%t error=%s", code, result.Valid, result.Verified, result.Error))
	}
	fmt.Println("PASS altered statement rejected")

	otherHardware := signed
	otherHardware.Statement.HardwareID += "-OTHER"
	if result, code, err := t.check(ctx, otherHardware); err != nil {
		fail(fmt.Errorf("wrong-HWID check: %w", err))
	} else if code != http.StatusForbidden || result.Valid || result.Verified || !strings.Contains(strings.ToLower(result.Error), "hardware_id mismatch") {
		fail(fmt.Errorf("wrong HWID was not explicitly rejected: status=%d valid=%t verified=%t error=%s", code, result.Valid, result.Verified, result.Error))
	}
	fmt.Println("PASS mismatched HWID rejected before accepting the statement")
	fmt.Println("PASS Horizon v1.4 live Backend -> Switch checks completed")
	fmt.Println("Note: one local validation license remains in Backend/Switch registries; no credentials or signatures were printed.")
}

func (t *tester) checkHealth(ctx context.Context) error {
	for _, target := range []struct{ url, service string }{
		{t.backendURL + "/api/v1/health", "horizon-backend"},
		{t.switchURL + "/health", "horizon-switch"},
	} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.url, nil)
		if err != nil { return err }
		res, err := t.client.Do(req)
		if err != nil { return err }
		body, readErr := io.ReadAll(io.LimitReader(res.Body, 64*1024))
		_ = res.Body.Close()
		if readErr != nil { return readErr }
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("%s returned HTTP %d", target.service, res.StatusCode)
		}
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil { return err }
		if data["status"] != "online" || data["service"] != target.service {
			return fmt.Errorf("unexpected %s health response", target.service)
		}
	}
	return nil
}

func (t *tester) getHardwareID(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.switchURL+"/api/v1/license/hardware-id", nil)
	if err != nil { return "", err }
	req.Header.Set("X-Admin-Token", t.adminToken)
	res, err := t.client.Do(req)
	if err != nil { return "", err }
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("Switch returned HTTP %d; check ADMIN_TOKEN and HORIZON_LICENSE_PUBLIC_KEY", res.StatusCode)
	}
	var data hwidResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, 64*1024)).Decode(&data); err != nil { return "", err }
	if strings.TrimSpace(data.HardwareID) == "" {
		return "", errors.New("Switch returned no hardware_id")
	}
	return strings.TrimSpace(data.HardwareID), nil
}

func (t *tester) issue(ctx context.Context, hardwareID string) (licenseprotocol.SignedLicense, error) {
	userID := fmt.Sprintf("v14-e2e-%d", time.Now().UnixNano())
	payload := map[string]any{
		"product_id": "horizon-core",
		"user_id": userID,
		"volume": 1000,
		"duration": 1,
		"hardware_id": hardwareID,
	}
	data, err := json.Marshal(payload)
	if err != nil { return licenseprotocol.SignedLicense{}, err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.backendURL+"/api/v1/license/generate", bytes.NewReader(data))
	if err != nil { return licenseprotocol.SignedLicense{}, err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", t.apiKey)
	res, err := t.client.Do(req)
	if err != nil { return licenseprotocol.SignedLicense{}, err }
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 64*1024))
		return licenseprotocol.SignedLicense{}, fmt.Errorf("Backend returned HTTP %d: %s", res.StatusCode, safeErrorBody(b))
	}
	var out issueResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, 256*1024)).Decode(&out); err != nil {
		return licenseprotocol.SignedLicense{}, err
	}
	if out.License.Statement.LicenseID == "" || out.License.Signature == "" {
		return licenseprotocol.SignedLicense{}, errors.New("Backend response did not contain a signed v1.4 envelope")
	}
	return out.License, nil
}

func (t *tester) check(ctx context.Context, signed licenseprotocol.SignedLicense) (checkResponse, int, error) {
	data, err := json.Marshal(signed)
	if err != nil { return checkResponse{}, 0, err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.switchURL+"/api/v1/license/check", bytes.NewReader(data))
	if err != nil { return checkResponse{}, 0, err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", t.adminToken)
	res, err := t.client.Do(req)
	if err != nil { return checkResponse{}, 0, err }
	defer res.Body.Close()
	var out checkResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, 128*1024)).Decode(&out); err != nil {
		return out, res.StatusCode, err
	}
	return out, res.StatusCode, nil
}

func requireLocalURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil { return err }
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme in %q", raw)
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	isLocal := strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
	if !isLocal {
		return fmt.Errorf("refusing non-local target %q; set HORIZON_E2E_ALLOW_REMOTE=YES only after confirming this is a test environment", raw)
	}
	return nil
}

func flipLastHexDigit(s string) string {
	if len(s) == 0 { return s }
	last := s[len(s)-1]
	if last == '0' { return s[:len(s)-1] + "1" }
	return s[:len(s)-1] + "0"
}

func safeErrorBody(b []byte) string {
	var v map[string]any
	if json.Unmarshal(b, &v) == nil {
		if e, ok := v["error"].(string); ok { return e }
		return "request rejected"
	}
	return "request rejected"
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "FAIL:", err)
	os.Exit(1)
}
