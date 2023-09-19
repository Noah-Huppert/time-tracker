import { Button, CircularProgress, Drawer, TextField, Typography } from "@mui/material"
import { useCallback, useContext, useEffect, useState } from "react"
import { InvoiceSettingsSchemaType, api } from "../../lib/api"
import { isLeft } from "fp-ts/lib/Either"
import WarningIcon from "@mui/icons-material/Warning";
import { ToastCtx } from "../Toast/Toast";

type DraftInvoiceSettings = {
  hourly_rate: string
}

function draftInvoiceSettingsFromSchemaType(settings: InvoiceSettingsSchemaType): DraftInvoiceSettings {
  return {
    hourly_rate: settings.hourly_rate.toString(),
  }
}

function schemaInvoiceSettingsFromDraftType(settings: DraftInvoiceSettings): InvoiceSettingsSchemaType {
  return {
    hourly_rate: parseFloat(settings.hourly_rate),
  }
}

export const SettingsDrawer = ({
  open,
  setOpen,
}: {
  readonly open: boolean
  readonly setOpen: (value: boolean) => void
}) => {
  const toast = useContext(ToastCtx);
  const [invoiceSettings, setInvoiceSettings] = useState<DraftInvoiceSettings | "loading" | "error">("loading");
  const [updateInvoiceSettingsLoading, setUpdateInvoiceSettingsLoading] = useState(false);

  const fetchInvoiceSettings = useCallback(async () => {
    const res = await api.invoiceSettings.get();
    if (isLeft(res)) {
      console.error(`failed to get invoice settings: ${res.left}`);
      setInvoiceSettings("error");
      return;
    }

    setInvoiceSettings(draftInvoiceSettingsFromSchemaType(res.right));
  }, [setInvoiceSettings])

  useEffect(() => {
    fetchInvoiceSettings()
  }, [fetchInvoiceSettings]);

  const updateInvoiceSettings = useCallback(function<K extends keyof DraftInvoiceSettings, V extends DraftInvoiceSettings[K]>(key: K, value: V) {
    setInvoiceSettings((comp) => {
      if (comp === "loading" || comp === "error") {
        return comp;
      }

      return {
        ...comp,
        [key]: value,
      };
    });
  }, [])

  const sendInvoiceSettingsUpdate = useCallback(async () => {
    if (invoiceSettings === "loading" || invoiceSettings === "error") {
      return;
    }

    setUpdateInvoiceSettingsLoading(true);

    const comp = schemaInvoiceSettingsFromDraftType(invoiceSettings);
    const res = await api.invoiceSettings.set({
      hourlyRate: comp.hourly_rate,
    });

    setUpdateInvoiceSettingsLoading(false);

    if (isLeft(res)) {
      toast({
        kind: "error",
        message: "Failed to update invoice settings",
      });
      console.error(`Failed to update invoice settings: ${res.left}`);
      return;
    }

    setInvoiceSettings(draftInvoiceSettingsFromSchemaType(res.right));

    toast({
      kind: "success",
      message: "Updated invoice settings",
    })
  }, [invoiceSettings, setInvoiceSettings, setUpdateInvoiceSettingsLoading]);

  return (
    <Drawer
      open={open}
      onClose={() => setOpen(false)}
      anchor="right"
    >
      <Typography variant="h5">
        Settings
      </Typography>

      <Typography variant="h6">
        Invoice
      </Typography>
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
          <TextField
            label="Hourly Rate"
            value={invoiceSettings.hourly_rate}
            onChange={(e) => updateInvoiceSettings("hourly_rate", e.target.value)}
          />

          <Button
            disabled={updateInvoiceSettingsLoading}
            type="submit"
            onClick={sendInvoiceSettingsUpdate}
          >
            {updateInvoiceSettingsLoading && (
              <CircularProgress />
            )}
            Save
          </Button>
        </>
      )}
      
    </Drawer>
  ) 
}