import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { PageTimeEntries } from "./pages/TimeEntries/TimeEntries";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import { createTheme, CssBaseline, ThemeProvider } from "@mui/material";
import { ToastProvider } from "./components/Toast/Toast";
import { ROUTES } from "./lib/routes";
import { PageViewInvoice } from "./pages/ViewInvoice/ViewInvoice";
import { PageHome } from "./pages/Home/Home";
import { PageInvoices } from "./pages/Invoices/Invoices";

const theme = createTheme({
  typography: {
    button: {
      textTransform: "none",
    },
  },
});

const router = createBrowserRouter([
  {
    path: ROUTES.home.pattern,
    element: <PageHome />,
  },
  {
    path: ROUTES.timeEntries.pattern,
    element: <PageTimeEntries />,
  },
  {
    path: ROUTES.invoices.pattern,
    element: <PageInvoices />
  },
  {
    path: ROUTES.viewInvoice.pattern,
    element: <PageViewInvoice />,
  },
]);

function App() {
  return (
    <>
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <ThemeProvider theme={theme}>
          <ToastProvider>
            <CssBaseline />
            <RouterProvider router={router} />
          </ToastProvider>
        </ThemeProvider>
      </LocalizationProvider>
    </>
  );
}

export default App;
