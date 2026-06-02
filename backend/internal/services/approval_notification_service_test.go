package services

import (
	"testing"

	"sts-backend/internal/config"
)

type fakeApprovalNotifier struct {
	requestRecipients []string
	requestTypes      []string
	requestDetails    [][]string
	decisionRecipient string
	decisionType      string
	decisionStatus    string
	decisionDetails   []string
}

func (f *fakeApprovalNotifier) SendApprovalRequestNotification(recipient, requestType string, details ...string) error {
	f.requestRecipients = append(f.requestRecipients, recipient)
	f.requestTypes = append(f.requestTypes, requestType)
	f.requestDetails = append(f.requestDetails, append([]string(nil), details...))
	return nil
}

func (f *fakeApprovalNotifier) SendApprovalDecisionNotification(recipient, requestType, decision string, details ...string) error {
	f.decisionRecipient = recipient
	f.decisionType = requestType
	f.decisionStatus = decision
	f.decisionDetails = append([]string(nil), details...)
	return nil
}

func TestNormalizeApprovalNotificationPhone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty uses default", in: "", want: defaultApprovalNotificationPhone},
		{name: "local number becomes msisdn", in: "077 123 4567", want: "94771234567"},
		{name: "plus sign removed", in: "+94 77 123 4567", want: "94771234567"},
		{name: "already normalized stays same", in: "94715342627", want: "94715342627"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeApprovalNotificationPhone(tt.in); got != tt.want {
				t.Fatalf("normalizeApprovalNotificationPhone(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeApprovalNotificationPhones(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty uses defaults", in: "", want: []string{"94715342627", "94772945875"}},
		{name: "comma separated", in: "94715342627,0772945875", want: []string{"94715342627", "94772945875"}},
		{name: "semicolon separated", in: "94715342627; +94 77 294 5875", want: []string{"94715342627", "94772945875"}},
		{name: "duplicates removed", in: "94715342627, 071 534 2627", want: []string{"94715342627"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeApprovalNotificationPhones(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("normalizeApprovalNotificationPhones(%q) length = %d, want %d (%v)", tt.in, len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("normalizeApprovalNotificationPhones(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNotifyApprovalRequestUsesConfiguredAdminPhone(t *testing.T) {
	origLoadConfig := loadApprovalNotificationConfig
	origNewNotifier := newApprovalSMSNotifier
	t.Cleanup(func() {
		loadApprovalNotificationConfig = origLoadConfig
		newApprovalSMSNotifier = origNewNotifier
	})

	fakeNotifier := &fakeApprovalNotifier{}
	loadApprovalNotificationConfig = func() *config.Config {
		return &config.Config{ApprovalNotificationPhone: " 077 123 4567 "}
	}
	newApprovalSMSNotifier = func(cfg *config.Config) approvalSMSNotifier {
		return fakeNotifier
	}

	notifyApprovalRequest("bus owner", "Company: Example", "Email: example@example.com")

	if len(fakeNotifier.requestRecipients) != 1 {
		t.Fatalf("request recipients length = %d, want 1", len(fakeNotifier.requestRecipients))
	}
	if got, want := fakeNotifier.requestRecipients[0], "94771234567"; got != want {
		t.Fatalf("request recipient = %q, want %q", got, want)
	}
	if got, want := fakeNotifier.requestTypes[0], "bus owner"; got != want {
		t.Fatalf("request type = %q, want %q", got, want)
	}
	if len(fakeNotifier.requestDetails[0]) != 2 {
		t.Fatalf("request details length = %d, want 2", len(fakeNotifier.requestDetails[0]))
	}
}

func TestNotifyApprovalRequestUsesDefaultAdminPhones(t *testing.T) {
	origLoadConfig := loadApprovalNotificationConfig
	origNewNotifier := newApprovalSMSNotifier
	t.Cleanup(func() {
		loadApprovalNotificationConfig = origLoadConfig
		newApprovalSMSNotifier = origNewNotifier
	})

	fakeNotifier := &fakeApprovalNotifier{}
	loadApprovalNotificationConfig = func() *config.Config {
		return &config.Config{}
	}
	newApprovalSMSNotifier = func(cfg *config.Config) approvalSMSNotifier {
		return fakeNotifier
	}

	notifyApprovalRequest("driver", "Name: Example Driver")

	wantRecipients := []string{"94715342627", "94772945875"}
	if len(fakeNotifier.requestRecipients) != len(wantRecipients) {
		t.Fatalf("request recipients length = %d, want %d", len(fakeNotifier.requestRecipients), len(wantRecipients))
	}
	for i, want := range wantRecipients {
		if got := fakeNotifier.requestRecipients[i]; got != want {
			t.Fatalf("request recipient[%d] = %q, want %q", i, got, want)
		}
	}
	for i := range wantRecipients {
		if got, want := fakeNotifier.requestTypes[i], "driver"; got != want {
			t.Fatalf("request type[%d] = %q, want %q", i, got, want)
		}
	}
}

func TestPendingApprovalStatusHelpers(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "empty defaults to pending", status: "", want: true},
		{name: "pending with whitespace", status: " Pending ", want: true},
		{name: "approved does not notify", status: "approved", want: false},
		{name: "verified does not notify", status: "verified", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPendingApprovalStatus(tt.status); got != tt.want {
				t.Fatalf("isPendingApprovalStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
