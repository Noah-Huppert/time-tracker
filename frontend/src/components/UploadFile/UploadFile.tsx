import { Box, Button } from "@mui/material";
import { ChangeEvent, useCallback, useState } from "react";

export const UploadFile = ({
  onUpload,
  uploadLabel,
}: {
  readonly onUpload: (fileList: FileList) => void
  readonly uploadLabel: string
}) => {
  const [files, setFiles] = useState<FileList | null>(null);

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
  }, [files, onUpload])

  return (
    <Box>
      <input type="file" onChange={onChange} />

      <Button
        disabled={files === null}
        onClick={onUploadButtonClick}
      >
        {uploadLabel}
      </Button>
    </Box>
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