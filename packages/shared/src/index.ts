export type CustomerProfile = {
  id: number;
  lineDisplayName: string;
  pictureUrl?: string;
  hasConnectedFacebook: boolean;
  activePageName?: string;
};

export type FacebookPageSummary = {
  pageId: string;
  pageName: string;
  category: string;
  isActive: boolean;
};

export type TopPost = {
  postId: string;
  reason: string;
  recommendation: string;
};

export type ImportantComment = {
  commentId: string;
  message: string;
  reason: string;
  suggestedReply: string;
};

export type PageMetrics = {
  clicks: number;
  engagements: number;
  impressions: number;
  reach: number;
};

export type PostingDayInsight = {
  day: string;
  postCount: number;
  averageEngagement: number;
};

export type PostingTimeInsight = {
  basedOnPosts: number;
  bestDay: string;
  bestTime: string;
  days: PostingDayInsight[];
};

export type WeeklyReport = {
  startDate: string;
  endDate: string;
  daysWithData: number;
  metrics: PageMetrics;
};

export type AnalysisReport = {
  id: string;
  pageName: string;
  summary: string;
  healthScore: number;
  topPosts: TopPost[];
  importantComments: ImportantComment[];
  contentRecommendations: string[];
  aiContentRecommendation?: string;
  bestPostingTimes: string[];
  postingTimeInsight?: PostingTimeInsight;
  postingTimeRecommendation?: string;
  lineSummaryMessage: string;
  createdAt: string;
  metrics?: PageMetrics;
};

export type ApiError = {
  error: string;
};
