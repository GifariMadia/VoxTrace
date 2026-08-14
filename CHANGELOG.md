# Changelog

Semua perubahan penting VoxTrace didokumentasikan di file ini. Format mengikuti prinsip [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Added

- Pembatalan job antrean maupun job aktif melalui dashboard dan endpoint API.
- Cache model Hugging Face worker disimpan di `storage/models` agar model tidak diunduh ulang saat container dibuat kembali.
- Pedoman kontribusi dan kebijakan dokumentasi setiap perubahan.
- Dokumentasi arsitektur, development workflow, dan validasi lokal.
- Worker CUDA 12.8 dan reservasi GPU NVIDIA melalui Docker Compose.

### Changed

- Dashboard kini dimulai dari workspace kosong tanpa recording dan transkrip demo; kegagalan upload ditampilkan sebagai gagal, bukan disimulasikan sebagai proses aktif.
- Profil WhisperX dioptimalkan untuk RTX 3050 4 GB dengan model `medium`, compute `int8`, batch size 1, dan pelepasan VRAM antar-tahap.
- Docker build context mengecualikan dependency, cache, output build, dan penyimpanan lokal agar build web lebih cepat dan kecil.
- Instalasi dependency pada image web dibuat deterministik tanpa lifecycle script, audit, dan funding request.
- Stage build web menggunakan npm 11 untuk menghindari kegagalan exit handler pada npm 10 bawaan Node 22 Alpine.
- Registry npm untuk Docker build dapat dikonfigurasi melalui `NPM_REGISTRY` ketika registry utama tidak dapat dijangkau dari container.
- Worker menggunakan repository Ubuntu HTTPS agar instalasi dependency berhasil pada jaringan Docker yang memblokir HTTP.
- Dependency Python worker memakai cache BuildKit serta retry/timeout panjang agar unduhan paket ML besar dapat dilanjutkan dengan aman.
- Registry Python untuk build worker dapat dikonfigurasi melalui `PIP_INDEX_URL`.
- Versi WhisperX dan Transformers dikunci untuk mencegah dependency backtracking dan menjaga build reproducibel.

### Fixed

- Dashboard melakukan polling API dan mengambil transcript selesai, sehingga status tidak lagi tertahan pada `queued` ketika worker sudah memproses job.
- Response job dan segment API menggunakan nama field JSON camelCase yang konsisten dengan frontend.
- Hydration mismatch pada label waktu recording demo akibat penggunaan `Date.now()` saat server render.

### Verified

- Seluruh stack Docker berhasil dibangun dan dijalankan: web, Go API, PostgreSQL, dan worker.
- Worker mengenali NVIDIA GeForce RTX 3050 Laptop GPU 4 GB melalui PyTorch CUDA 12.8 dan dapat mengimpor WhisperX.
- Endpoint web merespons HTTP 200 dan health check API mengembalikan status `ok`.

## 0.1.0 - 2026-08-14

### Added

- Dashboard upload dan riwayat processing job.
- Transcript viewer dengan speaker segment, pencarian, timeline, dan export TXT.
- Go REST API untuk upload, lifecycle job, retry, delete, transcript, dan export.
- PostgreSQL schema untuk recordings, jobs, transcripts, segments, dan metadata.
- Python worker dengan backend mock dan integrasi WhisperX opsional.
- Docker Compose untuk web, API, worker, dan PostgreSQL.
- Social preview khusus VoxTrace.
