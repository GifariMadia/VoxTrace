# VoxTrace Extraction Reports

Direktori ini berisi paket dokumen hasil ekstraksi enam rekaman yang selesai diproses pada 14 Agustus 2026.

## Isi paket

- `00 - Ikhtisar Hasil Ekstraksi VoxTrace.docx`: ringkasan lintas rekaman, tema bersama, katalog, dan prioritas verifikasi.
- Enam `Laporan Ekstraksi.docx`: satu laporan per rekaman yang memuat ringkasan eksekutif, pokok bahasan, tindak lanjut, peta waktu, dan transkrip tertata.

## Batas penggunaan

Dokumen mempertahankan isi hasil speech-to-text dan merapikannya untuk navigasi. Nama, istilah teknis, angka, set point, serta percakapan yang tumpang tindih harus diverifikasi terhadap audio sumber sebelum dipakai untuk audit, perubahan SOP, atau keputusan formal.

## Reproduksi

1. Jalankan stack VoxTrace dan pastikan API tersedia di `http://localhost:8080`.
2. Jalankan `tools/export_extractions.py` untuk mengambil seluruh job berstatus selesai.
3. Jalankan `tools/build_extraction_documents.py` untuk membuat ulang paket DOCX.

