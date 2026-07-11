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

export type ManualAnalysisRequest = {
  pageName: string;
  postContent: string;
  likes: number;
  comments: number;
  shares: number;
  importantComments: string;
  extraNotes: string;
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

export type AnalysisReport = {
  id: string;
  pageName: string;
  summary: string;
  healthScore: number;
  topPosts: TopPost[];
  importantComments: ImportantComment[];
  contentRecommendations: string[];
  bestPostingTimes: string[];
  lineSummaryMessage: string;
  createdAt: string;
  metrics?: PageMetrics;
};

export type ManualAnalysisResponse = {
  report: AnalysisReport;
};

export type ApiError = {
  error: string;
};
