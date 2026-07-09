import type { AlertChannel, AlertChannelKind } from "./api";

type ChannelInputMode = "email" | "numeric" | "text" | "tel" | "url";

export type ChannelField = {
  key: string;
  label: string;
  placeholder: string;
  hint?: string;
  inputMode?: ChannelInputMode;
  optional?: boolean;
  validate?: (value: string) => string | null;
};

export type ChannelKindDefinition = {
  value: AlertChannelKind;
  label: string;
  icon: string;
  fields: ChannelField[];
};

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
