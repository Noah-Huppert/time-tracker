import { Button, Chip } from "@mui/material";
import { DateTimePicker } from "@mui/x-date-pickers";
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import { useState } from "react";
import dayjs from "dayjs";

export const DateFilter = ({
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
};