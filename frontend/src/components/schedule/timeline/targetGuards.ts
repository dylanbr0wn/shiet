export function targetBlocksTimelineHover(target: EventTarget | null) {
  return (
    target instanceof Element &&
    target.closest("[data-scheduler-item]") !== null
  );
}

export function targetIgnoresTimelineCreate(target: EventTarget | null) {
  return (
    target instanceof Element &&
    target.closest("[data-scheduler-ignore-create]") !== null
  );
}

export function targetBlocksBackgroundMenu(target: EventTarget | null) {
  return (
    target instanceof Element &&
    target.closest("[data-scheduler-item], [data-scheduler-ignore-create]") !==
      null
  );
}
