// @vitest-environment jsdom

import { describe, expect, it } from "vitest";
import { integrationRegistry } from "./registry";

describe("integrationRegistry", () => {
  it("registers Google with ConfigSlot only", () => {
    const google = integrationRegistry.find((entry) => entry.id === "google");
    expect(google).toBeDefined();
    expect(google?.ConfigSlot).toBeDefined();
    expect(google?.Panel).toBeUndefined();
    expect(google?.kind).toBe("calendar_source");
  });

  it("registers GitHub and Slack with Panel only", () => {
    for (const id of ["github", "slack"] as const) {
      const entry = integrationRegistry.find((item) => item.id === id);
      expect(entry).toBeDefined();
      expect(entry?.Panel).toBeDefined();
      expect(entry?.ConfigSlot).toBeUndefined();
      expect(entry?.kind).toBe("activity_evidence");
    }
  });
});
