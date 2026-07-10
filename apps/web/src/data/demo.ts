import type { AnalysisReport, CustomerProfile, FacebookPageSummary } from "@linora/shared";

export const demoCustomer: CustomerProfile = {
  id: 1,
  lineDisplayName: "คุณ Linora",
  hasConnectedFacebook: false,
};

export const demoPages: FacebookPageSummary[] = [
  {
    pageId: "page-1",
    pageName: "Linora Cafe",
    category: "Local business",
    isActive: true,
  },
  {
    pageId: "page-2",
    pageName: "Linora Studio",
    category: "Creator",
    isActive: false,
  },
];

export const demoReport: AnalysisReport = {
  id: "demo-report",
  pageName: "ยังไม่ได้เชื่อมต่อเพจ",
  summary: "เชื่อมต่อเพจ Facebook หรือกรอกข้อมูลโพสต์เพื่อรับคำแนะนำเบื้องต้น",
  healthScore: 72,
  topPosts: [
    {
      postId: "demo-post",
      reason: "โพสต์ที่มีคำถามจากลูกค้าชัดเจนมักสร้างโอกาสในการปิดการขาย",
      recommendation: "เพิ่ม CTA ให้ลูกค้ากดทัก LINE ได้ทันที",
    },
  ],
  importantComments: [
    {
      commentId: "demo-comment",
      message: "ราคาเท่าไหร่ จองได้ไหม",
      reason: "เป็นสัญญาณซื้อ ควรตอบเร็ว",
      suggestedReply: "ขอบคุณที่สนใจค่ะ แจ้งรุ่นหรือวันที่ต้องการได้เลยนะคะ",
    },
  ],
  contentRecommendations: [
    "โพสต์รีวิวลูกค้าจริง",
    "เพิ่มโปรโมชันพร้อม call-to-action",
    "นำคำถามยอดฮิตมาทำ FAQ",
  ],
  bestPostingTimes: ["18:00 - 20:00", "11:00 - 13:00"],
  lineSummaryMessage: "สรุปเพจวันนี้จาก Linora\n\nคะแนนภาพรวม: 72/100\n\nคำแนะนำ: เพิ่ม CTA ให้ชัดขึ้น",
  createdAt: new Date().toISOString(),
};
