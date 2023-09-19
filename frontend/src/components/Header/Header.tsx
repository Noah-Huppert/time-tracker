import { AppBar, Container, Drawer, IconButton, TextField, Toolbar, Typography } from "@mui/material"
import SettingsIcon from '@mui/icons-material/Settings';
import { useState } from "react";
import { SettingsDrawer } from "../SettingsDrawer/SettingsDrawer";

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
            <Typography sx={{
              flexGrow: 1,
            }}>
              Time Tracker
            </Typography>

            <IconButton onClick={() => setSettingsOpen(true)}>
              <SettingsIcon />
            </IconButton>
          </Toolbar>
        </Container>
      </AppBar>
    </>
  )
}