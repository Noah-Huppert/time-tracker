import {
  Paper,
  Table,
  TableContainer,
  TableHead,
  TableRow,
  TableCell,
  CircularProgress,
  Typography,
  TableBody,
  Box,
} from "@mui/material";
import { useCallback, useEffect, useState } from "react";
import { api, InvoiceSchemaType } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import {
  DATE_FORMAT,
  DATE_TIME_FORMAT,
  DURATION_FORMAT,
  nanosecondsToDuration,
} from "../../lib/time";
import dayjs from "dayjs";
import { ROUTES } from "../../lib/routes";
import { useNavigate } from "react-router-dom";
import { Header } from "../../components/Header/Header";
import { Filters } from "../../components/Filters/Filters";
import { BooleanFilter } from "../../components/Filters/BooleanFilter";

type InvoiceFilters = {
  archived: boolean | null;
};

export const PageInvoices = () => {
  const navigate = useNavigate();

  const [filters, setFilters] = useState<InvoiceFilters>({
    archived: false,
  });
  const [invoices, setInvoices] = useState<
    InvoiceSchemaType[] | "loading" | "error"
  >("loading");

  const fetchInvoices = useCallback(async (filters: InvoiceFilters) => {
    const res = await api.invoices.list({
      archived: filters.archived || undefined
    });
    if (isLeft(res)) {
      console.error(`Failed to list invoices: ${res.left}`);
      setInvoices("error");
      return;
    }

    setInvoices(res.right);
  }, [setInvoices]);

  useEffect(() => {
    fetchInvoices({
      archived: filters.archived
    });
  }, [fetchInvoices, filters.archived]);

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

          <Box
            sx={{
              marginBottom: "1rem",
            }}
          >
            <Filters
              filterValues={filters}
              setFilterValues={setFilters}
              filterConditions={{
                archived: {
                  name: "Archived",
                  start: () => false,
                  display: (value) => JSON.stringify(value),
                  component: BooleanFilter,
                },
              }}
            />
          </Box>
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
                  onClick={() =>
                    navigate(
                      ROUTES.viewInvoice.make({
                        invoiceID: invoice.id,
                      }),
                    )
                  }
                  sx={{
                    cursor: "pointer",
                  }}
                >
                  <TableCell>{dayjs(invoice.start_date).format(DATE_FORMAT)}</TableCell>
                  <TableCell>{dayjs(invoice.end_date).format(DATE_FORMAT)}</TableCell>
                  <TableCell>
                    {nanosecondsToDuration(invoice.duration).format(
                      DURATION_FORMAT,
                    )}
                  </TableCell>
                  <TableCell>${invoice.amount_due.toFixed(2)}</TableCell>
                  <TableCell>
                    {invoice.sent_to_client === null
                      ? ""
                      : dayjs(invoice.sent_to_client).format(DATE_TIME_FORMAT)}
                  </TableCell>
                  <TableCell>
                    {invoice.paid_by_client === null
                      ? ""
                      : dayjs(invoice.paid_by_client).format(DATE_TIME_FORMAT)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    </>
  );
};
