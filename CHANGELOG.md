# Changelog

Semua perubahan penting VoxTrace didokumentasikan di file ini. Format mengikuti prinsip [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Added

- Pedoman kontribusi dan kebijakan dokumentasi setiap perubahan.
- Dokumentasi arsitektur, development workflow, dan validasi lokal.

## 0.1.0 - 2026-08-14

### Added

- Dashboard upload dan riwayat processing job.
- Transcript viewer dengan speaker segment, pencarian, timeline, dan export TXT.
- Go REST API untuk upload, lifecycle job, retry, delete, transcript, dan export.
- PostgreSQL schema untuk recordings, jobs, transcripts, segments, dan metadata.
- Python worker dengan backend mock dan integrasi WhisperX opsional.
- Docker Compose untuk web, API, worker, dan PostgreSQL.
- Social preview khusus VoxTrace.
