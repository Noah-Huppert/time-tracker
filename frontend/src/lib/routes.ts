export const ROUTES = {
  home: {
    pattern: "/",
    make: () => "/",
  },

  timeEntries: {
    pattern: "/time-entries",
    make: () => "/time-entries",
  },

  invoices: {
    pattern: "/invoices",
    make: () => "/invoices",
  },

  viewInvoice: {
    pattern: "/invoices/:id",
    make: ({
      invoiceID,
    }: {
      readonly invoiceID: number,
    }) => `/invoices/${invoiceID}`,
  }
};

export const QUERY_PARAMS = {
  invoice: {
    startDate: "start_date",
    endDate: "end_date",
  }
}