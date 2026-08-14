# VoxTrace

VoxTrace mengubah rekaman observasi menjadi transkrip terstruktur dengan word alignment dan speaker diarization. Arsitekturnya terdiri dari web workspace, Go API, PostgreSQL, dan Python ML worker.

## Menjalankan MVP

Persyaratan: Docker Desktop dan Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Buka `http://localhost:3000`. Mode bawaan adalah `mock`, sehingga seluruh flow upload sampai transcript dapat diuji tanpa GPU. API tersedia di `http://localhost:8080/api/health`.

## Mengaktifkan WhisperX

Worker produksi membutuhkan image CUDA/PyTorch/WhisperX yang sesuai dengan GPU host, NVIDIA Container Toolkit, serta token Hugging Face yang telah menerima syarat model PyAnnote. Tambahkan dependency ML ke image worker, lalu isi:

```dotenv
PIPELINE_BACKEND=whisperx
WHISPER_MODEL=large-v3
HF_TOKEN=hf_...
```

Kontrak worker sengaja tetap sama untuk mode mock dan WhisperX. Dengan begitu UI, API, database, retry, serta export dapat dikembangkan dan diuji tanpa menunggu GPU.

## Endpoint

- `POST /api/jobs` — multipart field `audio`
- `GET /api/jobs`
- `GET /api/jobs/:id`
- `GET /api/jobs/:id/transcript`
- `GET /api/jobs/:id/export`
- `POST /api/jobs/:id/retry`
- `DELETE /api/jobs/:id`

Format yang diterima: MP3, WAV, M4A, FLAC, OGG, dan WebM hingga 2 GB.

## Struktur

- `app/` — workspace web
- `apps/api/` — Go orchestration API
- `workers/transcription/` — ML worker
- `migrations/` — schema PostgreSQL
- `storage/` — local MVP blob storage

## Batas MVP

Belum mencakup authentication, object storage production, distributed queue, transcript editor, maupun LLM summary. Untuk deployment publik, tambahkan authentication, malware scanning, rate limiting, retention policy, TLS, dan secret management.
