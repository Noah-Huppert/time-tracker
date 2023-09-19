import { useCallback, useEffect, useState } from "react";
import { api, TimeEntrySchemaType } from "../../lib/api";
import { CircularProgress, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from "@mui/material";
import WarningIcon from '@mui/icons-material/Warning';
import { isLeft } from "fp-ts/lib/Either";

export const PageTimeEntries = () => {
  const [timeEntries, setTimeEntries] = useState<TimeEntrySchemaType[] | "loading" | "error">("loading")

  const fetchTimeEntries = useCallback(async () => {
    const res = await api.timeEntries.list();
    if (isLeft(res)) {
      console.error(`Failed to load time entries: ${res.left}`);
      setTimeEntries("error");
      return;
    }

    setTimeEntries(res.right.time_entries);
  }, [])

  useEffect(() => {
    fetchTimeEntries()
  })

  // If loading
  if (timeEntries === "loading") {
    return (
      <>
        <CircularProgress />
        <p>
          Loading time entries
        </p>
      </>
    );
  }

  // If error
  if (timeEntries === "error") {
    return (
      <>
        <WarningIcon />
        <p>
          Failed to load time entries
        </p>
      </>
    )
  }

  // Show time entries table
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>
              Start Time
            </TableCell>

            <TableCell>
              End Time
            </TableCell>

            <TableCell>
              Comment
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {timeEntries.map((timeEntry) => (
            <TableRow key={timeEntry.hash}>
              <TableCell>
                {timeEntry.start_time}
              </TableCell>
              <TableCell>
                {timeEntry.end_time}
              </TableCell>
              <TableCell>
                {timeEntry.comment}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};