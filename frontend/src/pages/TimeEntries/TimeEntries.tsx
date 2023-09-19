import { useCallback, useEffect, useState } from "react";
import { api, TimeEntrySchemaType } from "../../lib/api";
import {
  Button,
  CircularProgress,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@mui/material";
import WarningIcon from "@mui/icons-material/Warning";
import { isLeft } from "fp-ts/lib/Either";
import { DateTimePicker } from "@mui/x-date-pickers";

type WebDataTimeEntriesList = TimeEntrySchemaType[] | "loading" | "error"

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

    setTimeEntries(res.right.time_entries);
  }, [filterStartTime, filterEndTime]);

  useEffect(() => {
    fetchTimeEntries();
  }, [fetchTimeEntries]);

  return (
    <>
      <DateFilter
        label="Start Time"
        value={filterStartTime}
        onChange={setFilterStartTime}
      />

      <DateFilter
        label="End Time"
        value={filterEndTime}
        onChange={setFilterEndTime}
      />
      <TimeEntriesTable timeEntries={timeEntries} />
    </>
  )
};

const DateFilter = ({
  label,
  value,
  onChange,
}: {
  readonly label: string
  readonly value: Date | null
  readonly onChange: (date: Date | null) => void
}) => {
  return (
    <>
      <p>
        {label}
      </p>

      <DateTimePicker
        value={value}
        onChange={onChange}
      />

      {value !== null && (
        <Button onClick={() => onChange(null)}>
          Clear
        </Button>
      )}
    </>
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
        <p>Loading time entries</p>
      </>
    );
  }

  // If error
  if (timeEntries === "error") {
    return (
      <>
        <WarningIcon />
        <p>Failed to load time entries</p>
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

            <TableCell>Comment</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {timeEntries.map((timeEntry) => (
            <TableRow key={timeEntry.hash}>
              <TableCell>{timeEntry.start_time}</TableCell>
              <TableCell>{timeEntry.end_time}</TableCell>
              <TableCell>{timeEntry.comment}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}