# Kontribusi VoxTrace

## Aturan perubahan

Setiap fitur, perbaikan bug, perubahan konfigurasi, atau perubahan API harus:

1. Memiliki commit dengan pesan yang menjelaskan tujuan perubahan.
2. Menyertakan pengujian atau langkah validasi yang sesuai.
3. Memperbarui `README.md` bila cara instalasi atau penggunaan berubah.
4. Memperbarui dokumentasi endpoint bila kontrak API berubah.
5. Menambahkan entri di bagian `Unreleased` pada `CHANGELOG.md`.
6. Tidak menyimpan token, password, audio observasi, atau transcript sensitif ke Git.

## Format commit

Gunakan pesan singkat dan berbentuk imperatif:

```text
Add transcript search
Fix stale job recovery
Document WhisperX GPU setup
```

Satu commit sebaiknya mencakup satu perubahan logis. Hindari pesan seperti `update`, `fix`, atau `changes` tanpa konteks.

## Checklist pull request

- [ ] Fitur atau bug dijelaskan dengan jelas.
- [ ] Dokumentasi dan changelog diperbarui.
- [ ] Web lint dan build berhasil.
- [ ] Go test dan vet berhasil.
- [ ] Worker Python dapat dikompilasi.
- [ ] Docker Compose valid.
- [ ] Tidak ada secret atau data observasi sensitif.

## Branch

Gunakan branch pendek berdasarkan jenis pekerjaan:

- `feature/nama-fitur`
- `fix/nama-bug`
- `docs/topik`
- `chore/topik`

Branch `main` harus selalu dalam keadaan dapat dibangun.
