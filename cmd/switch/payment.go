package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// PAYMENT GATEWAY (MOCK MODE)
// ============================================================
//
// In mock mode, the user is redirected to a local
// /payment/mock page which simulates a successful payment.
//
// When the company is registered and a real ZarinPal
// merchant ID is available, set:
//
//   ZARINPAL_MERCHANT_ID=<merchant-id>
//   ZARINPAL_MODE=real
//
// The code will then call the real ZarinPal API instead.
//
// The rest of the application does not change.
//
// ============================================================

const (
	paymentModeMock = "mock"
	paymentModeReal = "real"
)

// paymentMode returns the current payment mode.
func paymentMode() string {
	m := os.Getenv("ZARINPAL_MODE")
	if m == "" {
		return paymentModeMock
	}
	return m
}

// generateAuthority creates a fake payment authority for
// mock mode.
func generateAuthority() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "MOCK-" + hex.EncodeToString(b)
}

// ============================================================
// REQUEST PAYMENT
// ============================================================

// initiatePaymentHandler creates a payment request for an
// invoice.
//
// POST /api/v1/payment/initiate
// Body: {"invoice_id": "..."}
func initiatePaymentHandler(c *gin.Context) {
	var req struct {
		InvoiceID string `json:"invoice_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inv Invoice
	if err := db.Where("invoice_id = ?", req.InvoiceID).First(&inv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	if inv.Status == "paid" {
		c.JSON(http.StatusConflict, gin.H{"error": "invoice already paid"})
		return
	}

	if inv.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice is not pending"})
		return
	}

	mode := paymentMode()

	if mode == paymentModeMock {
		authority := generateAuthority()

		// Store authority on invoice as a pending marker
		db.Model(&Invoice{}).
			Where("invoice_id = ?", inv.InvoiceID).
			Update("payment_ref", authority)

		redirectURL := fmt.Sprintf(
			"/payment/mock?authority=%s&invoice_id=%s",
			authority, inv.InvoiceID,
		)

		addAudit("payment_initiate_mock", inv.InvoiceID,
			fmt.Sprintf("amount=%d", inv.AmountIRR),
			c.ClientIP())

		c.JSON(http.StatusOK, gin.H{
			"mode":         "mock",
			"authority":    authority,
			"invoice_id":   inv.InvoiceID,
			"amount_irr":   inv.AmountIRR,
			"amount_toman": inv.AmountIRR,
			"redirect_url": redirectURL,
			"message":      "MOCK MODE - no real payment will occur",
		})
		return
	}

	// Real mode placeholder - will be filled when company is registered
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "real ZarinPal integration not yet configured",
	})
}

// ============================================================
// MOCK PAYMENT PAGE
// ============================================================

// mockPaymentPageHandler serves a simulated payment page.
//
// GET /payment/mock?authority=...&invoice_id=...
func mockPaymentPageHandler(c *gin.Context) {
	authority := c.Query("authority")
	invoiceID := c.Query("invoice_id")

	if authority == "" || invoiceID == "" {
		c.String(http.StatusBadRequest, "missing authority or invoice_id")
		return
	}

	var inv Invoice
	if err := db.Where("invoice_id = ?", invoiceID).First(&inv).Error; err != nil {
		c.String(http.StatusNotFound, "invoice not found")
		return
	}

	html := `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="UTF-8">
<title>Mock Payment - Horizon</title>
<style>
body{font-family:Tahoma;background:#0f172a;color:#eef2ff;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}
.box{background:#1e293b;padding:3rem;border-radius:1.5rem;max-width:500px;width:90%;border:1px solid #334155}
h1{color:#facc15;margin-top:0}
.row{display:flex;justify-content:space-between;padding:.7rem 0;border-bottom:1px solid #334155}
.val{font-family:monospace;color:#4ade80;font-weight:700}
.btn{background:linear-gradient(135deg,#4ade80,#22c55e);color:#fff;border:none;padding:1rem;border-radius:1rem;width:100%;font-size:1rem;font-weight:700;cursor:pointer;margin-top:1.5rem}
.btn:hover{transform:translateY(-2px);box-shadow:0 15px 40px rgba(74,222,128,.4)}
.note{background:rgba(250,204,21,.1);border:1px solid #facc15;border-radius:.8rem;padding:1rem;margin-top:1.5rem;color:#facc15;font-size:.85rem}
</style>
</head>
<body>
<div class="box">
<h1>MOCK PAYMENT</h1>
<p>This is a simulation. No real payment will occur.</p>

<div class="row"><span>Invoice</span><span class="val">` + inv.InvoiceID + `</span></div>
<div class="row"><span>License</span><span class="val">` + inv.LicenseID + `</span></div>
<div class="row"><span>Duration</span><span class="val">` + fmt.Sprintf("%d hours", inv.DurationH) + `</span></div>
<div class="row"><span>Amount</span><span class="val">` + fmt.Sprintf("%d Toman", inv.AmountIRR) + `</span></div>

<form method="POST" action="/api/v1/payment/mock/confirm">
<input type="hidden" name="authority" value="` + authority + `">
<input type="hidden" name="invoice_id" value="` + invoiceID + `">
<button type="submit" class="btn">Confirm Payment (Mock)</button>
</form>

<div class="note">
When the company is registered, this page is replaced by the real ZarinPal gateway.
</div>
</div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ============================================================
// MOCK CONFIRM
// ============================================================

// mockConfirmHandler simulates the payment gateway calling
// back after a successful payment.
//
// POST /api/v1/payment/mock/confirm
func mockConfirmHandler(c *gin.Context) {
	authority := c.PostForm("authority")
	invoiceID := c.PostForm("invoice_id")

	if authority == "" || invoiceID == "" {
		c.String(http.StatusBadRequest, "missing authority or invoice_id")
		return
	}

	var inv Invoice
	if err := db.Where("invoice_id = ?", invoiceID).First(&inv).Error; err != nil {
		c.String(http.StatusNotFound, "invoice not found")
		return
	}

	if inv.PaymentRef != authority {
		c.String(http.StatusBadRequest, "authority mismatch")
		return
	}

	if inv.Status == "paid" {
		c.String(http.StatusOK, "already paid")
		return
	}

	// Mark invoice as paid
	now := time.Now().Unix()
	db.Model(&Invoice{}).
		Where("invoice_id = ?", invoiceID).
		Updates(map[string]interface{}{
			"status":      "paid",
			"payment_ref": authority,
			"paid_at":     now,
		})

	addAudit("payment_mock_paid", invoiceID,
		fmt.Sprintf("amount=%d", inv.AmountIRR),
		c.ClientIP())

	// Auto-issue renewed license
	newLic, licErr := issueRenewalLicense(&inv)
	if licErr != nil {
		addAudit("payment_mock_renew_fail", invoiceID, licErr.Error(), c.ClientIP())
		c.String(http.StatusInternalServerError, "payment ok but license issue failed: "+licErr.Error())
		return
	}
	addAudit("payment_mock_renew_ok", newLic.LicenseID, "renewed from "+inv.LicenseID, c.ClientIP())

	// Redirect to success page
	c.Redirect(http.StatusFound, "/payment/success?invoice_id="+invoiceID)
}

// ============================================================
// SUCCESS PAGE
// ============================================================

// paymentSuccessPageHandler shows a success page after
// payment confirmation.
//
// GET /payment/success?invoice_id=...
func paymentSuccessPageHandler(c *gin.Context) {
	invoiceID := c.Query("invoice_id")

	var inv Invoice
	if err := db.Where("invoice_id = ?", invoiceID).First(&inv).Error; err != nil {
		c.String(http.StatusNotFound, "invoice not found")
		return
	}

	html := `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="UTF-8">
<title>Payment Success - Horizon</title>
<style>
body{font-family:Tahoma;background:#0f172a;color:#eef2ff;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}
.box{background:#1e293b;padding:3rem;border-radius:1.5rem;max-width:500px;width:90%;border:1px solid #4ade80;text-align:center}
h1{color:#4ade80;font-size:3rem;margin:0}
p{color:#94a3b8;margin:1rem 0}
.btn{display:inline-block;background:linear-gradient(135deg,#4ade80,#22c55e);color:#fff;text-decoration:none;padding:1rem 2rem;border-radius:1rem;font-weight:700;margin-top:1.5rem}
</style>
</head>
<body>
<div class="box">
<h1>OK</h1>
<h2>Payment Confirmed</h2>
<p>Invoice: ` + inv.InvoiceID + `</p>
<p>License renewal is now active.</p>
<a href="/" class="btn">Back to Store</a>
</div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
