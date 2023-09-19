import { z } from "zod";
import { Either, isLeft, left, right } from "fp-ts/lib/Either";

const timeEntrySchema = z.object({
  start_time: z.string(),
  end_time: z.string(),
  comment: z.string(),
  hash: z.string(),
})
export type TimeEntrySchemaType = z.infer<typeof timeEntrySchema>;

const BASE_URL = "http://localhost:4000/api/v0/"

async function makeReq<T>({
    path,
    method,
    shape,
    body,
}: {
    readonly path: string
    readonly method: string
    readonly shape: z.Schema<T>,
    readonly body?: string,
}): Promise<Either<Error, T>> {
    // Make request
    const res = await fetch(new URL(path, BASE_URL).href,
    {
        method,
        body,
    })

    // Decode response
    const respBody = await res.json()
    const decodeRes = shape.safeParse(respBody)
    if (decodeRes.success === false) {
      return left(new Error(`failed to parse body using schema: ${decodeRes.error}`));
    }

    return right(decodeRes.data);
}

export const api = {
    timeEntries: {
        list: () => makeReq({
            path: "time-entries",
            method: "GET",
            shape: z.object({
                time_entries: z.array(timeEntrySchema),
            }),
        }),
    }
}