# VoxTrace Operational Knowledge

Direktori ini berisi tujuh dokumen matang yang disusun dari enam hasil transkripsi VoxTrace serta catatan lapangan pendukung. Paket ini bukan lagi kumpulan ekstraksi atau transkrip.

## Isi paket

- `00 - Ikhtisar Pengetahuan Operasional VoxTrace.docx`: peta portofolio, aturan korelasi, dan antarmuka antardomain.
- `13-08-2026 10.36 - Panduan Alur Produksi dan Maintenance.docx`.
- `12-08-2026 09.05 - Panduan Observasi Mesin dan Suku Cadang.docx`.
- `12-08-2026 13.27 - Panduan Building Maintenance dan Project.docx`.
- `12-08-2026 14.28 - Panduan IPAL dan Limbah B3.docx`.
- `12-08-2026 11.22 - Panduan Kelistrikan dan Proteksi Kebakaran.docx`.
- `13-08-2026 13.37 - Kerangka Evaluasi Proses.docx`.

## Prinsip penyusunan

Setiap dokumen berdiri sendiri dan mempertahankan ranah rekaman sumbernya. Informasi dari catatan scan dan foto hanya digunakan untuk memperjelas konteks yang relevan. Informasi lintas domain ditempatkan sebagai antarmuka, bukan sebagai intervensi terhadap mandat, angka, atau keputusan dokumen lain.

Angka, nama, istilah, dan set point tulisan tangan yang belum dapat dipastikan ditempatkan sebagai catatan verifikasi. SOP, izin, drawing, hasil laboratorium, dan data sistem tetap menjadi sumber formal tertinggi.

## Reproduksi

1. Perbarui sumber transkripsi di `build/extractions/source` bila diperlukan.
2. Tinjau dan perbarui materi editorial pada `tools/build_mature_documents.py`.
3. Jalankan `tools/build_mature_documents.py` untuk membuat ulang paket DOCX.
4. Render seluruh dokumen ke PDF/PNG dan periksa setiap halaman sebelum diterbitkan.
