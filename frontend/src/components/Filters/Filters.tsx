import { Button, Popover } from "@mui/material";
import { useState } from "react";
import FilterAltIcon from '@mui/icons-material/FilterAlt';

export const Filters = <ValuesType, FilterValues extends {[key: string]: ValuesType | null},>({
  filterValues,
  setFilter,
  filterConditions,
}: {
  readonly filterValues: { [Property in keyof FilterValues]: FilterValues[keyof FilterValues] }
  readonly setFilter: <K extends keyof FilterValues,>(key: K, value: FilterValues[K] | null) => void
  readonly filterConditions: {
    [Property in keyof FilterValues]: {
      readonly name: string
      readonly start: () => FilterValues[keyof FilterValues]
      readonly component: ({
        value,
        setValue,
      }: {
        readonly value: FilterValues[Property] | null
        readonly setValue: (value: FilterValues[Property] | null) => void
      }) => JSX.Element
    }
  }
}) => {
  const [selectFilterPopoverAnchorEl, setSelectFilterPopoverAnchorEl] = useState<HTMLButtonElement | null>(null);

  const selectedFilters = Object.keys(filterValues).filter((key) => filterValues[key] !== null);
  console.log("selectedFilters=", selectedFilters, Object.keys(filterValues).map((key) => `${key}=${filterValues[key]}`))
  const nonSelectedFilters = Object.keys(filterValues).filter((key) => filterValues[key] === null);

  return (
    <>
      <Button
        onClick={(e) => setSelectFilterPopoverAnchorEl(e.currentTarget)}
      >
        <FilterAltIcon />
      </Button>

      {selectedFilters.map((key) => filterConditions[key].component({
          value: filterValues[key],
          setValue: (value) => setFilter(key, value)
        })
      )}

      <Popover
        open={selectFilterPopoverAnchorEl !== null}
        anchorEl={selectFilterPopoverAnchorEl}
        onClose={() => setSelectFilterPopoverAnchorEl(null)}
      >
        {nonSelectedFilters.map((key) => (
          <Button
            key={key}
            onClick={() => setFilter(key, filterConditions[key].start())}
          >
            {filterConditions[key].name}
          </Button>
        ))}
      </Popover>
    </>
  );
};