export interface EvalRun {
  id: number;
  category: string;
  total_tests: number;
  passed: number;
  failed: number;
  accuracy: number;
  metrics: string;
  created_at: string;
}

export interface EvalMetric {
  category: string;
  total_searches: number;
  avg_relevance: number;
  cart_adds: number;
  orders_created: number;
}

export interface MetricsResult {
  metrics: EvalMetric[];
}
