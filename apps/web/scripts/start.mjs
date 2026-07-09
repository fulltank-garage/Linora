import { spawn } from "node:child_process";

const args = ["preview", "--host", "0.0.0.0", "--port", "5173"];

const child = spawn("vite", args, {
  stdio: "inherit",
  shell: process.platform === "win32",
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }

  process.exit(code ?? 0);
});
