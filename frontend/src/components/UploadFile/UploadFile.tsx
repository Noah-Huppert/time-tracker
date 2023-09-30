import { Box, Button, Card, CardContent, Paper, Typography } from "@mui/material";
import { ChangeEvent, useCallback, useRef, useState } from "react";

export const UploadFile = ({
  onUpload,
  uploadLabel,
}: {
  readonly onUpload: (fileList: FileList) => void
  readonly uploadLabel: string
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [files, setFiles] = useState<FileList | null>(null);

  const [showingInput, setShowingInput] = useState(false);

  const onChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files === null) {
      return;
    }

    setFiles(e.target.files)
  }, [setFiles]);

  const onUploadButtonClick = useCallback(() => {
    if (files === null) {
      return;
    }

    onUpload(files);
    onClearButtonClick();
    setShowingInput(false);
  }, [files, onUpload])

  const onClearButtonClick = useCallback(() => {
    setShowingInput(false);

    if (fileInputRef.current === null) {
      return;
    }

    fileInputRef.current.value = "";
    setFiles(null);
  }, [setFiles, fileInputRef]);

  if (showingInput === false) {
    return (
      <Box>
        <Button
          variant="contained"
          onClick={() => setShowingInput(true)}
        >
          {uploadLabel}
        </Button>
      </Box>
    )
  }

  return (
    <Card>
      <CardContent>
        <Typography
          variant="body1"
          sx={{
            marginBottom: "1rem",
          }}
        >
          {uploadLabel}
        </Typography>

        <input
          ref={fileInputRef}
          type="file"
          onChange={onChange}
          multiple
          accept=".csv"
        />

        <Box
          sx={{
            display: "flex",
            flexDirection: "row",
            marginTop: "1rem",
          }}
        >
          <Button
            variant="contained"
            onClick={onClearButtonClick}
          >
            Cancel
          </Button>

          <Button
            disabled={files === null}
            variant="contained"
            onClick={onUploadButtonClick}
            sx={{
              marginLeft: "1rem",
            }}
          >
            Upload
          </Button>
        </Box>
      </CardContent>
    </Card>
  );
};

export type ReadFile = {
  readonly name: string
  readonly content: string
}

export async function ReadFiles(fileList: FileList): Promise<ReadFile[]> {
  // Ignore empty file list
  if (fileList.length === 0) {
    return [];
  }

  // Read all files
  const allFiles = []
  for (const file of fileList) {
    allFiles.push(file);
  }

  return await Promise.all(allFiles.map((file) => new Promise<ReadFile>((resolve, reject) => {
    const reader = new FileReader();
    reader.onabort = () => {
      reject("aborted");
    };
    reader.onerror = (e) => {
      reject(`error: ${e}`);
    };
    reader.onload = (e) => {
      if (e.target === null) {
        reject("loaded with null target");
        return;
      }

      if (typeof e.target.result !== "string") {
        reject(`loaded with non-string data type of: ${typeof e.target.result}`);
        return;
      }

      resolve({
        name: file.name,
        content: e.target.result,
      });
    };

    reader.readAsText(file);
  })));
}