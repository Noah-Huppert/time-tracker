import { Button, Chip } from "@mui/material";
import { DatePicker, DateTimePicker } from "@mui/x-date-pickers";
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

  const onSetButtonClick = () => {
    setShowingSelector(true)
  }

  if (value === null && showingSelector === false) {
    return (
      <>
        <Button
          startIcon={<AddCircleOutlineIcon />}
          variant="outlined"
          onClick={onSetButtonClick}
        >
          {label}
        </Button>
      </>
    );
  }

  if (showingSelector === true) {
    return (
      <>
        <DatePicker
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
        label={`${label}: ${dayjs(value).format("YYYY-MM-DD")}`}
        onDelete={() => onChange(null)}
        onClick={() => setShowingSelector(true)}
        variant="filled"
        color="primary"
      />
    </>
  );
};