import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import {
  PeriodService,
  type Period as WirePeriod,
} from "@/gen/shiet/app/v1/period_pb";
import type { Period } from "./types";

let client: Client<typeof PeriodService> | undefined;

function periodClient() {
  client ??= createClient(
    PeriodService,
    createConnectTransport({ baseUrl: rpcBaseUrl() }),
  );
  return client;
}

function rpcBaseUrl() {
  const configured = import.meta.env.VITE_SHIET_RPC_BASE_URL?.trim();
  if (configured) {
    return configured.replace(/\/$/, "");
  }
  const current = new URL(window.location.href);
  return `${current.protocol}//${current.host}/rpc`;
}

export async function listPeriodsRPC() {
  const response = await periodClient().listPeriods({});
  return response.periods.map(mapPeriod);
}

export async function ensureCurrentPeriodRPC(today: string, ianaTz: string) {
  const response = await periodClient().ensureCurrentPeriod({ today, ianaTz });
  return response.period ? mapPeriod(response.period) : null;
}

export function mapPeriod(period: WirePeriod): Period {
  const id = Number(period.id);
  if (!Number.isSafeInteger(id)) {
    throw new Error(`period id ${period.id} is outside JavaScript's safe integer range`);
  }
  return {
    id,
    startDate: period.startDate,
    endDate: period.endDate,
    cadence: period.cadence,
    anchorDate: period.anchorDate,
    targetHoursPerDay: period.targetHoursPerDay,
    ...(period.lastSyncedAt
      ? {
          lastSyncedAt: new Date(
            Number(period.lastSyncedAt.seconds) * 1_000 +
              period.lastSyncedAt.nanos / 1_000_000,
          ).toISOString(),
        }
      : {}),
  };
}
