interface ArticlePageProps {
  params: Promise<{ projectId: string }>;
}

export default async function ArticleProjectPage({ params }: ArticlePageProps) {
  const { projectId } = await params;
  return (
    <div>
      <h1>Article Content Planning</h1>
      <p>Project ID: {projectId}</p>
    </div>
  );
}
