import dayjs from "dayjs";
import dayjsDuration, { Duration } from "dayjs/plugin/duration";

dayjs.extend(dayjsDuration);

const MILLISECONDS_PER_NANOSECOND = 1e6;

export function nanosecondsToDuration(nanoseconds: number): Duration {
  return dayjs.duration(
    nanoseconds / MILLISECONDS_PER_NANOSECOND,
    "milliseconds",
  );
}
