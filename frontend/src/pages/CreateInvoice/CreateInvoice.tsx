import { Box, CircularProgress, Typography } from "@mui/material";
import { useCallback, useEffect, useState } from "react";
import { InvoiceSettingsSchemaType, ListTimeEntriesSchemaType, api } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import { useSearchParams } from "react-router-dom";
import { QUERY_PARAMS } from "../../lib/routes";
import dayjs from "dayjs";

const DATE_FORMAT = "YY-MM-DD HH:mm:ss";

export const PageCreateInvoice = () => {
  const [invoiceSettings, setInvoiceSettings] = useState<InvoiceSettingsSchemaType | "loading" | "error">("loading");
  const [timeEntries, setTimeEntries] = useState<ListTimeEntriesSchemaType | "loading" | "error">("loading");

  const [startDate, setStartDate] = useState<Date | null>(null);
  const [endDate, setEndDate] = useState<Date | null>(null);

  const [queryParams] = useSearchParams();
  const startDateStr = queryParams.get(QUERY_PARAMS.invoice.startDate);
  const endDateStr = queryParams.get(QUERY_PARAMS.invoice.endDate);

  const fetchInvoiceSettings = useCallback(async () => {
    const res = await api.invoiceSettings.get();
    if (isLeft(res)) {
      setInvoiceSettings("error");
      return;
    }

    setInvoiceSettings(res.right);
  }, []);

  const fetchTimeEntries = useCallback(async ({
    startDate,
    endDate,
  }: {
    readonly startDate: Date | null
    readonly endDate: Date | null
  }) => {
    const res = await api.timeEntries.list({
      startTime: startDate,
      endTime: endDate,
    });
    if (isLeft(res)) {
      setTimeEntries("error");
      return;
    }

    setTimeEntries(res.right);
  }, [])

  useEffect(() => {
    fetchInvoiceSettings();
  }, [])

  useEffect(() => {
    setStartDate(startDateStr !== null ? new Date(startDateStr) : null)
    setEndDate(endDateStr !== null ? new Date(endDateStr) : null);

    fetchTimeEntries({
      startDate,
      endDate,
    });
  }, [startDateStr, endDateStr])

  if (invoiceSettings === "loading" || timeEntries === "loading") {
    return (
      <>
        <CircularProgress size="medium" />
        <Typography>
          Loading
        </Typography>
      </>
    );
  }

  if (invoiceSettings === "error" || timeEntries === "error") {
    return (
      <>
        <WarningIcon />
        <Typography>
          Failed to load data
        </Typography>
      </>
    );
  }

  if (timeEntries.time_entries.length === 0) {
    return (
      <>
        <Typography>
          <WarningIcon />
          Cannot make invoice for no entries
        </Typography>
      </>
    )
  }
  
  // Calculate details about invoice
  let popStart = startDate;
  let popEnd = endDate;

  if (popStart === null) {
    popStart = new Date(timeEntries.time_entries[0].start_time)
    popStart.setHours(0);
    popStart.setMinutes(0);
    popStart.setSeconds(0);
  }

  if (popEnd === null) {
    popEnd = new Date(timeEntries.time_entries[timeEntries.time_entries.length - 1].end_time)
    popEnd.setHours(0);
    popEnd.setMinutes(0);
    popEnd.setSeconds(0);
  }

  return (
    <Box
      sx={{
        width: "100%",
      }}
    >
      <InvoiceHeader
        recipient={invoiceSettings.recipient}
        sender={invoiceSettings.sender}
        startDate={popStart}
        endDate={popEnd}
      />
    </Box>
  )
};

const InvoiceHeader = ({
  recipient,
  sender,
  startDate,
  endDate,
}: {
  readonly recipient: string
  readonly sender: string
  readonly startDate: Date
  readonly endDate: Date
}) => {
  const splitRecipient = recipient.split("\n");
  const splitSender = sender.split("\n");

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "row",
        justifyContent: "space-between",
      }}
    >
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Typography>
          Billed to:
        </Typography>

        <Box>
          {splitRecipient.map((line, i) => (
            <Typography key={`invoice-recipient-line-${i}`}>
              {line}
            </Typography>
          ))}
        </Box>
      </Box>

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Typography>
          Period of performance:
        </Typography>
        
        <Box
          sx={{
            display: "flex",
            flexDirection: "row",
          }}
        >
          <Typography
            sx={{
              paddingRight: "0.5rem",
            }}
          >
            {dayjs(startDate).format(DATE_FORMAT)}
          </Typography>
          <Typography>
            -
          </Typography>
          <Typography
            sx={{
              paddingLeft: "0.5rem",
            }}
          >
            {dayjs(endDate).format(DATE_FORMAT)}
          </Typography>
        </Box>
      </Box>

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Typography>
          From:
        </Typography>
        
        <Box>
          {splitSender.map((line, i) => (
            <Typography key={`invoice-sender-line-${i}`}>
              {line}
            </Typography>
          ))}
        </Box>
      </Box>
    </Box>
  )
}