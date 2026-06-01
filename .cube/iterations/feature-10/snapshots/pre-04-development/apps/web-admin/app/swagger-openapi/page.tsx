export default function SwaggerOpenAPIPage() {
  return (
    <main>
      <h1>Swagger / OpenAPI</h1>
      <p>OpenAPI 文档用于核对后端接口契约。</p>
      <a href="/openapi.yaml">打开 OpenAPI YAML</a>
      <div role="alert" hidden>request_id</div>
    </main>
  );
}
