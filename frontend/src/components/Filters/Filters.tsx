import { Button, Popover } from "@mui/material";
import React, { useState } from "react";
import FilterAltIcon from '@mui/icons-material/FilterAlt';

/**
 * Takes an object with arbitrary string keys and turns it into a concrete type with exact keys and values.
 */
type MappedStringObject<O> = {
  [K in keyof O]: O[K]
}

/**
 * Details about a filter condition.
 */
type FilterCondition<V> = {
  /**
   * Display name of condition.
   */
  readonly name: string

  /**
   * @returns Starting value for filter
   */
  readonly start: () => V

  /**
   * Component which allows user to select a filter value
   */
  readonly component: FilterConditionComponent<V>
}

/**
 * Component which allows user to select filter value.
 */
type FilterConditionComponent<V> = ({
  value,
  setValue,
}: {
  readonly value: V | null
  readonly setValue: (value: V | null) => void
}) => JSX.Element;

export const Filters = <ValuesType, FilterValues extends {[key: string]: ValuesType | null},>({
  filterValues,
  setFilterValues,
  filterConditions,
}: {
  readonly filterValues: MappedStringObject<FilterValues>
  readonly setFilterValues: (values: MappedStringObject<FilterValues>) => void
  readonly filterConditions: {
    [K in keyof FilterValues]: FilterCondition<FilterValues[K]>
  }
}) => {
  const [selectFilterPopoverAnchorEl, setSelectFilterPopoverAnchorEl] = useState<HTMLButtonElement | null>(null);

  const selectedFilters = Object.keys(filterValues).filter((key) => filterValues[key] !== null);
  const nonSelectedFilters = Object.keys(filterValues).filter((key) => filterValues[key] === null);

  const setFilter = <K extends keyof FilterValues>(key: K, value: FilterValues[K] | null) => {
    setFilterValues({
      ...filterValues,
      [key]: value,
    })
  };

  return (
    <>
      <Button
        onClick={(e) => setSelectFilterPopoverAnchorEl(e.currentTarget)}
      >
        <FilterAltIcon />
      </Button>

      {selectedFilters.map((key) => (
        <React.Fragment key={key}>
          {filterConditions[key].component({
            value: filterValues[key],
            setValue: (value) => setFilter(key, value)
          })}
        </React.Fragment>
      ))}

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