import { Alert, Snackbar } from "@mui/material";
import { ReactNode, createContext, useCallback, useState } from "react";
import { v4 as uuidv4 } from "uuid";

export type ToastOpts =
  | {
      readonly kind: "error";
      readonly message: string;
    }
  | {
      readonly kind: "success";
      readonly message: string;
    };
export const ToastCtx = createContext(() => {});
type ToastItem = ToastOpts & {
  readonly id: string;
};

export const ToastProvider = ({
  children,
}: {
  readonly children: ReactNode;
}) => {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const addToast = useCallback(
    (opts: ToastOpts) => {
      setToasts((toasts) => [
        ...toasts,
        {
          ...opts,
          id: uuidv4(),
        },
      ]);
    },
    [setToasts],
  );

  const makeOnClose = (toast: ToastItem) => {
    return () => {
      setToasts((toasts) => [
        ...toasts.filter((innerToast) => innerToast.id !== toast.id),
      ]);
    };
  };

  return (
    <>
      <ToastCtx.Provider value={addToast}>
        {children}
        {toasts.map((toast) => (
          <Snackbar
            key={toast.id}
            open={true}
            autoHideDuration={6000}
            onClose={makeOnClose(toast)}
          >
            <Alert onClose={makeOnClose(toast)} severity={toast.kind}>
              {toast.message}
            </Alert>
          </Snackbar>
        ))}
      </ToastCtx.Provider>
    </>
  );
};
