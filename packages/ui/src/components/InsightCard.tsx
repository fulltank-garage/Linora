import type { ReactNode } from "react";
import { Card, CardContent, Stack, Typography } from "@mui/material";

type InsightCardProps = {
  title: string;
  value?: ReactNode;
  helper?: string;
  action?: ReactNode;
};

export function InsightCard({ title, value, helper, action }: InsightCardProps) {
  return (
    <Card>
      <CardContent>
        <Stack spacing={1.25}>
          <Typography color="text.secondary" sx={{ fontSize: 13, fontWeight: 700 }}>
            {title}
          </Typography>
          {value ? (
            <Typography component="div" variant="h3">
              {value}
            </Typography>
          ) : null}
          {helper ? (
            <Typography color="text.secondary" sx={{ fontSize: 14, lineHeight: 1.55 }}>
              {helper}
            </Typography>
          ) : null}
          {action}
        </Stack>
      </CardContent>
    </Card>
  );
}
