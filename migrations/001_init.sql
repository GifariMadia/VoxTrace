CREATE TABLE IF NOT EXISTS recordings (
  id UUID PRIMARY KEY, original_filename TEXT NOT NULL, stored_path TEXT NOT NULL UNIQUE,
  mime_type TEXT NOT NULL, size_bytes BIGINT NOT NULL CHECK(size_bytes >= 0), duration_seconds DOUBLE PRECISION,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY, recording_id UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK(status IN ('queued','processing','completed','failed','cancelled')),
  stage TEXT NOT NULL, progress INTEGER NOT NULL DEFAULT 0 CHECK(progress BETWEEN 0 AND 100), model TEXT NOT NULL,
  error_message TEXT, attempt_count INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS jobs_queue_idx ON jobs(status, created_at);
CREATE TABLE IF NOT EXISTS transcripts (
  id UUID PRIMARY KEY, job_id UUID NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
  language TEXT NOT NULL, duration_seconds DOUBLE PRECISION NOT NULL, raw_text TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS segments (
  id UUID PRIMARY KEY, transcript_id UUID NOT NULL REFERENCES transcripts(id) ON DELETE CASCADE,
  sequence INTEGER NOT NULL, speaker TEXT NOT NULL, start_time DOUBLE PRECISION NOT NULL,
  end_time DOUBLE PRECISION NOT NULL, text TEXT NOT NULL, words JSONB NOT NULL DEFAULT '[]',
  UNIQUE(transcript_id, sequence)
);
CREATE TABLE IF NOT EXISTS processing_metadata (
  job_id UUID PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE, metadata JSONB NOT NULL DEFAULT '{}'
);
