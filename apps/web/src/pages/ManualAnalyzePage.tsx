import { FormEvent, useState } from "react";
import { Alert, Button, Stack, TextField, Typography } from "@mui/material";
import { AutoAwesome } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";
import type { AnalysisReport, ManualAnalysisRequest } from "@linora/shared";
import { runManualAnalysis } from "../api/client";

type ManualAnalyzePageProps = {
  onReport: (report: AnalysisReport) => void;
};

const initialForm: ManualAnalysisRequest = {
  pageName: "",
  postContent: "",
  likes: 0,
  comments: 0,
  shares: 0,
  importantComments: "",
  extraNotes: "",
};

export function ManualAnalyzePage({ onReport }: ManualAnalyzePageProps) {
  const [form, setForm] = useState(initialForm);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);
    try {
      const report = await runManualAnalysis(form);
      onReport(report);
      navigate("/");
    } catch {
      setError("วิเคราะห์ไม่สำเร็จ กรุณาตรวจข้อมูลและลองใหม่อีกครั้ง");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Stack component="form" onSubmit={handleSubmit} spacing={2}>
      <Typography variant="h1">วิเคราะห์แบบ Manual</Typography>
      <Typography color="text.secondary">
        ใช้ได้ก่อน Meta App Review โดยกรอกข้อมูลโพสต์และคอมเมนต์เอง
      </Typography>
      {error ? <Alert severity="error">{error}</Alert> : null}
      <TextField label="ชื่อเพจ" required value={form.pageName} onChange={(event) => setForm({ ...form, pageName: event.target.value })} />
      <TextField label="เนื้อหาโพสต์" minRows={4} multiline required value={form.postContent} onChange={(event) => setForm({ ...form, postContent: event.target.value })} />
      <Stack direction="row" spacing={1}>
        <TextField label="Likes" type="number" value={form.likes} onChange={(event) => setForm({ ...form, likes: Number(event.target.value) })} />
        <TextField label="Comments" type="number" value={form.comments} onChange={(event) => setForm({ ...form, comments: Number(event.target.value) })} />
        <TextField label="Shares" type="number" value={form.shares} onChange={(event) => setForm({ ...form, shares: Number(event.target.value) })} />
      </Stack>
      <TextField label="คอมเมนต์สำคัญ" minRows={3} multiline value={form.importantComments} onChange={(event) => setForm({ ...form, importantComments: event.target.value })} />
      <TextField label="หมายเหตุเพิ่มเติม" minRows={2} multiline value={form.extraNotes} onChange={(event) => setForm({ ...form, extraNotes: event.target.value })} />
      <Button disabled={isSubmitting} fullWidth startIcon={<AutoAwesome />} type="submit" variant="contained">
        {isSubmitting ? "กำลังวิเคราะห์..." : "เริ่มวิเคราะห์"}
      </Button>
    </Stack>
  );
}
