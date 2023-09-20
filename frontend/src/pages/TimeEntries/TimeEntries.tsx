import { useCallback, useEffect, useState } from "react";
import { api, ListTimeEntriesSchemaType, TimeEntrySchemaType } from "../../lib/api";
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  IconButton,
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
import { DateTimePicker } from "@mui/x-date-pickers";
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import RemoveCircleOutlineIcon from '@mui/icons-material/RemoveCircleOutline';
import dayjsDuration, { Duration } from "dayjs/plugin/duration";
import dayjs from "dayjs";

dayjs.extend(dayjsDuration);

const MILLISECONDS_PER_NANOSECOND = 1e6;

function nanosecondsToDuration(nanoseconds: number): Duration {
  return dayjs.duration(nanoseconds / MILLISECONDS_PER_NANOSECOND, "milliseconds");
}

type WebDataTimeEntriesList = ListTimeEntriesSchemaType | "loading" | "error"

export const PageTimeEntries = () => {
  const [filterStartTime, setFilterStartTime] = useState<Date | null>(null);
  const [filterEndTime, setFilterEndTime] = useState<Date | null>(null);
  const [timeEntries, setTimeEntries] = useState<WebDataTimeEntriesList>("loading");

  const fetchTimeEntries = useCallback(async () => {
    const res = await api.timeEntries.list({
      startTime: filterStartTime,
      endTime: filterEndTime,
    });

    if (isLeft(res)) {
      console.error(`Failed to load time entries: ${res.left}`);
      setTimeEntries("error");
      return;
    }

    setTimeEntries(res.right);
  }, [filterStartTime, filterEndTime]);

  useEffect(() => {
    fetchTimeEntries();
  }, [fetchTimeEntries]);

  return (
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
          filterStartTime={filterStartTime}
          setFilterStartTime={setFilterStartTime}
          filterEndTime={filterEndTime}
          setFilterEndTime={setFilterEndTime}
        />

        <PageTimeInformation timeEntries={timeEntries} />

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
            >
              Create Invoice
            </Button>
          </Box>
        </Box>
      </Box>

      <TimeEntriesTable timeEntries={timeEntries} />
    </Box>
  )
};

const PageTimeFilters = ({
  filterStartTime,
  setFilterStartTime,
  filterEndTime,
  setFilterEndTime,
}: {
  readonly filterStartTime: Date | null
  readonly setFilterStartTime: (value: Date | null) => void
  readonly filterEndTime: Date | null
  readonly setFilterEndTime: (value: Date | null) => void
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
            label="Start Time"
            value={filterStartTime}
            onChange={setFilterStartTime}
          />
        </Box>

        <Box
          sx={{
            marginTop: "0.5rem",
          }}
        >
          <DateFilter
            label="End Time"
            value={filterEndTime}
            onChange={setFilterEndTime}
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
  if (timeEntries === "loading") {
    return (
      <>
        <CircularProgress size="small" />
        <Typography>
          Loading
        </Typography>
      </>
    );
  }

  if (timeEntries === "error") {
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
        </TableBody>
      </Table>
    </TableContainer>
  );
}

const DateFilter = ({
  label,
  value,
  onChange,
}: {
  readonly label: string
  readonly value: Date | null
  readonly onChange: (date: Date | null) => void
}) => {
  const [showingSelector, setShowingSelector] = useState(false);

  if (value === null && showingSelector === false) {
    return (
      <>
        <Button
          startIcon={<AddCircleOutlineIcon />}
          variant="outlined"
          onClick={() => setShowingSelector(true)}
        >
          {label}
        </Button>
      </>
    );
  }

  if (showingSelector === true) {
    return (
      <>
        <DateTimePicker
          label={label}
          value={value}
          onChange={onChange}
          open={true}
          onClose={() => setShowingSelector(false)}
        />
      </>
    );
  }

  return (
    <>
      <Chip
        label={`${label}: ${dayjs(value).format("YYYY-MM-DD HH:mm:ss")}`}
        onDelete={() => onChange(null)}
        variant="filled"
        color="primary"
      />
    </>
  );

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "row",
        marginTop: "1rem",
        alignItems: "center",
      }}
    >
      <DateTimePicker
        label={label}
        value={value}
        onChange={onChange}
      />

      {value !== null && (
        <Box>
          <Button
            variant="outlined"
            size="small"
            onClick={() => onChange(null)}
            startIcon={<HighlightOffIcon />}
            sx={{
              marginLeft: "1rem",
            }}
          >
            Clear
          </Button>
        </Box>
      )}
    </Box>
  )
}

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