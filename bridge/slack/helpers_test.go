package bslack

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/mspgeek-community/matterbridge/bridge"
	"github.com/mspgeek-community/matterbridge/bridge/config"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestExtractTopicOrPurpose(t *testing.T) {
	testcases := map[string]struct {
		input          string
		wantChangeType string
		wantOutput     string
	}{
		"success - topic type":   {"@someone set channel topic: foo bar", "topic", "foo bar"},
		"success - purpose type": {"@someone set channel purpose: foo bar", "purpose", "foo bar"},
		"success - one line":     {"@someone set channel topic: foo bar", "topic", "foo bar"},
		"success - multi-line":   {"@someone set channel topic: foo\nbar", "topic", "foo\nbar"},
		"success - cleared":      {"@someone cleared channel topic", "topic", ""},
		"error - unhandled":      {"some unmatched message", "unknown", ""},
	}

	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	cfg := &bridge.Config{Bridge: &bridge.Bridge{Log: logrus.NewEntry(logger)}}
	b := newBridge(cfg)
	for name, tc := range testcases {
		gotChangeType, gotOutput := b.extractTopicOrPurpose(tc.input)

		assert.Equalf(t, tc.wantChangeType, gotChangeType, "This testcase failed: %s", name)
		assert.Equalf(t, tc.wantOutput, gotOutput, "This testcase failed: %s", name)
	}
}

func TestIsFreshEdit(t *testing.T) {
	ts := func(d time.Duration) *slack.Edited {
		return &slack.Edited{Timestamp: fmt.Sprintf("%d.000100", time.Now().Add(-d).Unix())}
	}
	assert.False(t, isFreshEdit(nil))
	assert.False(t, isFreshEdit(&slack.Edited{}))
	assert.False(t, isFreshEdit(&slack.Edited{Timestamp: "garbage"}))
	assert.True(t, isFreshEdit(ts(10*time.Second)))
	assert.False(t, isFreshEdit(ts(config.EditMaxAge+time.Minute)))
	assert.False(t, isFreshEdit(ts(30*24*time.Hour)))
}
