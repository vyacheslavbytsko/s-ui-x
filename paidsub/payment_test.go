package paidsub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/deposist/s-ui-x/database"
	"github.com/deposist/s-ui-x/database/model"
)

func TestApplyPaidOrderIdempotentRenewal(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}

	// Disabled, expired-by-default client with usage, no inbounds (no restart).
	client := model.Client{
		Enable:    false,
		Name:      "tg42",
		Inbounds:  json.RawMessage("[]"),
		Volume:    0,
		Expiry:    0,
		Up:        100,
		Down:      200,
		TotalUp:   0,
		TotalDown: 0,
	}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}

	tariff := Tariff{Name: "Month", Price: 10000, Currency: "RUB", AddDays: 30, AddTrafficBytes: 1 << 30, Enabled: true}
	if err := db.Create(&tariff).Error; err != nil {
		t.Fatalf("create tariff: %v", err)
	}

	order := PaymentOrder{
		ClientId: client.Id, TariffId: tariff.Id, Provider: "yookassa",
		Amount: 10000, Currency: "RUB", Status: StatusPending,
		TelegramUserId: 42, IdempotencyKey: "key-1", CreatedAt: time.Now().Unix(),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	ps := NewPaymentService()
	applied, tgID, err := ps.ApplyPaidOrder(order.Id, "charge-1", nil)
	if err != nil {
		t.Fatalf("ApplyPaidOrder: %v", err)
	}
	if !applied {
		t.Fatal("expected first apply to succeed")
	}
	if tgID != 42 {
		t.Fatalf("expected tgID 42, got %d", tgID)
	}

	var got model.Client
	db.Where("id = ?", client.Id).First(&got)
	if !got.Enable {
		t.Error("client should be re-enabled")
	}
	if got.Volume != 1<<30 {
		t.Errorf("volume = %d, want %d", got.Volume, int64(1<<30))
	}
	if got.Up != 0 || got.Down != 0 {
		t.Errorf("up/down should reset, got up=%d down=%d", got.Up, got.Down)
	}
	if got.TotalUp != 100 || got.TotalDown != 200 {
		t.Errorf("totals = %d/%d, want 100/200", got.TotalUp, got.TotalDown)
	}
	now := time.Now().Unix()
	if got.Expiry < now+29*86400 || got.Expiry > now+31*86400 {
		t.Errorf("expiry not extended ~30d: %d (now %d)", got.Expiry, now)
	}

	var paidOrder PaymentOrder
	db.Where("id = ?", order.Id).First(&paidOrder)
	if paidOrder.Status != StatusPaid || paidOrder.ProviderChargeID != "charge-1" {
		t.Errorf("order not marked paid: %+v", paidOrder)
	}

	// Second apply must be an idempotent no-op (no double renewal).
	applied2, _, err := ps.ApplyPaidOrder(order.Id, "charge-1", nil)
	if err != nil {
		t.Fatalf("second ApplyPaidOrder: %v", err)
	}
	if applied2 {
		t.Fatal("second apply must be a no-op")
	}
	var got2 model.Client
	db.Where("id = ?", client.Id).First(&got2)
	if got2.Volume != 1<<30 {
		t.Errorf("volume changed on replay: %d", got2.Volume)
	}
	if got2.Expiry != got.Expiry {
		t.Errorf("expiry changed on replay: %d != %d", got2.Expiry, got.Expiry)
	}
}

func TestApplyPaidOrderRejectsZeroPriceTariff(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	client := model.Client{Enable: false, Name: "tg99", Inbounds: json.RawMessage("[]"), Expiry: 100}
	db.Create(&client)
	// Price 0 and StarsAmount 0 → must never grant a renewal.
	tariff := Tariff{Name: "Free", Price: 0, StarsAmount: 0, Currency: "RUB", AddDays: 30, Enabled: true}
	db.Create(&tariff)
	order := PaymentOrder{ClientId: client.Id, TariffId: tariff.Id, Provider: "yookassa", Amount: 0, Currency: "RUB", Status: StatusPending, IdempotencyKey: "zero"}
	db.Create(&order)

	ps := NewPaymentService()
	applied, _, err := ps.ApplyPaidOrder(order.Id, "c", nil)
	if err == nil {
		t.Fatal("expected error for zero-price tariff")
	}
	if applied {
		t.Fatal("zero-price tariff must not apply a renewal")
	}
	// Transaction rolled back: order stays pending, client not renewed.
	var o PaymentOrder
	db.Where("id = ?", order.Id).First(&o)
	if o.Status != StatusPending {
		t.Errorf("order should remain pending after rejected apply, got %s", o.Status)
	}
	var c model.Client
	db.Where("id = ?", client.Id).First(&c)
	if c.Enable || c.Expiry != 100 {
		t.Errorf("client must be unchanged, got enable=%v expiry=%d", c.Enable, c.Expiry)
	}
}

func TestExpireStaleOrders(t *testing.T) {
	db := openTestDB(t)
	if err := EnsureSchema(db); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	now := time.Now().Unix()
	stale := PaymentOrder{ClientId: 1, TariffId: 1, Provider: "cryptobot", Amount: 1, Currency: "RUB", Status: StatusPending, IdempotencyKey: "stale", ExpiresAt: now - 10}
	fresh := PaymentOrder{ClientId: 1, TariffId: 1, Provider: "cryptobot", Amount: 1, Currency: "RUB", Status: StatusPending, IdempotencyKey: "fresh", ExpiresAt: now + 3600}
	db.Create(&stale)
	db.Create(&fresh)

	ps := NewPaymentService()
	if err := ps.ExpireStaleOrders(); err != nil {
		t.Fatalf("ExpireStaleOrders: %v", err)
	}
	var s, f PaymentOrder
	db.Where("idempotency_key = ?", "stale").First(&s)
	db.Where("idempotency_key = ?", "fresh").First(&f)
	if s.Status != StatusExpired {
		t.Errorf("stale order not expired: %s", s.Status)
	}
	if f.Status != StatusPending {
		t.Errorf("fresh order should stay pending: %s", f.Status)
	}
	_ = database.GetDB()
}
