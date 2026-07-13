import { describe, expect, it } from "vitest";
import {
  DEFAULT_PRIVACY_FIELDS,
  formatPrivacySharingSummary,
} from "./privacy";

describe("formatPrivacySharingSummary", () => {
  it("summarizes default cloud fields", () => {
    expect(formatPrivacySharingSummary(DEFAULT_PRIVACY_FIELDS)).toBe(
      "title+domains",
    );
  });

  it("handles empty sharing", () => {
    expect(
      formatPrivacySharingSummary({
        title: false,
        attendees: false,
        description: false,
        location: false,
      }),
    ).toBe("category list only");
  });
});
