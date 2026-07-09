import { Box, Button } from "@mui/material";
import { Link } from "react-router-dom";

export function ComplianceLinks() {
  return (
    <Box
      sx={{
        display: "flex",
        flexWrap: "wrap",
        gap: 0.75,
        justifyContent: "center",
      }}
    >
      <Button component={Link} size="small" to="/privacy" variant="text">
        ความเป็นส่วนตัว
      </Button>
      <Button component={Link} size="small" to="/terms" variant="text">
        เงื่อนไข
      </Button>
      <Button component={Link} size="small" to="/data-deletion" variant="text">
        ลบข้อมูล
      </Button>
    </Box>
  );
}
