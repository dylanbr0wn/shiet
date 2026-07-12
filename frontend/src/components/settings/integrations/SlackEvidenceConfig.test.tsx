// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SlackEvidenceConfig } from "./SlackEvidenceConfig";

const mocks = vi.hoisted(() => ({
  setChannelSelectedMutate: vi.fn(),
  connections: [] as Array<{ provider: string; accountId: string; accountLabel: string }>,
  channels: [] as Array<{
    id: number;
    accountId: string;
    name: string;
    private: boolean;
    selected: boolean;
  }>,
  channelsLoading: false,
}));

vi.mock("@/lib/api", () => ({
  useIntegrationConnections: () => ({
    data: mocks.connections,
    isLoading: false,
  }),
  useSlackChannels: () => ({
    data: mocks.channels,
    isLoading: mocks.channelsLoading,
  }),
  useSetSlackChannelSelected: () => ({
    isPending: false,
    mutateAsync: mocks.setChannelSelectedMutate,
  }),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  }

  return Wrapper;
}

describe("SlackEvidenceConfig", () => {
  beforeEach(() => {
    cleanup();
    mocks.connections = [];
    mocks.channels = [];
    mocks.channelsLoading = false;
    mocks.setChannelSelectedMutate.mockReset();
  });

  it("shows connect prompt when no Slack workspace is connected", () => {
    render(<SlackEvidenceConfig />, { wrapper: createWrapper() });

    expect(
      screen.getByText("Connect a Slack workspace to see channels here."),
    ).toBeTruthy();
  });

  it("shows refresh hint when connected but no channels", () => {
    mocks.connections = [{ provider: "slack", accountId: "T1", accountLabel: "Acme" }];

    render(<SlackEvidenceConfig />, { wrapper: createWrapper() });

    expect(
      screen.getByText(
        "No channels found. Try Refresh channels on the connection above.",
      ),
    ).toBeTruthy();
  });

  it("toggles channel track selection", () => {
    mocks.connections = [{ provider: "slack", accountId: "T1", accountLabel: "Acme" }];
    mocks.channels = [
      {
        id: 42,
        accountId: "T1",
        name: "general",
        private: false,
        selected: false,
      },
    ];

    render(<SlackEvidenceConfig />, { wrapper: createWrapper() });

    fireEvent.click(
      screen.getByRole("button", { name: "Track #general as evidence" }),
    );

    expect(mocks.setChannelSelectedMutate).toHaveBeenCalledWith({
      channelID: 42,
      selected: true,
    });
  });
});
