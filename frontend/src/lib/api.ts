import { z } from "zod";
import { Either, left, right } from "fp-ts/lib/Either";

const timeEntrySchema = z.object({
  id: z.number(),
  start_time: z.coerce.date(),
  end_time: z.coerce.date(),
  comment: z.string(),
  duration: z.number(),
});
export type TimeEntrySchemaType = z.infer<typeof timeEntrySchema>;

const listTimeEntriesSchema = z.object({
  time_entries: z.array(timeEntrySchema),
  total_duration: z.number(),
});
export type ListTimeEntriesSchemaType = z.infer<typeof listTimeEntriesSchema>;

const timeEntriesUploadCSVListItemSchema = z.object({
  id: z.number(),
  start_time: z.string(),
  end_time: z.string(),
  comment: z.string(),
});
const timeEntriesUploadCSVSchema = z.object({
  existing_time_entries: z.array(timeEntriesUploadCSVListItemSchema),
  new_time_entries: z.array(timeEntriesUploadCSVListItemSchema),
});

const invoiceSettingsSchema = z.object({
  id: z.number(),
  hourly_rate: z.number(),
  recipient: z.string(),
  sender: z.string(),
});
export type InvoiceSettingsSchemaType = z.infer<typeof invoiceSettingsSchema>;

const invoiceTimeEntrySchema = z.object({
  id: z.number(),
  invoice_id: z.number(),
  time_entry_id: z.number(),
  time_entry: timeEntrySchema,
});
export type InvoiceTimeEntrySchemaType = z.infer<typeof invoiceTimeEntrySchema>;

const invoiceSchema = z.object({
  id: z.number(),
  invoice_settings_id: z.number(),
  start_date: z.coerce.date(),
  end_date: z.coerce.date(),
  duration: z.number(),
  amount_due: z.number(),
  sent_to_client: z.nullable(z.coerce.date()),
  paid_by_client: z.nullable(z.coerce.date()),
  invoice_settings: invoiceSettingsSchema,
  invoice_time_entries: z.array(invoiceTimeEntrySchema),
});
export type InvoiceSchemaType = z.infer<typeof invoiceSchema>;

export type CSVFile = {
  readonly name: string;
  readonly content: string;
};

export type UpdateInvoiceOpts = {
  readonly sentToClient?: Date
  readonly paidByClient?: Date
}

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
  readonly queryParams?: { [key: string]: string | boolean | undefined | null };
  readonly body?: object;
}): Promise<Either<Error, T>> {
  if (path[0] === "/") {
    return left(
      new Error(
        "path argument cannot start with leading slash, as this will clobber the base URL",
      ),
    );
  }

  // Make request
  const setQueryParams: { [key: string]: string } = {};
  if (queryParams !== undefined) {
    Object.entries(queryParams).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        if (typeof value === "boolean") {
          setQueryParams[key] = JSON.stringify(value);
        } else {
          setQueryParams[key] = value;
        }
      }
    });
  }

  const headers: { [key: string]: string } = {};
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const url =
    new URL(path, BASE_URL).href +
    "?" +
    new URLSearchParams(setQueryParams || {});
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
      readonly startDate: Date | null;
      readonly endDate: Date | null;
    }) =>
      makeReq({
        path: "time-entries/",
        method: "GET",
        shape: listTimeEntriesSchema,
        queryParams: {
          start_date: startDate?.toISOString(),
          end_date: endDate?.toISOString(),
        },
      }),

    uploadCSV: ({ csvFiles }: { readonly csvFiles: CSVFile[] }) =>
      makeReq({
        path: "time-entries/upload-csv/",
        method: "POST",
        shape: timeEntriesUploadCSVSchema,
        body: {
          csv_files: csvFiles,
        },
      }),
  },

  invoiceSettings: {
    get: () =>
      makeReq({
        path: "invoice-settings/",
        method: "GET",
        shape: z.nullable(invoiceSettingsSchema),
      }),

    set: ({
      hourlyRate,
      recipient,
      sender,
    }: {
      readonly hourlyRate: number;
      readonly recipient: string;
      readonly sender: string;
    }) =>
      makeReq({
        path: "invoice-settings/",
        method: "PUT",
        shape: z.nullable(invoiceSettingsSchema),
        body: {
          hourly_rate: hourlyRate,
          recipient,
          sender,
        },
      }),
  },

  invoices: {
    list: ({
      ids,
      archived,
    }: {
      readonly ids?: number[];
      readonly archived?: boolean;
    }) =>
      makeReq({
        path: "invoices/",
        method: "GET",
        shape: z.array(invoiceSchema),
        queryParams: {
          ids: ids?.join(",") || undefined,
          archived,
        },
      }),

    create: ({
      invoiceSettingsID,
      startDate,
      endDate,
    }: {
      readonly invoiceSettingsID: number;
      readonly startDate: Date;
      readonly endDate: Date;
    }) =>
      makeReq({
        path: "invoices/",
        method: "POST",
        shape: invoiceSchema,
        body: {
          invoice_settings_id: invoiceSettingsID,
          start_date: startDate.toISOString(),
          end_date: endDate.toISOString(),
        },
      }),

    update: ({
      id,
      sentToClient,
      paidByClient,
    }: {
      readonly id: number; 
    } & UpdateInvoiceOpts) => makeReq({
      path: `invoices/${id}/`,
      method: "PATCH",
      shape: invoiceSchema,
      body: {
        sent_to_client: sentToClient,
        paid_by_client: paidByClient,
      },
    }),
  },
};
