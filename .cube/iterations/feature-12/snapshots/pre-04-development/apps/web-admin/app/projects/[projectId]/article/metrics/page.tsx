interface ArticleMetricsPageProps {
  params: Promise<{ projectId: string }>;
}

export default async function ArticleMetricsPage({ params }: ArticleMetricsPageProps) {
  const { projectId } = await params;
  return (
    <div>
      <h1>Article Metrics Configuration</h1>
      <p>Project ID: {projectId}</p>
    </div>
  );
}
