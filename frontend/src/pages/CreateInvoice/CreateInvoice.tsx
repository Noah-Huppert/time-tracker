import { Box, CircularProgress, TableContainer, Typography, TableHead, Table, TableCell, TableBody, TableRow, styled, tableCellClasses, Button, Link, AppBar, Container, Toolbar } from "@mui/material";
import { MutableRefObject, ReactInstance, forwardRef, useCallback, useEffect, useRef, useState } from "react";
import { InvoiceSettingsSchemaType, ListTimeEntriesSchemaType, TimeEntrySchemaType, api } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import { Link as RouterLink, useSearchParams } from "react-router-dom";
import { QUERY_PARAMS, ROUTES } from "../../lib/routes";
import dayjs from "dayjs";
import dayjsDuration, { Duration } from "dayjs/plugin/duration";
import { nanosecondsToDuration } from "../../lib/time";
import { useReactToPrint } from "react-to-print";
import PrintIcon from '@mui/icons-material/Print';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';

dayjs.extend(dayjsDuration);

const DATE_FORMAT = "YY-MM-DD HH:mm:ss";
const DURATION_FORMAT = "HH:mm:ss";

const BorderedTableCell = styled(TableCell)(({ theme }) => ({
  [`&.${tableCellClasses.root}`]: {
    border: `1px solid ${theme.palette.common.black}`,
  },
}));

const HeaderTableCell = styled(BorderedTableCell)(({ theme }) => ({
  [`&.${tableCellClasses.head}`]: {
    backgroundColor: theme.palette.grey[700],
    color: theme.palette.common.white,
  },
}));

const TABLE_COL_WIDTHS = [
  150,
  150,
  50,
  300,
]

export const PageCreateInvoice = () => {
  const ref = useRef(null);
  const onPrint = useReactToPrint({
    content: () => ref.current,
  })

  return (
    <>
      <AppBar component="nav" position="static">
        <Container>
          <Toolbar
            sx={{
              display: "flex",
              flexDirection: "row",
              justifyContent: "space-between",
            }}
          >
            <Button
              component={RouterLink}
              startIcon={<ArrowBackIcon />}
              to={ROUTES.time_entries.make()}
              variant="contained"
              color="secondary"
            >
              Back
            </Button>

            <Typography variant="h5">
              Invoice
            </Typography>

            <Button
              startIcon={<PrintIcon />}
              variant="contained"
              onClick={onPrint}
              color="secondary"
            >
              Print
            </Button>
          </Toolbar>
        </Container>
      </AppBar>

      <Invoice ref={ref} />
    </>
  )
}

export const Invoice = forwardRef((props, ref) => {
  const [invoiceSettings, setInvoiceSettings] = useState<InvoiceSettingsSchemaType | "loading" | "error">("loading");
  const [timeEntries, setTimeEntries] = useState<ListTimeEntriesSchemaType | "loading" | "error">("loading");

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
      startDate: startDate,
      endDate: endDate,
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
    const startDate = startDateStr !== null ? new Date(startDateStr) : null;
    const endDate = endDateStr !== null ? new Date(endDateStr) : null;

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
  let popStart = startDateStr !== null ? new Date(startDateStr) : null;
  let popEnd = endDateStr !== null ? new Date(endDateStr) : null;

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

  const totalDuration = nanosecondsToDuration(timeEntries.total_duration);
  const amountDue = invoiceSettings.hourly_rate * totalDuration.asHours();

  return (
    <Box
      ref={ref}
      sx={{
        display: "flex",
        flexDirection: "row",
        justifyContent: "center",
      }}
    >
      <Box
        sx={{
          width: "50rem",
          padding: "2rem",
        }}
      >
        <InvoiceHeader
          recipient={invoiceSettings.recipient}
          sender={invoiceSettings.sender}
          periodStart={popStart}
          periodEnd={popEnd}
        />

        <Box
          sx={{
            marginTop: "2rem",
          }}
        >
          <SummaryTable
            periodStart={popStart}
            periodEnd={popEnd}
            totalDuration={totalDuration}
            amountDue={amountDue}
          />
        </Box>

        <Box
          sx={{
            marginTop: "2rem",
          }}
        >
          <TimeEntriesTable
            timeEntries={timeEntries.time_entries}
          />
        </Box>
      </Box>
    </Box>
  )
});

const InvoiceHeader = ({
  recipient,
  sender,
  periodStart,
  periodEnd,
}: {
  readonly recipient: string
  readonly sender: string
  readonly periodStart: Date
  readonly periodEnd: Date
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
            {dayjs(periodStart).format(DATE_FORMAT)}
          </Typography>
          <Typography>
            -
          </Typography>
          <Typography
            sx={{
              paddingLeft: "0.5rem",
            }}
          >
            {dayjs(periodEnd).format(DATE_FORMAT)}
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

const SummaryTable = ({
  periodStart,
  periodEnd,
  totalDuration,
  amountDue,
}: {
  readonly periodStart: Date
  readonly periodEnd: Date
  readonly totalDuration: Duration
  readonly amountDue: number
}) => {
  return (
    <>
      Owed:

      <TableContainer component={Box}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <HeaderTableCell width={TABLE_COL_WIDTHS[0]}>Period Start</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[1]}>Period End</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[2]}>Duration</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[3]}>Amount Due</HeaderTableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <BorderedTableCell>
                {dayjs(periodStart).format(DATE_FORMAT)}
              </BorderedTableCell>

              <BorderedTableCell>
                {dayjs(periodEnd).format(DATE_FORMAT)}
              </BorderedTableCell>

              <BorderedTableCell>
                {totalDuration.format(DURATION_FORMAT)}
              </BorderedTableCell>

              <BorderedTableCell>
                ${amountDue.toFixed(2)}
              </BorderedTableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </>
  )
}

const TimeEntriesTable = ({
  timeEntries
}: {
  readonly timeEntries: TimeEntrySchemaType[]
}) => {
  return (
    <>
      Timesheet:

      <TableContainer component={Box}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <HeaderTableCell width={TABLE_COL_WIDTHS[0]}>Time Started</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[1]}>Time Ended</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[2]}>Duration</HeaderTableCell>

              <HeaderTableCell width={TABLE_COL_WIDTHS[3]}>Comment</HeaderTableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {timeEntries.map((timeEntry) => (
              <TableRow key={`invoice-timesheet-${timeEntry.hash}`}>
                <BorderedTableCell>
                  {dayjs(timeEntry.start_time).format(DATE_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {dayjs(timeEntry.end_time).format(DATE_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {nanosecondsToDuration(timeEntry.duration).format(DURATION_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {timeEntry.comment.length > 0 ? timeEntry.comment : '\u00A0'}
                </BorderedTableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  )
}