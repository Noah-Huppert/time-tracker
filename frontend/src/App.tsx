import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { PageTimeEntries } from "./pages/TimeEntries/TimeEntries";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";

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
        <RouterProvider router={router} />
      </LocalizationProvider>
    </>
  );
}

export default App;
