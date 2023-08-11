# Time Tracker
Track time spent on projects.

# Table Of Contents
- [Overview](#overview)

# Overview
Aggregates multiple time tracking spreadsheets and computes reports.

Use the `-help` option for details about available options.

Time tracking spreadsheets must have the following columns:

- Start time (Default name: `time started`, format: `YYYY-MM-DD HH:MM:SS`)
- End time (Default name: `time ended`, format: `YYYY-MM-DD HH:MM:SS`)
- Comment (Default name: `comment`)

Use the `-column-start-time`, `-column-end-time`, and `-column-comment` options to specify the names of these columns.

All times must be from the same time zone. Although the time zone cannot be specified in spreadsheets it can be specified by the `-timezone` option (Default: `EST`).

Place all time tracking spreadsheets in an input directory. Use the `-in-dir` option to customize the location of this directory, by default it is the `times/` directory.

The tool will combine times into billing periods. Available billing periods are: `weekly`, `bi-weekly` (Default), and `monthly`. Billing periods start on the first of the month. Use the `-billing-period` option to set the billing period.

The amount owed for each billing period is calculated using an hour rate, which is set via the `-hourly-rate` option.

The tool can output results depending on the output format set by the `-output` option:

- `print`: Prints results to the console
- `dir=<DIR>`: Writes report CSV files to the `<DIR>` directory. One CSV containing time entries is written for each billing period. A CSV with all the billing periods summarized is written as well.

A typical run of the tool:

```
go run . -billing-period monthly -hourly-rate 30.0
```

A more advanced run with every option:
```
go run . \
    -column-start-date start \
    -column-end-date end \
    -timezone PDT \
    -in-dir ./time-sheets \
    -billing-period weekly \
    -hourly-rate 30.0 \
    -output dir=out-reports
```