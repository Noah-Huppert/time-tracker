import { Box, CircularProgress, TableContainer, Typography, TableHead, Table, TableCell, TableBody, TableRow, styled, tableCellClasses, Button, AppBar, Container, Toolbar } from "@mui/material";
import { forwardRef, useCallback, useEffect, useRef, useState } from "react";
import { InvoiceSchemaType, InvoiceTimeEntrySchemaType, api } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import { Link as RouterLink, useParams } from "react-router-dom";
import { ROUTES } from "../../lib/routes";
import dayjs from "dayjs";
import dayjsDuration from "dayjs/plugin/duration";
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

export const PageViewInvoice = () => {
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
              to={ROUTES.timeEntries.make()}
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
  const [invoice, setInvoice] = useState<InvoiceSchemaType | "loading" | "error">("loading");
  const { id: invoiceID } = useParams();

  const fetchInvoice = useCallback(async (invoiceID: number) => {
    const res = await api.invoices.list({
      ids: [
        invoiceID,
      ]
    });
    if (isLeft(res)) {
      setInvoice("error")
      return;
    }

    setInvoice(res.right[0]);
  }, []);

  useEffect(() => {
    if (invoiceID === undefined) {
      setInvoice("error");
      return;
    }

    fetchInvoice(Number(invoiceID));
  }, [invoiceID, fetchInvoice])

  if (invoice === "loading") {
    return (
      <>
        <CircularProgress size="medium" />
        <Typography>
          Loading
        </Typography>
      </>
    );
  }

  if (invoice === "error") {
    return (
      <>
        <WarningIcon />
        <Typography>
          Failed to load data
        </Typography>
      </>
    );
  }

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
          recipient={invoice.invoice_settings.recipient}
          sender={invoice.invoice_settings.sender}
          periodStart={invoice.start_date}
          periodEnd={invoice.end_date}
        />

        <Box
          sx={{
            marginTop: "2rem",
          }}
        >
          <SummaryTable
            periodStart={invoice.start_date}
            periodEnd={invoice.end_date}
            totalDuration={invoice.duration}
            amountDue={invoice.amount_due}
          />
        </Box>

        <Box
          sx={{
            marginTop: "2rem",
          }}
        >
          <TimeEntriesTable
            timeEntries={invoice.invoice_time_entries}
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
  readonly totalDuration: number
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
                {nanosecondsToDuration(totalDuration).format(DURATION_FORMAT)}
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
  readonly timeEntries: InvoiceTimeEntrySchemaType[]
}) => {
  console.log(timeEntries[0].time_entry.duration)
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
              <TableRow key={`invoice-timesheet-${timeEntry.id}`}>
                <BorderedTableCell>
                  {dayjs(timeEntry.time_entry.start_time).format(DATE_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {dayjs(timeEntry.time_entry.end_time).format(DATE_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {nanosecondsToDuration(timeEntry.time_entry.duration).format(DURATION_FORMAT)}
                </BorderedTableCell>

                <BorderedTableCell>
                  {timeEntry.time_entry.comment.length > 0 ? timeEntry.time_entry.comment : '\u00A0'}
                </BorderedTableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  )
}