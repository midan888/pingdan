import { describe, expect, it } from "vitest";
import {
  CHANNEL_KINDS,
  channelIcon,
  channelKind,
  channelTarget,
  configForKind,
  validateChannelConfig,
} from "./channels";

describe("alert channel schema", () => {
  it("defines every supported channel kind once", () => {
    const kinds = CHANNEL_KINDS.map((kind) => kind.value);

    expect(kinds).toEqual([
      "email",
      "telegram",
      "slack",
      "discord",
      "teams",
      "webhook",
      "pagerduty",
      "ntfy",
      "pushover",
      "twilio_sms",
      "opsgenie",
    ]);
    expect(new Set(kinds).size).toBe(kinds.length);
  });

  it("falls back to the first channel definition and generic icon", () => {
    expect(channelKind("email").label).toBe("Email");
    expect(channelKind("not-real" as never).value).toBe("email");
    expect(channelIcon("not-real")).toBe("•");
  });

  it("validates required fields and provider-specific URLs", () => {
    expect(validateChannelConfig("email", { to: "" })).toBe("Email address is required.");
    expect(validateChannelConfig("email", { to: "ops@example.com" })).toBeNull();
    expect(validateChannelConfig("slack", { webhookUrl: "https://example.com" })).toContain(
      "https://hooks.slack.com/",
    );
    expect(validateChannelConfig("discord", { webhookUrl: "https://discordapp.com/api/webhooks/abc" })).toBeNull();
    expect(validateChannelConfig("teams", { webhookUrl: "http://logic.azure.com/hook" })).toBe(
      "Teams webhook URLs must use https.",
    );
    expect(validateChannelConfig("webhook", { url: "file:///tmp/hook" })).toBe("Enter a valid http or https URL.");
    expect(validateChannelConfig("twilio_sms", { to: "5551234567" })).toContain("E.164");
    expect(validateChannelConfig("twilio_sms", { to: "+15551234567" })).toBeNull();
  });

  it("validates option fields and applies default values", () => {
    expect(validateChannelConfig("opsgenie", { apiKey: "key" })).toBeNull();
    expect(validateChannelConfig("opsgenie", { apiKey: "key", region: "apac" })).toBe(
      "Choose a valid region.",
    );

    expect(configForKind("opsgenie", { apiKey: " key " })).toEqual({ apiKey: "key", region: "us" });
    expect(configForKind("webhook", { url: " https://example.com/hook ", secret: " " })).toEqual({
      url: "https://example.com/hook",
    });
  });

  it("uses only non-sensitive config fields for display targets", () => {
    expect(channelTarget({ kind: "webhook", config: { url: "https://example.com/hook", secret: "hidden" } })).toBe(
      "https://example.com/hook",
    );
    expect(channelTarget({ kind: "pagerduty", config: { routingKey: "secret" } })).toBe("Configured");
    expect(channelTarget({ kind: "ntfy", config: { topic: "ops", server: "https://ntfy.sh" } })).toBe(
      "ops · https://ntfy.sh",
    );
  });
});
