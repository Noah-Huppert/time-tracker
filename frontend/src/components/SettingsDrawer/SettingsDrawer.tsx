import {
  Box,
  Button,
  CircularProgress,
  Drawer,
  TextField,
  Typography,
} from "@mui/material";
import { useCallback, useContext, useEffect, useState } from "react";
import { InvoiceSettingsSchemaType, api } from "../../lib/api";
import { isLeft } from "fp-ts/lib/Either";
import WarningIcon from "@mui/icons-material/Warning";
import { ToastCtx } from "../Toast/Toast";

import "./SettingsDrawer.css";

type DraftInvoiceSettings = {
  hourly_rate: string;
  recipient: string;
  sender: string;
};

function emptyInvoiceSettingsDraft(): DraftInvoiceSettings {
  return {
    hourly_rate: "",
    recipient: "",
    sender: "",
  };
}

function draftInvoiceSettingsFromSchemaType(
  settings: InvoiceSettingsSchemaType,
): DraftInvoiceSettings {
  return {
    hourly_rate: settings.hourly_rate.toString(),
    recipient: settings.recipient,
    sender: settings.sender,
  };
}

function schemaInvoiceSettingsFromDraftType(
  settings: DraftInvoiceSettings,
): Omit<InvoiceSettingsSchemaType, "id"> {
  return {
    hourly_rate: parseFloat(settings.hourly_rate),
    recipient: settings.recipient,
    sender: settings.sender,
  };
}

export const SettingsDrawer = ({
  open,
  setOpen,
}: {
  readonly open: boolean;
  readonly setOpen: (value: boolean) => void;
}) => {
  const toast = useContext(ToastCtx);
  const [invoiceSettings, setInvoiceSettings] = useState<
    DraftInvoiceSettings | "loading" | "error"
  >("loading");
  const [updateInvoiceSettingsLoading, setUpdateInvoiceSettingsLoading] =
    useState(false);

  const fetchInvoiceSettings = useCallback(async () => {
    const res = await api.invoiceSettings.get();
    if (isLeft(res)) {
      console.error(`failed to get invoice settings: ${res.left}`);
      setInvoiceSettings("error");
      return;
    }

    if (res.right === null) {
      setInvoiceSettings(emptyInvoiceSettingsDraft());
      return;
    }

    setInvoiceSettings(draftInvoiceSettingsFromSchemaType(res.right));
  }, [setInvoiceSettings]);

  useEffect(() => {
    fetchInvoiceSettings();
  }, [fetchInvoiceSettings]);

  const updateInvoiceSettings = useCallback(function <
    K extends keyof DraftInvoiceSettings,
    V extends DraftInvoiceSettings[K],
  >(key: K, value: V) {
    setInvoiceSettings((comp) => {
      if (comp === "loading" || comp === "error") {
        return comp;
      }

      return {
        ...comp,
        [key]: value,
      };
    });
  }, []);

  const sendInvoiceSettingsUpdate = useCallback(async () => {
    if (invoiceSettings === "loading" || invoiceSettings === "error") {
      return;
    }

    // Set settings
    setUpdateInvoiceSettingsLoading(true);

    const settings = schemaInvoiceSettingsFromDraftType(invoiceSettings);
    const res = await api.invoiceSettings.set({
      hourlyRate: settings.hourly_rate,
      recipient: settings.recipient,
      sender: settings.sender,
    });

    setUpdateInvoiceSettingsLoading(false);

    // Handle error
    if (isLeft(res)) {
      toast({
        kind: "error",
        message: "Failed to update invoice settings",
      });
      console.error(`Failed to update invoice settings: ${res.left}`);
      return;
    }

    // Check if no settings exist
    if (res.right === null) {
      setInvoiceSettings(emptyInvoiceSettingsDraft());
      return;
    }

    setInvoiceSettings(draftInvoiceSettingsFromSchemaType(res.right));

    toast({
      kind: "success",
      message: "Updated invoice settings",
    });
  }, [
    toast,
    invoiceSettings,
    setInvoiceSettings,
    setUpdateInvoiceSettingsLoading,
  ]);

  return (
    <Drawer open={open} onClose={() => setOpen(false)} anchor="right">
      <Box
        sx={{
          padding: "1rem",
          display: "flex",
          flexDirection: "column",
          minWidth: "30rem",
        }}
      >
        <Typography variant="h5">Settings</Typography>

        <Typography
          variant="h6"
          sx={{
            marginBottom: "1rem",
          }}
        >
          Invoice
        </Typography>

        <Box
          sx={{
            paddingLeft: "2rem",
            paddingRight: "2rem",
            paddingBottom: "2rem",
          }}
        >
          {invoiceSettings === "loading" ? (
            <>
              <CircularProgress />
              <Typography>Loading invoice settings</Typography>
            </>
          ) : invoiceSettings === "error" ? (
            <>
              <WarningIcon />
              <Typography>Failed to load invoice settings</Typography>
            </>
          ) : (
            <>
              <Box className="setting-input">
                <TextField
                  label="Hourly Rate"
                  value={invoiceSettings.hourly_rate}
                  onChange={(e) =>
                    updateInvoiceSettings("hourly_rate", e.target.value)
                  }
                  sx={{
                    width: "100%",
                  }}
                />
              </Box>

              <Box className="setting-input">
                <TextField
                  multiline
                  rows={3}
                  label="Recipient"
                  value={invoiceSettings.recipient}
                  onChange={(e) =>
                    updateInvoiceSettings("recipient", e.target.value)
                  }
                  sx={{
                    width: "100%",
                  }}
                />
              </Box>

              <Box className="setting-input">
                <TextField
                  multiline
                  rows={3}
                  label="Sender"
                  value={invoiceSettings.sender}
                  onChange={(e) =>
                    updateInvoiceSettings("sender", e.target.value)
                  }
                  sx={{
                    width: "100%",
                  }}
                />
              </Box>

              <Button
                disabled={updateInvoiceSettingsLoading}
                type="submit"
                onClick={sendInvoiceSettingsUpdate}
                variant="outlined"
              >
                {updateInvoiceSettingsLoading && <CircularProgress />}
                Save
              </Button>
            </>
          )}
        </Box>
      </Box>
    </Drawer>
  );
};
