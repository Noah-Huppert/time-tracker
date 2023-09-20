import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { PageTimeEntries } from "./pages/TimeEntries/TimeEntries";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";
import { Header } from "./components/Header/Header";

import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import { createTheme, CssBaseline, ThemeProvider } from "@mui/material";
import { ToastProvider } from "./components/Toast/Toast";

const theme = createTheme({      
  typography: {
    button: {
      textTransform: 'none'
    }
  }
});


const router = createBrowserRouter([
  {
    path: "/",
    element: <PageTimeEntries />,
  },
]);

function App() {
  return (
    <>
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <ThemeProvider theme={theme}>
          <ToastProvider>
            <CssBaseline />
            <Header />
            <RouterProvider router={router} />
          </ToastProvider>
        </ThemeProvider>
      </LocalizationProvider>
    </>
  );
}

export default App;
