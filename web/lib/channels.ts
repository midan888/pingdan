import type { AlertChannel, AlertChannelKind } from "./api";

type ChannelInputMode = "email" | "numeric" | "text" | "tel" | "url";

export type ChannelField = {
  key: string;
  label: string;
  placeholder: string;
  hint?: string;
  inputMode?: ChannelInputMode;
  optional?: boolean;
  sensitive?: boolean;
  validate?: (value: string) => string | null;
};

export type ChannelKindDefinition = {
  value: AlertChannelKind;
  label: string;
  icon: string;
  fields: ChannelField[];
};

function isHTTPURL(value: string): boolean {
  try {
    const u = new URL(value);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

function isHTTPSURL(value: string): boolean {
  try {
    return new URL(value).protocol === "https:";
  } catch {
    return false;
  }
}

export const CHANNEL_KINDS: ChannelKindDefinition[] = [
  {
    value: "email",
    label: "Email",
    icon: "✉",
    fields: [
      {
        key: "to",
        label: "Email address",
        placeholder: "oncall@company.com",
        hint: "We'll send alerts to this address when an endpoint goes down or recovers.",
        inputMode: "email",
        validate: (value) =>
          /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) ? null : "Enter a valid email address.",
      },
    ],
  },
  {
    value: "telegram",
    label: "Telegram",
    icon: "✈",
    fields: [
      {
        key: "chatId",
        label: "Telegram chat ID",
        placeholder: "123456789",
        hint: "Message @userinfobot on Telegram to get your chat ID, then add our bot to that chat.",
        inputMode: "numeric",
        validate: (value) => (/^-?\d+$/.test(value) ? null : "Chat ID should be a number (e.g. 123456789)."),
      },
    ],
  },
  {
    value: "slack",
    label: "Slack",
    icon: "#",
    fields: [
      {
        key: "webhookUrl",
        label: "Slack webhook URL",
        placeholder: "https://hooks.slack.com/services/...",
        hint: "Create an incoming webhook in Slack and paste its URL here.",
        inputMode: "url",
        validate: (value) =>
          value.startsWith("https://hooks.slack.com/")
            ? null
            : "Slack webhook URLs should start with https://hooks.slack.com/.",
      },
    ],
  },
  {
    value: "discord",
    label: "Discord",
    icon: "D",
    fields: [
      {
        key: "webhookUrl",
        label: "Discord webhook URL",
        placeholder: "https://discord.com/api/webhooks/...",
        hint: "Create a Discord channel webhook and paste its URL here.",
        inputMode: "url",
        validate: (value) =>
          value.startsWith("https://discord.com/api/webhooks/") ||
          value.startsWith("https://discordapp.com/api/webhooks/")
            ? null
            : "Discord webhook URLs should start with https://discord.com/api/webhooks/.",
      },
    ],
  },
  {
    value: "teams",
    label: "Teams",
    icon: "T",
    fields: [
      {
        key: "webhookUrl",
        label: "Power Automate webhook URL",
        placeholder: "https://...logic.azure.com/...",
        hint: "Use a Microsoft Teams workflow trigger URL from Power Automate.",
        inputMode: "url",
        validate: (value) => (isHTTPSURL(value) ? null : "Teams webhook URLs must use https."),
      },
    ],
  },
  {
    value: "webhook",
    label: "Webhook",
    icon: "{}",
    fields: [
      {
        key: "url",
        label: "Webhook URL",
        placeholder: "https://example.com/pingdan-alerts",
        hint: "We'll POST the full structured alert payload to this URL.",
        inputMode: "url",
        validate: (value) => (isHTTPURL(value) ? null : "Enter a valid http or https URL."),
      },
      {
        key: "secret",
        label: "Signing secret",
        placeholder: "Optional shared secret",
        hint: "When set, requests include an X-Pingdan-Signature HMAC header.",
        optional: true,
        sensitive: true,
      },
    ],
  },
];

export function channelKind(kind: AlertChannelKind): ChannelKindDefinition {
  return CHANNEL_KINDS.find((k) => k.value === kind) ?? CHANNEL_KINDS[0];
}

export function channelIcon(kind: AlertChannelKind | string): string {
  return CHANNEL_KINDS.find((k) => k.value === kind)?.icon ?? "•";
}

export function channelTarget(channel: Pick<AlertChannel, "kind" | "config">): string {
  const def = channelKind(channel.kind);
  return def.fields
    .filter((field) => !field.sensitive)
    .map((field) => channel.config[field.key])
    .filter((value): value is string => typeof value === "string" && value.length > 0)
    .join(" · ");
}

export function configForKind(kind: AlertChannelKind, values: Record<string, string>): Record<string, string> {
  const config: Record<string, string> = {};
  for (const field of channelKind(kind).fields) {
    const value = (values[field.key] ?? "").trim();
    if (value || !field.optional) config[field.key] = value;
  }
  return config;
}

export function validateChannelConfig(kind: AlertChannelKind, values: Record<string, string>): string | null {
  for (const field of channelKind(kind).fields) {
    const value = (values[field.key] ?? "").trim();
    if (!field.optional && !value) return `${field.label} is required.`;
    if (value && field.validate) {
      const error = field.validate(value);
      if (error) return error;
    }
  }
  return null;
}
