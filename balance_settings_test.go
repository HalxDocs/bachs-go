package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const balanceSettingsExample = `{
	"schedule_by_currency": {
		"NGN": {
			"currency": "NGN",
			"payout_currency": "NGN",
			"interval": "weekly",
			"weekly_payout_days": ["monday", "thursday"],
			"monthly_payout_days": null,
			"anchor_hour_utc": 9,
			"minimum_amount": "5000.00",
			"next_run_at": "2026-08-13T09:00:00.000Z",
			"last_run_at": null,
			"last_withdrawal_id": null,
			"disabled_reason": null
		}
	}
}`

func TestGetBalanceSettings(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/balance_settings" {
			t.Errorf("path = %q, want /v1/balance_settings", r.URL.Path)
		}
		io.WriteString(w, balanceSettingsExample)
	})

	settings, _, err := c.Misc.GetBalanceSettings(context.Background())
	if err != nil {
		t.Fatalf("GetBalanceSettings returned error: %v", err)
	}
	sched, ok := settings.ScheduleByCurrency["NGN"]
	if !ok {
		t.Fatalf("ScheduleByCurrency = %v, want an NGN entry", settings.ScheduleByCurrency)
	}
	if sched.Interval != PayoutScheduleIntervalWeekly {
		t.Errorf("Interval = %q, want weekly", sched.Interval)
	}
	if len(sched.WeeklyPayoutDays) != 2 || sched.WeeklyPayoutDays[0] != "monday" {
		t.Errorf("WeeklyPayoutDays = %v", sched.WeeklyPayoutDays)
	}
	if sched.AnchorHourUTC != 9 {
		t.Errorf("AnchorHourUTC = %d, want 9", sched.AnchorHourUTC)
	}
	if sched.MinimumAmount == nil || *sched.MinimumAmount != "5000.00" {
		t.Errorf("MinimumAmount = %v", sched.MinimumAmount)
	}
	if sched.NextRunAt == nil {
		t.Error("NextRunAt is nil")
	}
	if sched.MonthlyPayoutDays != nil {
		t.Errorf("MonthlyPayoutDays = %v, want nil", sched.MonthlyPayoutDays)
	}
}

func TestUpdateBalanceSettings(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/balance_settings" {
			t.Errorf("path = %q, want /v1/balance_settings", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		schedules, ok := got["schedule_by_currency"].(map[string]any)
		if !ok {
			t.Fatalf("schedule_by_currency missing: %s", body)
		}
		ngn, ok := schedules["NGN"].(map[string]any)
		if !ok {
			t.Fatalf("NGN schedule missing: %s", body)
		}
		if ngn["interval"] != "weekly" {
			t.Errorf("interval = %v, want weekly", ngn["interval"])
		}
		if _, present := ngn["monthly_payout_days"]; present {
			t.Errorf("monthly_payout_days should be omitted for a weekly schedule: %s", body)
		}
		io.WriteString(w, balanceSettingsExample)
	})

	anchor := 9
	settings, _, err := c.Misc.UpdateBalanceSettings(context.Background(), UpdateBalanceSettingsRequest{
		ScheduleByCurrency: map[string]UpdatePayoutScheduleRequest{
			"NGN": {
				Interval:         PayoutScheduleIntervalWeekly,
				WeeklyPayoutDays: []string{"monday", "thursday"},
				AnchorHourUTC:    &anchor,
				MinimumAmount:    strptr("5000.00"),
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateBalanceSettings returned error: %v", err)
	}
	if _, ok := settings.ScheduleByCurrency["NGN"]; !ok {
		t.Errorf("ScheduleByCurrency = %v", settings.ScheduleByCurrency)
	}
}

func strptr(s string) *string { return &s }
