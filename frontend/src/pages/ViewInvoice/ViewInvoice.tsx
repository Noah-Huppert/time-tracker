import {
  Box,
  CircularProgress,
  TableContainer,
  Typography,
  TableHead,
  Table,
  TableCell,
  TableBody,
  TableRow,
  styled,
  tableCellClasses,
  Button,
  AppBar,
  Container,
  Toolbar,
  Drawer,
  IconButton,
} from "@mui/material";
import { forwardRef, useCallback, useContext, useEffect, useRef, useState } from "react";

import SettingsIcon from "@mui/icons-material/Settings";
import {
  InvoiceSchemaType,
  InvoiceTimeEntrySchemaType,
  UpdateInvoiceOpts,
  api,
} from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import { useNavigate, useParams } from "react-router-dom";
import dayjs from "dayjs";
import dayjsDuration from "dayjs/plugin/duration";
import {
  DATE_FORMAT,
  DATE_TIME_FORMAT,
  DURATION_FORMAT,
  nanosecondsToDuration,
} from "../../lib/time";
import { useReactToPrint } from "react-to-print";
import PrintIcon from "@mui/icons-material/Print";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import CancelIcon from '@mui/icons-material/Cancel';
import { DatePicker } from "@mui/x-date-pickers";
import { ToastCtx } from "../../components/Toast/Toast";

import "./ViewInvoice.scss";

dayjs.extend(dayjsDuration);

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

const TABLE_COL_WIDTHS = [150, 150, 50, 300];

export const PageViewInvoice = () => {

  const navigate = useNavigate();
  const toast = useContext(ToastCtx);
  const { id: invoiceIDStr } = useParams();
  const invoiceID = Number(invoiceIDStr);

  const [invoice, setInvoice] = useState<
    InvoiceSchemaType | "loading" | "error"
  >("loading");
  const [metadataDrawerOpen, setMetadataDrawerOpen] = useState(false);

  const ref = useRef(null);
  const onPrint = useCallback(useReactToPrint({
    content: () => ref.current,
    documentTitle: invoice !== "loading" && invoice !== "error" ? `Invoice-${dayjs(invoice.start_date).format(DATE_FORMAT)}-${dayjs(invoice.end_date).format(DATE_FORMAT)}` : "Invoice",
    copyStyles: true,
  }), [invoice]);

  const fetchInvoice = useCallback(async (invoiceID: number) => {
    setInvoice("loading");

    const res = await api.invoices.list({
      ids: [invoiceID],
    });
    if (isLeft(res)) {
      setInvoice("error");
      return;
    }

    setInvoice(res.right[0]);
  }, []);

  const onUpdateInvoice = useCallback(async ({
    sentToClient,
    paidByClient,
  }: UpdateInvoiceOpts) => {
    const res = await api.invoices.update({
      id: invoiceID,
      sentToClient,
      paidByClient,
    });
    if (isLeft(res)) {
      toast({
        kind: "error",
        message: "Failed to update invoice",
      });
      return;
    }

    toast({
      kind: "success",
      message: "Updated invoice"
    })
    await fetchInvoice(invoiceID);
  }, [fetchInvoice])

  useEffect(() => {
    if (invoiceID === undefined) {
      setInvoice("error");
      return;
    }

    fetchInvoice(invoiceID);
  }, [invoiceID, fetchInvoice]);

  return (
    <Box
    sx={{
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
    }}
    >
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
              startIcon={<ArrowBackIcon />}
              onClick={() => navigate(-1)}
              variant="contained"
              color="secondary"
            >
              Back
            </Button>

            <Typography variant="h5">Invoice</Typography>

            <Box>
              <Button
                onClick={() => setMetadataDrawerOpen(true)}
                startIcon={<SettingsIcon />}
                sx={{
                  marginRight: "1rem",
                }}
                variant="contained"
                color="secondary"
              >
                Metadata
              </Button>

              <Button
                startIcon={<PrintIcon />}
                variant="contained"
                onClick={onPrint}
                color="secondary"
              >
                Print
              </Button>
            </Box>
          </Toolbar>
        </Container>
      </AppBar>

      {invoice === "loading" || invoice === "error" ? null : (
        <MetadataDrawer
          open={metadataDrawerOpen}
          setOpen={setMetadataDrawerOpen}
          invoice={invoice}
          onUpdateInvoice={onUpdateInvoice}
        />
      )}

      <Invoice
        ref={ref}
        invoice={invoice}
      />
    </Box>
  );
};

export const MetadataDrawer = ({
  open,
  setOpen,
  invoice,
  onUpdateInvoice,
}: {
  readonly open: boolean
  readonly setOpen: (open: boolean) => void
  readonly invoice: InvoiceSchemaType
  readonly onUpdateInvoice: (opts: UpdateInvoiceOpts) => Promise<void>
}) => {
  const [draftSentToClient, setDraftSentToClient] = useState<Date | null>(invoice.sent_to_client)
  const [draftPaidByClient, setDraftPaidByClient] = useState<Date | null>(invoice.paid_by_client);

  return (
    <Drawer
      open={open}
      onClose={() => setOpen(false)}
      anchor="right"
    >
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          padding: "1rem",
        }}
      >
        <Typography variant="h5">Metadata</Typography>

        <Table>
          <TableBody>
            <TableRow>
              <TableCell>
                <b>Sent To Client</b>
              </TableCell>

              <TableCell
                sx={{
                  display: "flex",
                  flexDirection: "row",
                  justifyContent: "center",
                }}
              >
                <DatePicker<dayjs.Dayjs>
                  value={draftSentToClient !== null ? dayjs(draftSentToClient) : null}
                  onChange={(d) => {
                    console.log("set", d, d?.toDate());
                    setDraftSentToClient(d?.toDate() || null)
                  }}
                />
                <IconButton
                  onClick={() => setDraftSentToClient(null)}
                >
                  <CancelIcon />
                </IconButton>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>
                <b>Paid By Client</b>
              </TableCell>

              <TableCell
                sx={{
                  display: "flex",
                  flexDirection: "row",
                  justifyContent: "center",
                }}
              >
                <DatePicker<dayjs.Dayjs>
                  value={draftPaidByClient !== null ? dayjs(draftPaidByClient) : null}
                  onChange={(d) => setDraftPaidByClient(d?.toDate() || null)}
                />
                <IconButton
                  onClick={() => setDraftPaidByClient(null)}
                >
                  <CancelIcon />
                </IconButton>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <Box
          sx={{
            marginTop: "1rem",
            display: "flex",
            flexDirection: "row",
            alignItems: "flex-end",
          }}
        >
          <Button
            onClick={() => onUpdateInvoice({
              sentToClient: draftSentToClient,
              paidByClient: draftPaidByClient,
            })}
            sx={{
              display: "flex",
              marginLeft: "auto",
            }}
          >
            Save
          </Button>
        </Box>
      </Box>
    </Drawer>
  )
}

export const Invoice = forwardRef<
  HTMLDivElement,
  {
    readonly invoice: InvoiceSchemaType | "loading" | "error"
  }
>(({ invoice }, ref) => {

  if (invoice === "loading") {
    return (
      <>
        <CircularProgress size="medium" />
        <Typography>Loading</Typography>
      </>
    );
  }

  if (invoice === "error") {
    return (
      <>
        <WarningIcon />
        <Typography>Failed to load data</Typography>
      </>
    );
  }

  return (
      <div
        ref={ref}
        style={{
          width: "200mm",
          padding: "20mm",
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
          <TimeEntriesTable timeEntries={invoice.invoice_time_entries} />
        </Box>
      </div>
  );
});

const InvoiceHeader = ({
  recipient,
  sender,
  periodStart,
  periodEnd,
}: {
  readonly recipient: string;
  readonly sender: string;
  readonly periodStart: Date;
  readonly periodEnd: Date;
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
        <Typography>Billed to:</Typography>

        <Box>
          {splitRecipient.map((line, i) => (
            <Typography key={`invoice-recipient-line-${i}`}>{line}</Typography>
          ))}
        </Box>
      </Box>

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Typography>Period of performance:</Typography>

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
            {dayjs(periodStart).format(DATE_TIME_FORMAT)}
          </Typography>
          <Typography>-</Typography>
          <Typography
            sx={{
              paddingLeft: "0.5rem",
            }}
          >
            {dayjs(periodEnd).format(DATE_TIME_FORMAT)}
          </Typography>
        </Box>
      </Box>

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Typography>From:</Typography>

        <Box>
          {splitSender.map((line, i) => (
            <Typography key={`invoice-sender-line-${i}`}>{line}</Typography>
          ))}
        </Box>
      </Box>
    </Box>
  );
};

const SummaryTable = ({
  periodStart,
  periodEnd,
  totalDuration,
  amountDue,
}: {
  readonly periodStart: Date;
  readonly periodEnd: Date;
  readonly totalDuration: number;
  readonly amountDue: number;
}) => {
  return (
    <>
      Owed:
      <div  className="view-invoice-table">
        <table
          style={{
            width: "100%",
            textAlign: "left",
          }}
        >
          <thead>
            <tr>
              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[0]}px`,
                }}
              >
                Period Start
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[1]}px`,
                }}
              >
                Period End
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[2]}px`,
                }}
              >
                Duration
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[3]}px`,
                }}
              >
                Amount Due
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                {dayjs(periodStart).format(DATE_TIME_FORMAT)}
              </td>

              <td>
                {dayjs(periodEnd).format(DATE_TIME_FORMAT)}
              </td>

              <td>
                {nanosecondsToDuration(totalDuration).format(DURATION_FORMAT)}
              </td>

              <td>${amountDue.toFixed(2)}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </>
  );
};

const TimeEntriesTable = ({
  timeEntries,
}: {
  readonly timeEntries: InvoiceTimeEntrySchemaType[];
}) => {
  return (
    <>
      Timesheet:
      <div className="view-invoice-table">
        <table>
          <thead>
            <tr>
              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[0]}px`,
                }}
              >
                Time Started
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[1]}px`,
                }}
              >
                Time Ended
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[2]}px`,
                }}
              >
                Duration
              </th>

              <th
                style={{
                  width: `${TABLE_COL_WIDTHS[3]}px`,
                }}
              >
                Comment
              </th>
            </tr>
          </thead>
          <tbody>
            {timeEntries.map((timeEntry) => (
              <tr key={`invoice-timesheet-${timeEntry.id}`}>
                <td>
                  {dayjs(timeEntry.time_entry.start_time).format(DATE_TIME_FORMAT)}
                </td>

                <td>
                  {dayjs(timeEntry.time_entry.end_time).format(DATE_TIME_FORMAT)}
                </td>

                <td>
                  {nanosecondsToDuration(timeEntry.time_entry.duration).format(
                    DURATION_FORMAT,
                  )}
                </td>

                <td>
                  {timeEntry.time_entry.comment.length > 0
                    ? timeEntry.time_entry.comment
                    : "\u00A0"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};
