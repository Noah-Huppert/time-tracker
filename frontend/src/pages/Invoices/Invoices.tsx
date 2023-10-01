import { Paper, Table, TableContainer, TableHead, TableRow, TableCell, CircularProgress, Typography, TableBody, Button, Box } from "@mui/material";
import { useCallback, useContext, useEffect, useState } from "react";
import { api, InvoiceSchemaType } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import { ToastCtx } from "../../components/Toast/Toast";
import WarningIcon from "@mui/icons-material/Warning";
import { DATE_FORMAT, DURATION_FORMAT, nanosecondsToDuration } from "../../lib/time";
import dayjs from "dayjs";
import { ROUTES } from "../../lib/routes";
import { Link, useNavigate } from "react-router-dom";
import { Header } from "../../components/Header/Header";
import { Filters } from "../../components/Filters/Filters";

const BoolFilter = ({
  value,
  setValue,
}: {
  readonly value: boolean | null
  readonly setValue: (value: boolean | null) => void
}) => {
  return (
    <>
      <Button
        onClick={() => value === null ? setValue(false) : setValue(!value)}
      >
        Select boolean:
        {JSON.stringify(value)}
      </Button>
    </>
  )
}

const NumberFilter = ({
  value,
  setValue,
}: {
  readonly value: number | null
  readonly setValue: (value: number | null) => void
}) => {
  return (
    <>
      <Button
        onClick={() => value === null ? setValue(0) : setValue(value+1)}
      >
        Click to increment:
        {JSON.stringify(value)}
      </Button>
    </>
  )
}

export const PageInvoices = () => {
  const toast = useContext(ToastCtx);
  const navigate = useNavigate();

  const [invoices, setInvoices] = useState<InvoiceSchemaType[] | "loading" | "error">("loading");

  const fetchInvoices = useCallback(async() => {
    const res = await api.invoices.list({})
    if (isLeft(res)) {
      console.error(`Failed to list invoices: ${res.left}`);
      setInvoices("error");
      return;
    }

    setInvoices(res.right);
  }, [setInvoices]);

  useEffect(() => {
    fetchInvoices();
  }, [fetchInvoices]);

  const [filters, setFilters] = useState<{
    archived: boolean | null,
    counter: number | null,
  }>({
    archived: null,
    counter: null,
  })

  if (invoices === "loading") {
    return (
      <>
        <CircularProgress size="medium" />
        <Typography>Loading</Typography>
      </>
    );
  }

  if (invoices === "error") {
    return (
      <>
        <WarningIcon />
        <Typography>Failed to load data</Typography>
      </>
    );
  }

  return (
    <>
      <Header />
      <Box
        sx={{
          width: "50rem",
          padding: "2rem",
        }}
      >
        <Box>
          <Typography
            variant="h5"
            sx={{
              marginBottom: "1rem",
            }}
          >
            Invoices
          </Typography>

          <Filters
            filterValues={filters}
            setFilterValues={setFilters}
            filterConditions={{
              archived: {
                name: "Archived",
                start: () => false,
                component: BoolFilter,
              },
              counter: {
                name: "Counter",
                start: () => 0,
                component: NumberFilter,
              }
            }}
          />
        </Box>

        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell variant="head">Start</TableCell>
                <TableCell variant="head">End</TableCell>
                <TableCell variant="head">Duration</TableCell>
                <TableCell variant="head">Amount Due</TableCell>
                <TableCell variant="head">Sent</TableCell>
                <TableCell variant="head">Paid</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {invoices.map((invoice) => (
                <TableRow
                  key={`invoice-${invoice.id}`}
                  onClick={() => navigate(ROUTES.viewInvoice.make({
                    invoiceID: invoice.id,
                  }))}
                  sx={{
                    cursor: "pointer",
                  }}
                >
                  <TableCell>{invoice.start_date.toISOString()}</TableCell>
                  <TableCell>{invoice.end_date.toISOString()}</TableCell>
                  <TableCell>{nanosecondsToDuration(invoice.duration).format(DURATION_FORMAT)}</TableCell>
                  <TableCell>${invoice.amount_due.toFixed(2)}</TableCell>
                  <TableCell>{invoice.sent_to_client === null ? "" : dayjs(invoice.sent_to_client).format(DATE_FORMAT)}</TableCell>
                  <TableCell>{invoice.paid_by_client === null ? "" : dayjs(invoice.paid_by_client).format(DATE_FORMAT)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    </>
  );
};