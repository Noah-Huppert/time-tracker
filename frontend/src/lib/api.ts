import { z } from "zod";
import { Either, left, right } from "fp-ts/lib/Either";

const timeEntrySchema = z.object({
  start_time: z.string(),
  end_time: z.string(),
  comment: z.string(),
  hash: z.string(),
});
export type TimeEntrySchemaType = z.infer<typeof timeEntrySchema>;

const BASE_URL = "http://localhost:4000/api/v0/";

async function makeReq<T>({
  path,
  method,
  shape,
  queryParams,
  body,
}: {
  readonly path: string;
  readonly method: string;
  readonly shape: z.Schema<T>;
  readonly queryParams?: {[key: string]: string | undefined | null};
  readonly body?: string;
}): Promise<Either<Error, T>> {
  if (path[0] === "/") {
    return left(new Error("path argument cannot start with leading slash, as this will clobber the base URL"));
  }

  // Make request
  const setQueryParams: {[key: string]: string} = {}
  if (queryParams !== undefined) {
    Object.entries(queryParams).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        setQueryParams[key] = value;
      }
    });
  }
  const url = new URL(path, BASE_URL).href + "?" + new URLSearchParams(setQueryParams || {})
  const res = await fetch(url, {
    method,
    body,
  });

  // Decode response
  const respBody = await res.json();
  const decodeRes = shape.safeParse(respBody);
  if (decodeRes.success === false) {
    return left(
      new Error(`failed to parse body using schema: ${decodeRes.error}`),
    );
  }

  return right(decodeRes.data);
}

export const api = {
  timeEntries: {
    list: ({
      startTime,
      endTime,
    }: {
      readonly startTime?: Date
      readonly endTime?: Date
    }) =>
      makeReq({
        path: "time-entries",
        method: "GET",
        shape: z.object({
          time_entries: z.array(timeEntrySchema),
        }),
        queryParams: {
          "start_time": startTime?.toISOString(),
          "end_time": endTime?.toISOString(),
        }
      }),
  },
};
