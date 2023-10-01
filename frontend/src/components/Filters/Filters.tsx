import { Box, Button, Chip, Popover } from "@mui/material";
import React, { useState } from "react";
import FilterAltIcon from "@mui/icons-material/FilterAlt";

/**
 * Takes an object with arbitrary string keys and turns it into a concrete type with exact keys and values.
 */
type MappedStringObject<O> = {
  [K in keyof O]: O[K];
};

/**
 * Details about a filter condition.
 */
type FilterCondition<V> = {
  /**
   * Display name of condition.
   */
  readonly name: string;

  /**
   * @returns Starting value for filter
   */
  readonly start: () => V;

  /**
   * @returns Human friendly representation of the value
   */
  readonly display: (value: V) => string;

  /**
   * Component which allows user to select a filter value
   */
  readonly component: FilterConditionComponent<V>;
};

/**
 * Component which allows user to select filter value.
 */
type FilterConditionComponent<V> = ({
  name,
  value,
  setValue,
  hide,
}: {
  readonly name: string;
  readonly value: V | null;
  readonly setValue: (value: V | null) => void;
  readonly hide: () => void,
}) => JSX.Element;

export const Filters = <
  ValuesType,
  FilterValues extends { [key: string]: ValuesType | null },
>({
  filterValues,
  setFilterValues,
  filterConditions,
}: {
  readonly filterValues: MappedStringObject<FilterValues>;
  readonly setFilterValues: (values: MappedStringObject<FilterValues>) => void;
  readonly filterConditions: {
    [K in keyof FilterValues]: FilterCondition<FilterValues[K]>;
  };
}) => {
  const [selectFilterPopoverAnchorEl, setSelectFilterPopoverAnchorEl] =
    useState<HTMLButtonElement | null>(null);
  const [showingFilterComponent, setShowingFilterComponent] = useState<Array<keyof FilterValues>>([]);

  // Categorize which filter conditions are currently applied
  const selectedFilters = Object.keys(filterValues).filter(
    (key) => filterValues[key] !== null || showingFilterComponent.includes(key)
  );
  const nonSelectedFilters = Object.keys(filterValues).filter(
    (key) => filterValues[key] === null
  );

  const setFilter = <K extends keyof FilterValues>(
    key: K,
    value: FilterValues[K] | null,
  ) => {
    showFilterComponent(key, false);

    setFilterValues({
      ...filterValues,
      [key]: value,
    });
  };

  const showFilterComponent = <K extends keyof FilterValues>(
    key: K,
    show: boolean
  ) => {
    setShowingFilterComponent((showingFilterComponent) => {
      if (show == showingFilterComponent.includes(key)) {
        return showingFilterComponent;
      } else if (show === true) {
        return [
          ...showingFilterComponent,
          key,
        ];
      } else {
        return [
          ...showingFilterComponent.filter((item) => item !== key),
        ];
      }
    });
  }

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "row",
      }}
    >
      <Button
        disabled={nonSelectedFilters.length === 0}
        onClick={(e) => setSelectFilterPopoverAnchorEl(e.currentTarget)}
        variant="contained"
        sx={{
          marginRight: "1rem",
        }}
      >
        <FilterAltIcon />
      </Button>

      <Box
        sx={{
          display: "flex",
          flexDirection: "row",
        }}
      >
        {selectedFilters.map((key) => (
          <Box
            key={key}
            sx={{
              marginLeft: "0.25rem",
              marginRight: "0.25rem",
            }}
          >
            {showingFilterComponent.includes(key) ?
              filterConditions[key].component({
                name: filterConditions[key].name,
                value: filterValues[key],
                setValue: (value) => setFilter(key, value),
                hide: () => showFilterComponent(key, false),
              })
            : (
              <Chip
                label={`${filterConditions[key].name}: ${filterConditions[key].display(filterValues[key])}`}
                onDelete={() => setFilter(key, null)}
                onClick={() => showFilterComponent(key, true)}
                variant="filled"
                color="primary"
              />
            )}
          </Box>
        ))}
      </Box>

      <Popover
        open={selectFilterPopoverAnchorEl !== null}
        anchorEl={selectFilterPopoverAnchorEl}
        onClose={() => setSelectFilterPopoverAnchorEl(null)}
        anchorOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
          }}
        >
          {nonSelectedFilters.map((key) => (
            <Button
              key={key}
              onClick={() => {
                setSelectFilterPopoverAnchorEl(null);
                setFilter(key, filterConditions[key].start());
                showFilterComponent(key, true);
              }}
            >
              {filterConditions[key].name}
            </Button>
          ))}
        </Box>
      </Popover>
    </Box>
  );
};
