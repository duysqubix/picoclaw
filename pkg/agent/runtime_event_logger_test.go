package agent

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
	runtimeevents "github.com/sipeed/picoclaw/pkg/events"
)

func TestRuntimeEventLoggerFiltering(t *testing.T) {
	cfg := config.DefaultConfig()
	eventLogger := newRuntimeEventLogger(cfg)
	if eventLogger == nil {
		t.Fatal("default runtime event logger is nil")
	}

	if !eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindAgentTurnStart,
		Severity: runtimeevents.SeverityInfo,
	}) {
		t.Fatal("default config should log agent events")
	}
	if eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindChannelLifecycleStarted,
		Severity: runtimeevents.SeverityInfo,
	}) {
		t.Fatal("default config should not log non-agent events")
	}

	cfg.Events.Logging.Include = []string{"*"}
	cfg.Events.Logging.Exclude = []string{"mcp.*"}
	eventLogger = newRuntimeEventLogger(cfg)
	if !eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindGatewayReady,
		Severity: runtimeevents.SeverityInfo,
	}) {
		t.Fatal("include * should log gateway events")
	}
	if eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindMCPServerConnected,
		Severity: runtimeevents.SeverityInfo,
	}) {
		t.Fatal("exclude mcp.* should suppress MCP events")
	}

	cfg.Events.Logging.Exclude = nil
	cfg.Events.Logging.MinSeverity = "warn"
	eventLogger = newRuntimeEventLogger(cfg)
	if eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindGatewayReady,
		Severity: runtimeevents.SeverityInfo,
	}) {
		t.Fatal("min severity warn should suppress info events")
	}
	if !eventLogger.shouldLog(runtimeevents.Event{
		Kind:     runtimeevents.KindGatewayReloadFailed,
		Severity: runtimeevents.SeverityError,
	}) {
		t.Fatal("min severity warn should allow error events")
	}

	cfg.Events.Logging.Enabled = false
	if newRuntimeEventLogger(cfg) != nil {
		t.Fatal("disabled config should not create runtime event logger")
	}
}

func TestRuntimeEventLogFieldsSummarizeAgentPayload(t *testing.T) {
	fields := runtimeEventLogFields(runtimeevents.Event{
		ID:       "evt-test",
		Kind:     runtimeevents.KindAgentToolExecStart,
		Severity: runtimeevents.SeverityInfo,
		Source: runtimeevents.Source{
			Component: "agent",
			Name:      "main",
		},
		Scope: runtimeevents.Scope{
			AgentID:    "main",
			SessionKey: "session-1",
			TurnID:     "turn-1",
		},
		Payload: ToolExecStartPayload{
			Tool: "exec",
			Arguments: map[string]any{
				"secret": "should-not-be-logged-by-default",
			},
		},
	})

	if fields["event_id"] != "evt-test" || fields["source_component"] != "agent" {
		t.Fatalf("missing common event fields: %#v", fields)
	}
	if fields["tool"] != "exec" || fields["args_count"] != 1 {
		t.Fatalf("missing safe agent payload summary fields: %#v", fields)
	}
	if _, ok := fields["payload"]; ok {
		t.Fatalf("raw payload should not be included by runtimeEventLogFields: %#v", fields)
	}
}
