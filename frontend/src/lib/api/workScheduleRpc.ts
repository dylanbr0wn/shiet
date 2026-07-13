import { createClient, type Client } from "@connectrpc/connect";

import {
  WorkScheduleService,
  type ExpectedTime as WireExpectedTime,
  type ScheduleException as WireScheduleException,
  type WorkSchedule as WireWorkSchedule,
  type WorkingWindow as WireWorkingWindow,
  type WorkScheduleDay as WireWorkScheduleDay,
} from "@/gen/shiet/app/v1/work_schedule_pb";
import type {
  ExpectedTime,
  ScheduleException,
  WorkSchedule,
  WorkScheduleDay,
  WorkingWindow,
} from "./types";
import { rpcTransport } from "./rpcTransport";

let client: Client<typeof WorkScheduleService> | undefined;

function workScheduleClient() {
  client ??= createClient(WorkScheduleService, rpcTransport());
  return client;
}

function safeInt(value: bigint, label: string): number {
  const n = Number(value);
  if (!Number.isSafeInteger(n)) {
    throw new Error(`${label} ${value} is outside JavaScript's safe integer range`);
  }
  return n;
}

function mapWindow(w: WireWorkingWindow): WorkingWindow {
  return { startMinutes: w.startMinutes, endMinutes: w.endMinutes };
}

function mapDay(d: WireWorkScheduleDay): WorkScheduleDay {
  return {
    weekday: d.weekday,
    expectedMinutes: d.expectedMinutes,
    windows: d.windows.map(mapWindow),
  };
}

export function mapWorkSchedule(schedule: WireWorkSchedule): WorkSchedule {
  return {
    id: safeInt(schedule.id, "work schedule id"),
    timezone: schedule.timezone,
    workweekStart: schedule.workweekStart,
    effectiveFrom: schedule.effectiveFrom,
    ...(schedule.effectiveTo ? { effectiveTo: schedule.effectiveTo } : {}),
    days: schedule.days.map(mapDay),
  };
}

export function mapScheduleException(exc: WireScheduleException): ScheduleException {
  return {
    id: safeInt(exc.id, "schedule exception id"),
    date: exc.date,
    kind: exc.kind,
    expectedMinutes: exc.expectedMinutes,
    windows: exc.windows.map(mapWindow),
  };
}

export function mapExpectedTime(item: WireExpectedTime): ExpectedTime {
  return {
    date: item.date,
    expectedMinutes: item.expectedMinutes,
    windows: item.windows.map(mapWindow),
    source: item.source,
    ...(item.exceptionKind ? { exceptionKind: item.exceptionKind } : {}),
    ...(item.timezone ? { timezone: item.timezone } : {}),
    ...(item.workweekStart ? { workweekStart: item.workweekStart } : {}),
  };
}

export async function listWorkSchedulesRPC() {
  const response = await workScheduleClient().listWorkSchedules({});
  return response.schedules.map(mapWorkSchedule);
}

export async function getWorkScheduleRPC(id: number) {
  const response = await workScheduleClient().getWorkSchedule({ id: BigInt(id) });
  if (!response.schedule) throw new Error("get work schedule response is missing schedule");
  return mapWorkSchedule(response.schedule);
}

export async function replaceActiveWorkScheduleRPC(input: {
  timezone: string;
  workweekStart: string;
  effectiveFrom: string;
  days: WorkScheduleDay[];
}) {
  const response = await workScheduleClient().replaceActiveWorkSchedule({
    timezone: input.timezone,
    workweekStart: input.workweekStart,
    effectiveFrom: input.effectiveFrom,
    days: input.days,
  });
  if (!response.schedule) throw new Error("replace work schedule response is missing schedule");
  return mapWorkSchedule(response.schedule);
}

export async function listScheduleExceptionsRPC() {
  const response = await workScheduleClient().listScheduleExceptions({});
  return response.exceptions.map(mapScheduleException);
}

export async function upsertScheduleExceptionRPC(input: {
  date: string;
  kind: string;
  expectedMinutes: number;
  windows?: WorkingWindow[];
}) {
  const response = await workScheduleClient().upsertScheduleException({
    date: input.date,
    kind: input.kind,
    expectedMinutes: input.expectedMinutes,
    windows: input.windows ?? [],
  });
  if (!response.exception) throw new Error("upsert schedule exception response is missing exception");
  return mapScheduleException(response.exception);
}

export async function deleteScheduleExceptionRPC(date: string) {
  await workScheduleClient().deleteScheduleException({ date });
}

export async function expectedTimeForDateRPC(date: string) {
  const response = await workScheduleClient().expectedTimeForDate({ date });
  if (!response.expectedTime) throw new Error("expected time response is missing expected_time");
  return mapExpectedTime(response.expectedTime);
}

export async function expectedTimeForRangeRPC(startDate: string, endDate: string) {
  const response = await workScheduleClient().expectedTimeForRange({ startDate, endDate });
  return response.days.map(mapExpectedTime);
}
