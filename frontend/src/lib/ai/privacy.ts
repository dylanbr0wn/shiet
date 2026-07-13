export interface PrivacyFields {
  title: boolean;
  attendees: boolean;
  description: boolean;
  location: boolean;
}

export const DEFAULT_PRIVACY_FIELDS: PrivacyFields = {
  title: true,
  attendees: true,
  description: false,
  location: false,
};

export function formatPrivacySharingSummary(fields: PrivacyFields): string {
  const parts: string[] = [];
  if (fields.title) {
    parts.push("title");
  }
  if (fields.attendees) {
    parts.push("domains");
  }
  if (fields.description) {
    parts.push("description");
  }
  if (fields.location) {
    parts.push("location");
  }
  if (parts.length === 0) {
    return "category list only";
  }
  return parts.join("+");
}

export const AI_CLOUD_PRESETS = [
  {
    id: "openai",
    name: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    local: false,
  },
  {
    id: "anthropic",
    name: "Anthropic",
    baseUrl: "https://api.anthropic.com",
    local: false,
  },
] as const;
