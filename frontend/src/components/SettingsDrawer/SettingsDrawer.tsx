import { Button, CircularProgress, Drawer, TextField, Typography } from "@mui/material"
import { useCallback, useContext, useEffect, useState } from "react"
import { CompensationSchemaType, api } from "../../lib/api"
import { isLeft } from "fp-ts/lib/Either"
import WarningIcon from "@mui/icons-material/Warning";
import { ToastCtx } from "../Toast/Toast";

type DraftCompensation = {
  hourly_rate: string
}

function draftCompensationFromSchemaType(comp: CompensationSchemaType): DraftCompensation {
  return {
    hourly_rate: comp.hourly_rate.toString(),
  }
}

function schemaCompensationFromDraftType(comp: DraftCompensation): CompensationSchemaType {
  return {
    hourly_rate: parseFloat(comp.hourly_rate),
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
  const [compensation, setCompensation] = useState<DraftCompensation | "loading" | "error">("loading");
  const [updateCompensationLoading, setUpdateCompensationLoading] = useState(false);

  const fetchCompensation = useCallback(async () => {
    const res = await api.compensation.get();
    if (isLeft(res)) {
      console.error(`failed to get compensation: ${res.left}`);
      setCompensation("error");
      return;
    }

    setCompensation(draftCompensationFromSchemaType(res.right));
  }, [setCompensation])

  useEffect(() => {
    fetchCompensation()
  }, [fetchCompensation]);

  const updateCompensation = useCallback(function<K extends keyof DraftCompensation, V extends DraftCompensation[K]>(key: K, value: V) {
    setCompensation((comp) => {
      if (comp === "loading" || comp === "error") {
        return comp;
      }

      return {
        ...comp,
        [key]: value,
      };
    });
  }, [])

  const sendCompensationUpdate = useCallback(async () => {
    if (compensation === "loading" || compensation === "error") {
      return;
    }

    setUpdateCompensationLoading(true);

    const comp = schemaCompensationFromDraftType(compensation);
    const res = await api.compensation.set({
      hourlyRate: comp.hourly_rate,
    });

    setUpdateCompensationLoading(false);

    if (isLeft(res)) {
      toast({
        kind: "error",
        message: "Failed to update compensation settings",
      });
      console.error(`Failed to update compensation settings: ${res.left}`);
      return;
    }

    setCompensation(draftCompensationFromSchemaType(res.right));

    toast({
      kind: "success",
      message: "Updated compensation settings",
    })
  }, [compensation, setCompensation, setUpdateCompensationLoading]);

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
        Compensation
      </Typography>
      {compensation === "loading" ? (
        <>
          <CircularProgress />
          <Typography>Loading compensation settings</Typography>
        </>
      ) : compensation === "error" ? (
        <>
          <WarningIcon />
          <Typography>Failed to load compensation settings</Typography>
        </>
      ) : (
        <>
          <TextField
            label="Hourly Rate"
            value={compensation.hourly_rate}
            onChange={(e) => updateCompensation("hourly_rate", e.target.value)}
          />

          <Button
            disabled={updateCompensationLoading}
            type="submit"
            onClick={sendCompensationUpdate}
          >
            {updateCompensationLoading && (
              <CircularProgress />
            )}
            Save
          </Button>
        </>
      )}
      
    </Drawer>
  ) 
}