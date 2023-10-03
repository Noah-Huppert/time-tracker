import { FormControlLabel, FormGroup, Switch } from "@mui/material";

export const BooleanFilter = ({
  name,
  value,
  setValue,
}: {
  readonly name: string;
  readonly value: boolean | null;
  readonly setValue: (value: boolean | null) => void;
  readonly hide: () => void,
}) => {
  return (
    <FormGroup>
      <FormControlLabel
        control={(
          <Switch
            checked={value || false}
            onChange={(e) => setValue(e.target.checked)}
          />
        )}
        label={name}
      />
    </FormGroup>
  );
};