import { useCallback, useContext, useEffect, useState } from "react";
import {
  api,
  InvoiceSettingsSchemaType,
  ListTimeEntriesSchemaType,
  TimeEntrySchemaType,
} from "../../lib/api";
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
import { DateFilter } from "../../components/Filters/DateFilter";
import { useNavigate } from "react-router-dom";
import { ROUTES } from "../../lib/routes";
import { Header } from "../../components/Header/Header";
import { DATE_FORMAT, DATE_TIME_FORMAT, nanosecondsToDuration } from "../../lib/time";
import {
  ReadFile,
  ReadFiles,
  UploadFile,
} from "../../components/UploadFile/UploadFile";
import { ToastCtx } from "../../components/Toast/Toast";
import { Filters } from "../../components/Filters/Filters";

dayjs.extend(dayjsDuration);

type WebDataTimeEntriesList = ListTimeEntriesSchemaType | "loading" | "error";

type TimeEntriesFilters = {
  startDate: Date | null,
  endDate: Date | null,
}

export const PageTimeEntries = () => {
  const navigate = useNavigate();
  const toast = useContext(ToastCtx);

  const [filters, setFilters] = useState<TimeEntriesFilters>({
    startDate: null,
    endDate: null,
  });
  const [timeEntries, setTimeEntries] =
    useState<WebDataTimeEntriesList>("loading");

  const fetchTimeEntries = useCallback(async () => {
    setTimeEntries("loading");

    const res = await api.timeEntries.list({
      startDate: filters.startDate,
      endDate: filters.endDate,
    });

    if (isLeft(res)) {
      console.error(`Failed to load time entries: ${res.left}`);
      setTimeEntries("error");
      return;
    }

    setTimeEntries(res.right);
  }, [filters.startDate, filters.endDate]);

  const onCreateInvoice = useCallback(async () => {
    if (filters.startDate === null) {
      toast({
        kind: "error",
        message: "Start date required to make an invoice",
      });
      return;
    }

    if (filters.endDate === null) {
      toast({
        kind: "error",
        message: "End date required to make an invoice",
      });
      return;
    }

    const invoiceSettings = await api.invoiceSettings.get();
    if (isLeft(invoiceSettings)) {
      console.error(`Failed to get invoice settings: ${invoiceSettings.left}`);
      toast({
        kind: "error",
        message:
          "Failed to create invoice, problem while getting invoice settings",
      });
      return;
    }

    if (invoiceSettings.right === null) {
      toast({
        kind: "error",
        message: "Failed to create invoice, please configure invoice settings first",
      });
      return;
    }

    const startDateOpt = filters.startDate;
    startDateOpt.setHours(0);
    startDateOpt.setMinutes(0);
    startDateOpt.setSeconds(0);
    startDateOpt.setMilliseconds(0);

    const endDateOpt = filters.endDate;
    endDateOpt.setHours(23);
    endDateOpt.setMinutes(59);
    endDateOpt.setSeconds(59);
    endDateOpt.setMilliseconds(999);

    console.log("calling create invoice", 
      {
        invoiceSettingsID: invoiceSettings.right.id,
        startDate: startDateOpt,
        endDate: endDateOpt,
      }
    )
    const invoice = await api.invoices.create({
      invoiceSettingsID: invoiceSettings.right.id,
      startDate: startDateOpt,
      endDate: endDateOpt,
    });
    if (isLeft(invoice)) {
      console.error(`Failed to create invoice: ${invoice.left}`);
      toast({
        kind: "error",
        message: "Failed to create invoice",
      });
      return;
    }

    navigate(
      ROUTES.viewInvoice.make({
        invoiceID: invoice.right.id,
      }),
    );
  }, [filters.startDate, filters.endDate, navigate, toast]);

  const onUploadTimeSheets = useCallback(
    async (fileList: FileList) => {
      // Get file content
      const readFiles: ReadFile[] = [];

      try {
        readFiles.push(...(await ReadFiles(fileList)));
      } catch (e) {
        console.error(`Failed to read time entry CSV files: ${e}`);
        toast({
          kind: "error",
          message: "Failed to read time entry CSV files",
        });
        return;
      }

      // Make request
      const res = await api.timeEntries.uploadCSV({
        csvFiles: readFiles,
      });
      if (isLeft(res)) {
        console.error(`Failed to upload time entries CSV: ${res.left}`);
        toast({
          kind: "error",
          message: "Failed to upload time entry CSV files",
        });
        return;
      }

      toast({
        kind: "success",
        message: `Created ${res.right.new_time_entries.length} time entry(s) (${res.right.existing_time_entries.length} time entry(s) already existed)`,
      });

      await fetchTimeEntries();
    },
    [fetchTimeEntries, toast],
  );

  useEffect(() => {
    fetchTimeEntries();
  }, [fetchTimeEntries]);

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Header />
      <Box
        sx={{
          padding: "1rem",
          maxWidth: "50rem",
          alignSelf: "center",
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
          <PageTimeInformation timeEntries={timeEntries} />

          <PageTimeActions
            onCreateInvoice={onCreateInvoice}
            onUploadTimeSheets={onUploadTimeSheets}
          />
        </Box>

        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
          }}
        >
          <Box
            sx={{
              marginBottom: "1rem",
            }}
          >
            <Filters
              filterValues={filters}
              setFilterValues={setFilters}
              filterConditions={{
                startDate: {
                  name: "Start Date",
                  start: () => null,
                  display: (value) => dayjs(value).format(DATE_FORMAT),
                  component: DateFilter,
                },
                endDate: {
                  name: "End Date",
                  start: () => null,
                  display: (value) => dayjs(value).format(DATE_FORMAT),
                  component: DateFilter,
                },
              }}
            />
          </Box>

          <TimeEntriesTable timeEntries={timeEntries} />
        </Box>
      </Box>
    </Box>
  );
};

const PageTimeInformation = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Typography variant="h5">Information</Typography>

      <PageTimeInformationContent timeEntries={timeEntries} />
    </Box>
  );
};

const PageTimeInformationContent = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList;
}) => {
  const [invoiceSettings, setInvoiceSettings] = useState<
    InvoiceSettingsSchemaType | "loading" | "error" | "notfound"
  >("loading");

  const fetchInvoiceSettings = useCallback(async () => {
    const res = await api.invoiceSettings.get();
    if (isLeft(res)) {
      setInvoiceSettings("error");
      return;
    }

    setInvoiceSettings(res.right || "notfound");
  }, []);

  useEffect(() => {
    fetchInvoiceSettings();
  }, [fetchInvoiceSettings]);

  if (timeEntries === "loading" || invoiceSettings === "loading") {
    return (
      <>
        <CircularProgress size="small" />
        <Typography>Loading</Typography>
      </>
    );
  }

  if (timeEntries === "error" || invoiceSettings === "error") {
    return (
      <>
        <Typography>Failed to load time entries</Typography>
      </>
    );
  }

  if (invoiceSettings === "notfound") {
    return (
      <>
        <Typography>Please configure invoice settings to see more information</Typography>
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
            <TableCell>
              $
              {(invoiceSettings.hourly_rate * totalDuration.asHours()).toFixed(
                2,
              )}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  );
};

const PageTimeActions = ({
  onCreateInvoice,
  onUploadTimeSheets,
}: {
  readonly onCreateInvoice: () => void;
  readonly onUploadTimeSheets: (fileList: FileList) => void;
}) => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Typography variant="h5">Actions</Typography>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          marginTop: "1rem",
        }}
      >
        <Box
          sx={{
            display: "flex",
          }}
        >
          <Button variant="contained" onClick={onCreateInvoice}>
            Create Invoice
          </Button>
        </Box>

        <Box
          sx={{
            display: "flex",
            marginTop: "1rem",
          }}
        >
          <UploadFile
            onUpload={onUploadTimeSheets}
            uploadLabel="Upload Time Sheets"
          />
        </Box>
      </Box>
    </Box>
  );
};

const TimeEntriesTable = ({
  timeEntries,
}: {
  readonly timeEntries: WebDataTimeEntriesList;
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
              key={`time-entry-table-id-${timeEntry.id}`}
              timeEntry={timeEntry}
            />
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

const TimeEntryTableRow = ({
  timeEntry,
}: {
  readonly timeEntry: TimeEntrySchemaType;
}) => {
  const duration = nanosecondsToDuration(timeEntry.duration);

  return (
    <TableRow>
      <TableCell>{dayjs(timeEntry.start_time).format(DATE_TIME_FORMAT)}</TableCell>
      <TableCell>{dayjs(timeEntry.end_time).format(DATE_TIME_FORMAT)}</TableCell>
      <TableCell>{duration.format("HH:mm:ss")}</TableCell>
      <TableCell>{timeEntry.comment}</TableCell>
    </TableRow>
  );
};
