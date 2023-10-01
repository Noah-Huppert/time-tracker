import { DatePicker } from "@mui/x-date-pickers";
import dayjs from "dayjs";

export const DateFilter = ({
  name,
  value,
  setValue,
  hide,
}: {
  readonly name: string;
  readonly value: Date | null;
  readonly setValue: (value: Date | null) => void;
  readonly hide: () => void,
}) => {
  return (
    <>
      <DatePicker<dayjs.Dayjs>
        label={name}
        value={dayjs(value)}
        onChange={(d) => setValue(d?.toDate() || null)}
        onClose={hide}
        open={true}
      />
    </>
  )
}