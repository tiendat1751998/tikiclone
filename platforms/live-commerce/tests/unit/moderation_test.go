package unit

import (
	"testing"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/moderation"
)

func TestFilterSpam(t *testing.T) {
	f := moderation.NewFilter()
	if !f.IsSpam("check out http://spam.com") {
		t.Error("expected http link to be spam")
	}
	if !f.IsSpam("buy now cheap stuff") {
		t.Error("expected 'buy now' to be spam")
	}
	if !f.IsSpam("click here for free money") {
		t.Error("expected 'click here' to be spam")
	}
	if f.IsSpam("hello, how are you?") {
		t.Error("expected normal message not to be spam")
	}
}

func TestValidateContent(t *testing.T) {
	f := moderation.NewFilter()
	valid, reason := f.ValidateContent("hello world")
	if !valid {
		t.Errorf("expected valid, got reason: %s", reason)
	}
	valid, reason = f.ValidateContent("")
	if valid {
		t.Error("expected empty to be invalid")
	}
	if reason != "empty_message" {
		t.Errorf("expected empty_message, got %s", reason)
	}
	spamMsg := ""
	for i := 0; i < 600; i++ {
		spamMsg += "a"
	}
	valid, reason = f.ValidateContent(spamMsg)
	if valid {
		t.Error("expected long message to be invalid")
	}
}

func TestAddSpamPattern(t *testing.T) {
	f := moderation.NewFilter()
	f.AddSpamPattern("discord.gg")
	if !f.IsSpam("join discord.gg/cheapstuff") {
		t.Error("expected custom pattern to be detected")
	}
}

func TestBannedWords(t *testing.T) {
	f := moderation.NewFilter()
	f.AddBannedWord("badword")
	has, word := f.ContainsBannedWords("this contains a badword here")
	if !has {
		t.Error("expected banned word to be detected")
	}
	if word != "badword" {
		t.Errorf("expected 'badword', got %s", word)
	}
}

func TestModerationQueue(t *testing.T) {
	q := moderation.NewQueue()
	if q.PendingCount() != 0 {
		t.Error("expected empty queue")
	}
	action := &domain.ModerationAction{
		ID: "test", RoomID: "room1", UserID: "user1",
		Action: "mute", Reason: "spam", ModeratedBy: "mod1",
	}
	q.Enqueue(action)
	if q.PendingCount() != 1 {
		t.Errorf("expected 1 item, got %d", q.PendingCount())
	}
	result := q.Dequeue()
	if result == nil {
		t.Fatal("expected non-nil dequeue")
	}
	if result.ID != "test" {
		t.Errorf("expected id 'test', got %s", result.ID)
	}
}
