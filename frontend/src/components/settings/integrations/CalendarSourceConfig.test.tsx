// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CalendarSourceConfig } from "./CalendarSourceConfig";

const mocks = vi.hoisted(() => ({
  setCalendarSelectedMutate: vi.fn(),
  setCalendarDefaultCategoryMutate: vi.fn(),
  connections: [] as Array<{ provider: string }>,
  calendars: [] as Array<{
    id: number;
    provider: string;
    name: string;
    isPrimary: boolean;
    selected: boolean;
    defaultCategoryId?: number;
  }>,
  categories: [{ id: 1, name: "Work" }],
}));

vi.mock("@/lib/api", () => ({
  useIntegrationConnections: () => ({
    data: mocks.connections,
    isLoading: false,
  }),
  useCalendars: () => ({
    data: mocks.calendars,
    isLoading: false,
  }),
  useCategories: () => ({
    data: mocks.categories,
    isLoading: false,
  }),
  useSetCalendarSelected: () => ({
    isPending: false,
    mutateAsync: mocks.setCalendarSelectedMutate,
  }),
  useSetCalendarDefaultCategory: () => ({
    isPending: false,
    mutateAsync: mocks.setCalendarDefaultCategoryMutate,
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

describe("CalendarSourceConfig", () => {
  beforeEach(() => {
    cleanup();
    Element.prototype.scrollIntoView = vi.fn();
    mocks.connections = [];
    mocks.calendars = [];
    mocks.setCalendarSelectedMutate.mockReset();
    mocks.setCalendarDefaultCategoryMutate.mockReset();
  });

  it("shows connect prompt when no Google account is connected", () => {
    render(<CalendarSourceConfig />, { wrapper: createWrapper() });

    expect(
      screen.getByText("Connect a Google account to see calendars here."),
    ).toBeTruthy();
  });

  it("toggles calendar import selection", () => {
    mocks.connections = [{ provider: "google" }];
    mocks.calendars = [
      {
        id: 42,
        provider: "google",
        name: "Work Calendar",
        isPrimary: true,
        selected: false,
      },
    ];

    render(<CalendarSourceConfig />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByRole("button", { name: "Import Work Calendar" }));

    expect(mocks.setCalendarSelectedMutate).toHaveBeenCalledWith({
      calendarID: 42,
      selected: true,
    });
  });

  it("clears default category when No default is selected", () => {
    mocks.connections = [{ provider: "google" }];
    mocks.calendars = [
      {
        id: 7,
        provider: "google",
        name: "Personal",
        isPrimary: false,
        selected: true,
        defaultCategoryId: 1,
      },
    ];

    render(<CalendarSourceConfig />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByLabelText("Default category for Personal"));
    fireEvent.click(screen.getByRole("option", { name: "No default" }));

    expect(mocks.setCalendarDefaultCategoryMutate).toHaveBeenCalledWith({
      calendarID: 7,
      categoryID: null,
    });
  });
});
