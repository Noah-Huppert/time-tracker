export const ROUTES = {
  home: {
    pattern: "/",
    make: () => "/",
  },

  time_entries: {
    pattern: "/time-entries",
    make: () => "/time-entries",
  },

  invoices: {
    pattern: "/invoices",
    make: () => "/invoices",
  },

  createInvoice: {
    pattern: "/invoice",
    make: ({
      startDate,
      endDate,
    }: {
      readonly startDate: Date | null
      readonly endDate: Date | null
    }) => {
      // Include query parameters
      const queryParams: {[key: string]: string} = {};
      if (startDate !== null) {
        queryParams[QUERY_PARAMS.invoice.startDate] = startDate.toISOString();
      }

      if (endDate !== null) {
        queryParams[QUERY_PARAMS.invoice.endDate] = endDate.toISOString();
      }

      const queryParamsStr = new URLSearchParams(queryParams).toString();

      // URL
      return `/invoice?${queryParamsStr}`;
    }
  },
};

export const QUERY_PARAMS = {
  invoice: {
    startDate: "start_date",
    endDate: "end_date",
  }
}