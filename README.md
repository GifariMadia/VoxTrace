# VoxTrace

VoxTrace mengubah rekaman speech menjadi transkrip terstruktur. Arsitekturnya terdiri dari web workspace, Go API, PostgreSQL, dan Python ML worker.

## Status

MVP tersedia dengan dua mode pemrosesan:

- `mock`: menguji alur upload sampai transcript tanpa GPU.
- `whisper`: menjalankan Faster-Whisper Medium pada GPU.

## Arsitektur

```text
Web workspace (Next.js-compatible vinext)
             |
             v
        Go REST API ------ PostgreSQL
             |
             v
    Python transcription worker
      Faster-Whisper Medium
```

Penyimpanan file MVP menggunakan filesystem lokal. PostgreSQL menyimpan recording metadata, lifecycle job, transcript, segment, dan processing metadata.

## Menjalankan MVP

Persyaratan: Docker Desktop dan Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Jika registry npm utama terputus dari jaringan Docker, atur mirror hanya pada `.env` lokal:

```dotenv
NPM_REGISTRY=https://registry.npmmirror.com
PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
```

Buka `http://localhost:3000`. Mode bawaan adalah `whisper` untuk transkripsi GPU. API tersedia di `http://localhost:8080/api/health`.

Periksa service yang sedang berjalan dengan `docker compose ps`. Instalasi lokal telah diverifikasi pada RTX 3050 Laptop GPU 4 GB dengan CUDA 12.8.

### Development per service

Web:

```bash
npm install
npm run dev
```

API:

```bash
cd apps/api
go run .
```

Worker:

```bash
python -m pip install -r workers/transcription/requirements.txt
python workers/transcription/main.py
```

## Profil Faster-Whisper

Worker GPU menggunakan CUDA 12.8 dan meminta satu GPU NVIDIA melalui Docker Compose. Untuk RTX 3050 4 GB gunakan profil bawaan berikut:

```dotenv
PIPELINE_BACKEND=whisper
WHISPER_MODEL=medium
DEVICE=cuda
COMPUTE_TYPE=int8
BATCH_SIZE=1
```

Model `medium` dengan compute type `int8` adalah profil aman untuk VRAM 4 GB. Pipeline ini hanya melakukan speech-to-text dan word timestamps; alignment WhisperX serta speaker diarization tidak dijalankan.

Pastikan Docker Desktop berjalan, lalu validasi akses GPU:

```bash
docker run --rm --gpus all nvidia/cuda:12.8.1-base-ubuntu22.04 nvidia-smi
docker compose up --build
```

Kontrak worker tetap sama untuk mode mock dan Whisper. Dengan begitu UI, API, database, retry, serta export dapat dikembangkan dan diuji tanpa GPU.

## Endpoint

- `POST /api/jobs` — multipart field `audio`
- `GET /api/jobs`
- `GET /api/jobs/:id`
- `GET /api/jobs/:id/transcript`
- `GET /api/jobs/:id/export`
- `POST /api/jobs/:id/retry`
- `DELETE /api/jobs/:id`

Format yang diterima: MP3, WAV, M4A, FLAC, OGG, dan WebM hingga 2 GB.

## Validasi

Jalankan pemeriksaan berikut sebelum membuka pull request:

```bash
npm run lint
npm run build
cd apps/api && go test ./... && go vet ./...
python -m py_compile workers/transcription/main.py
docker compose config --quiet
```

## Struktur

- `app/` — workspace web
- `apps/api/` — Go orchestration API
- `workers/transcription/` — ML worker
- `migrations/` — schema PostgreSQL
- `storage/` — local MVP blob storage

## Batas MVP

Belum mencakup authentication, object storage production, distributed queue, transcript editor, maupun LLM summary. Untuk deployment publik, tambahkan authentication, malware scanning, rate limiting, retention policy, TLS, dan secret management.

## Dokumentasi perubahan

Setiap fitur atau perubahan perilaku wajib memperbarui dokumentasi yang relevan dan menambahkan entri ke [CHANGELOG.md](CHANGELOG.md). Aturan lengkap tersedia di [CONTRIBUTING.md](CONTRIBUTING.md).
