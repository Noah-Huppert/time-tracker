import { useCallback, useEffect, useState } from "react";
import { api, InvoiceSettingsSchemaType, ListTimeEntriesSchemaType, TimeEntrySchemaType } from "../../lib/api";
import {
  Box,
  Button,
  CircularProgress,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import WarningIcon from "@mui/icons-material/Warning";
import { isLeft } from "fp-ts/lib/Either";
import dayjsDuration from "dayjs/plugin/duration";
import dayjs from "dayjs";
import { DateFilter } from "../../components/DateFilter/DateFilter";
import { useNavigate } from "react-router-dom";
import { ROUTES } from "../../lib/routes";
import { Header } from "../../components/Header/Header";
import { nanosecondsToDuration } from "../../lib/time";

dayjs.extend(dayjsDuration);

type WebDataTimeEntriesList = ListTimeEntriesSchemaType | "loading" | "error"

export const PageTimeEntries = () => {
  const navigate = useNavigate();

  const [filterStartDate, setFilterStartDate] = useState<Date | null>(null);
  const [filterEndDate, setFilterEndDate] = useState<Date | null>(null);
  const [timeEntries, setTimeEntries] = useState<WebDataTimeEntriesList>("loading");

  const fetchTimeEntries = useCallback(async () => {
    const res = await api.timeEntries.list({
      startDate: filterStartDate,
      endDate: filterEndDate,
    });

    if (isLeft(res)) {
      console.error(`Failed to load time entries: ${res.left}`);
      setTimeEntries("error");
      return;
    }

    setTimeEntries(res.right);
  }, [filterStartDate, filterEndDate]);

  const onCreateInvoice = useCallback(() => {
    const qpEndDate = filterEndDate;
    if (qpEndDate !== null) {
      qpEndDate.setHours(23);
      qpEndDate.setMinutes(59);
      qpEndDate.setSeconds(59);
    }

    navigate(ROUTES.createInvoice.make({
      startDate: filterStartDate,
      endDate: qpEndDate,
    }));
  }, [filterStartDate, filterEndDate]);

  useEffect(() => {
    fetchTimeEntries();
  }, [fetchTimeEntries]);

  return (
    <>
      <Header />
      <Box
        sx={{
          padding: "1rem",
        }}
      >
        <Box
          sx={{
            display: "flex",
            flexDirection: "row",
            justifyContent: "space-between",
            marginBottom: "1rem",
          }}
        >
          <PageTimeFilters
            filterStartDate={filterStartDate}
            setFilterStartDate={setFilterStartDate}
            filterEndDate={filterEndDate}
            setFilterEndDate={setFilterEndDate}
          />

          <PageTimeInformation timeEntries={timeEntries} />

          <PageTimeActions
            onCreateInvoice={onCreateInvoice}
          />
        </Box>

        <TimeEntriesTable timeEntries={timeEntries} />
      </Box>
    </>
  );
};

const PageTimeFilters = ({
  filterStartDate,
  setFilterStartDate,
  filterEndDate,
  setFilterEndDate,
}: {
  readonly filterStartDate: Date | null
  readonly setFilterStartDate: (value: Date | null) => void
  readonly filterEndDate: Date | null
  readonly setFilterEndDate: (value: Date | null) => void
}) => {
  return (
    <Box>
      <Typography variant="h5">
        Filters
      </Typography>

      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <Box
          sx={{
            marginTop: "0.5rem",
          }}
        >
          <DateFilter
            label="Start Date"
            value={filterStartDate}
            onChange={setFilterStartDate}
          />
        </Box>

        <Box
          sx={{
            marginTop: "0.5rem",
          }}
        >
          <DateFilter
            label="End Date"
            value={filterEndDate}
            onChange={setFilterEndDate}
          />
        </Box>
      </Box>
    </Box>
  )
}

const PageTimeInformation = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Typography variant="h5">
        Information
      </Typography>

      <PageTimeInformationContent timeEntries={timeEntries} />
    </Box>
  )
}

const PageTimeInformationContent = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList
}) => {
  const [invoiceSettings, setInvoiceSettings] = useState<InvoiceSettingsSchemaType | "loading" | "error">("loading");

  const fetchInvoiceSettings = useCallback(async () => {
    const res = await api.invoiceSettings.get();
    if (isLeft(res)) {
      setInvoiceSettings("error");
      return;
    }

    setInvoiceSettings(res.right);
  }, []);

  useEffect(() => {
    fetchInvoiceSettings();
  }, [])

  if (timeEntries === "loading"  || invoiceSettings === "loading") {
    return (
      <>
        <CircularProgress size="small" />
        <Typography>
          Loading
        </Typography>
      </>
    );
  }

  if (timeEntries === "error" || invoiceSettings === "error") {
    return (
      <>
        <Typography>
          Failed to load time entries
        </Typography>
      </>
    )
  }

  const totalDuration = nanosecondsToDuration(timeEntries.total_duration);

  return (
    <TableContainer
      component={Paper}
      sx={{
        marginTop: "1rem",
      }}
    >
      <Table>
        <TableBody>
          <TableRow>
            <TableCell variant="head">Total Duration</TableCell>
            <TableCell>{totalDuration.format("YY-MM-DD HH:mm:ss")}</TableCell>
          </TableRow>

          <TableRow>
            <TableCell></TableCell>
            <TableCell>{totalDuration.asHours().toFixed(2)} Hour(s)</TableCell>
          </TableRow>

          <TableRow>
            <TableCell variant="head">Value</TableCell>
            <TableCell>${(invoiceSettings.hourly_rate * totalDuration.asHours()).toFixed(2)}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  );
};

const PageTimeActions = ({
  onCreateInvoice,
}: {
  readonly onCreateInvoice: () => void
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Typography variant="h5">
        Actions
      </Typography>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          marginTop: "1rem",
        }}
      >
        <Button
          variant="contained"
          onClick={onCreateInvoice}
        >
          Create Invoice
        </Button>
      </Box>
    </Box>
  );
};

const TimeEntriesTable = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList
}) => {
  // If loading
  if (timeEntries === "loading") {
    return (
      <>
        <CircularProgress />
        <Typography>Loading time entries</Typography>
      </>
    );
  }

  // If error
  if (timeEntries === "error") {
    return (
      <>
        <WarningIcon />
        <Typography>Failed to load time entries</Typography>
      </>
    );
  }

  // Show time entries table
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Start Time</TableCell>

            <TableCell>End Time</TableCell>

            <TableCell>Duration</TableCell>

            <TableCell>Comment</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {timeEntries.time_entries.map((timeEntry) => (
            <TimeEntryTableRow
              key={timeEntry.hash}
              timeEntry={timeEntry}
            />
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

const TimeEntryTableRow = ({
  timeEntry,
}: {
  readonly timeEntry: TimeEntrySchemaType
}) => {
  const duration = nanosecondsToDuration(timeEntry.duration);

  return (
    <TableRow>
      <TableCell>{timeEntry.start_time}</TableCell>
      <TableCell>{timeEntry.end_time}</TableCell>
      <TableCell>{duration.format("HH:mm:ss")}</TableCell>
      <TableCell>{timeEntry.comment}</TableCell>
    </TableRow>
  )
}