CREATE TABLE urls(
  id BIGINT PRIMARY KEY GENERATED ALWAYS as IDENTITY,
  url TEXT NOT NULL,
  short TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE url_views(
  url_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX url_view_time on url_views (url_id, created_at);