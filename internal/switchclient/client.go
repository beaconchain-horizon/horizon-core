package switchclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SwitchClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type VerifyRequest struct {
	LicenseID string `json:"license_id"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
}

type VerifyResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

func NewSwitchClient(baseURL string) *SwitchClient {
	return &SwitchClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *SwitchClient) VerifyLicense(licenseID, signature, data string) (*VerifyResponse, error) {
	reqBody := VerifyRequest{LicenseID: licenseID, Signature: signature, Data: data}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := c.HTTPClient.Post(c.BaseURL+"/verify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("switch returned status %d", resp.StatusCode)
	}
	var result VerifyResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}
nano cmd/api/main.go
func verifyLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID string `json:"license_id"`
		Signature string `json:"signature"`
		Data      string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switchURL := os.Getenv("SWITCH_URL")
	if switchURL == "" {
		switchURL = "https://horizon-switch.liara.run"
	}
	client := switchclient.NewSwitchClient(switchURL)
	resp, err := client.VerifyLicense(req.LicenseID, req.Signature, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp.Verified {
		// ثبت تراکنش موفق در دیتابیس (اگر نیاز است)
		c.JSON(http.StatusOK, gin.H{"status": "verified", "message": resp.Message})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": resp.Message})
	}
}

