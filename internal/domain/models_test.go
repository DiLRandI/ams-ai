package domain

import (
	"testing"
	"time"
)

func TestWarrantyState(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	expired := now.AddDate(0, 0, -1)
	soon := now.AddDate(0, 0, 10)
	active := now.AddDate(0, 0, 45)

	tests := []struct {
		name   string
		expiry *time.Time
		want   string
	}{
		{name: "not set", expiry: nil, want: WarrantyNotSet},
		{name: "expired", expiry: &expired, want: WarrantyExpired},
		{name: "expiring soon", expiry: &soon, want: WarrantyExpiringSoon},
		{name: "active", expiry: &active, want: WarrantyActive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WarrantyState(tt.expiry, now, 30)
			if got != tt.want {
				t.Fatalf("WarrantyState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReminderState(t *testing.T) {
	now := time.Date(2026, 4, 19, 18, 0, 0, 0, time.UTC)
	if got := ReminderState(now.AddDate(0, 0, -1), now); got != ReminderOverdue {
		t.Fatalf("past reminder = %q", got)
	}
	if got := ReminderState(now, now); got != ReminderDue {
		t.Fatalf("same-day reminder = %q", got)
	}
	if got := ReminderState(now.AddDate(0, 0, 1), now); got != ReminderUpcoming {
		t.Fatalf("future reminder = %q", got)
	}
}

func TestAssetAccessAllowed(t *testing.T) {
	assignedID := int64(7)
	asset := Asset{CreatedBy: 5, AssignedUserID: &assignedID}

	if !AssetAccessAllowed(User{ID: 1, Role: RoleAdmin}, asset) {
		t.Fatal("admin should access all assets")
	}
	if !AssetAccessAllowed(User{ID: 5, Role: RoleUser}, asset) {
		t.Fatal("creator should access asset")
	}
	if !AssetAccessAllowed(User{ID: 7, Role: RoleUser}, asset) {
		t.Fatal("assigned user should access asset")
	}
	if AssetAccessAllowed(User{ID: 8, Role: RoleUser}, asset) {
		t.Fatal("unrelated standard user should not access asset")
	}
}
