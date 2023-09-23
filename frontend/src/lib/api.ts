import { z } from "zod";
import { Either, left, right } from "fp-ts/lib/Either";

const timeEntrySchema = z.object({
  start_time: z.string(),
  end_time: z.string(),
  comment: z.string(),
  hash: z.string(),
  duration: z.number(),
});
export type TimeEntrySchemaType = z.infer<typeof timeEntrySchema>;

const listTimeEntriesSchema = z.object({
  time_entries: z.array(timeEntrySchema),
  total_duration: z.number(),
});
export type ListTimeEntriesSchemaType = z.infer<typeof listTimeEntriesSchema>;

const invoiceSettingsSchema = z.object({
  hourly_rate: z.number(),
  recipient: z.string(),
  sender: z.string(),
});
export type InvoiceSettingsSchemaType = z.infer<typeof invoiceSettingsSchema>;

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
  readonly body?: object;
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

  const headers: {[key: string]: string} = {};
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const url = new URL(path, BASE_URL).href + "?" + new URLSearchParams(setQueryParams || {})
  const res = await fetch(url, {
    method,
    body: JSON.stringify(body),
    headers,
  });

  // Check response
  if (res.status >= 299) {
    return left(new Error(`request failed: ${res.status} - ${res.statusText}`));
  }

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
      startDate,
      endDate,
    }: {
      readonly startDate: Date | null
      readonly endDate: Date | null
    }) =>
      makeReq({
        path: "time-entries",
        method: "GET",
        shape: listTimeEntriesSchema,
        queryParams: {
          "start_date": startDate?.toISOString(),
          "end_date": endDate?.toISOString(),
        }
      }),
  },

  invoiceSettings: {
    get: () => makeReq({
      path: "invoice-settings",
      method: "GET",
      shape: invoiceSettingsSchema,
    }),

    set: ({
      hourlyRate,
      recipient,
      sender,
    }: {
      readonly hourlyRate: number
      readonly recipient: string
      readonly sender: string
    }) => makeReq({
      path: "invoice-settings",
      method: "PUT",
      shape: invoiceSettingsSchema,
      body: {
        hourly_rate: hourlyRate,
        recipient,
        sender,
      }
    })
  }
};
