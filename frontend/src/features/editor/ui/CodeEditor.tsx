"use client";

import CodeMirror from "@uiw/react-codemirror";
import { python } from "@codemirror/lang-python";
import { oneDark } from "@codemirror/theme-one-dark";

type Props = {
  value: string;
  onChange: (value: string) => void;
};

export default function CodeEditor({ value, onChange }: Props) {
  return (
    <CodeMirror
      value={value}
      height="100%"
      theme={oneDark}
      extensions={[python()]}
      onChange={onChange}
      style={{ height: "100%", fontSize: "14px" }}
      basicSetup={{
        lineNumbers: true,
        highlightActiveLineGutter: true,
        foldGutter: false,
        autocompletion: true,
      }}
    />
  );
}
