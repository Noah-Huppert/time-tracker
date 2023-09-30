import { AppBar, Box, Container, Drawer, IconButton, TextField, Toolbar, Typography } from "@mui/material"
import SettingsIcon from '@mui/icons-material/Settings';
import { useState } from "react";
import { SettingsDrawer } from "../SettingsDrawer/SettingsDrawer";
import { Link } from "react-router-dom";
import { ROUTES } from "../../lib/routes";

export const Header = () => {
  const [settingsOpen, setSettingsOpen] = useState(false);

  return (
    <>
      <SettingsDrawer
        open={settingsOpen}
        setOpen={setSettingsOpen}
      />

      <AppBar component="nav" position="static">
        <Container>
          <Toolbar>
            <Box
              sx={{
                display: "flex",
                flexGrow: 1,
                flexDirection: "row",
                alignItems: "center",
              }}
            >
              <Typography variant="h6">
                Time Tracker
              </Typography>

              <Box
                sx={{
                  display: "flex",
                  flexDirection: "row",
                  justifyContent: "space-around",
                  flexGrow: "0.1",
                  marginLeft: "1rem",
                  "a": {
                    color: "white",
                    textDecoration: "none",
                  },
                }}
              >
                <Link to={ROUTES.home.make()}>
                  Home
                </Link>

                <Link to={ROUTES.time_entries.make()}>
                  Time Entries
                </Link>

                <Link
                  to={ROUTES.invoices.make()}
                >
                  Invoices
                </Link>
              </Box>
            </Box>

            <IconButton onClick={() => setSettingsOpen(true)}>
              <SettingsIcon />
            </IconButton>
          </Toolbar>
        </Container>
      </AppBar>
    </>
  )
}