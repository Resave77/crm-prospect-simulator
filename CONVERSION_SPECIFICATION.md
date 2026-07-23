# Spesifikasi Konversi Prospect menjadi Customer Existing

Status: `DRAFT UNTUK PERSETUJUAN MENTOR — DOKUMENTASI SAJA`

Dokumen ini mendefinisikan ruang lingkup simulasi bisnis poin 10 sampai 14. Dokumen ini tidak mengubah arsitektur yang telah disetujui dan belum menjadi izin untuk membuat layar, API, migrasi, model database, atau implementasi CRM.

## 1. Tujuan dan Prinsip

- Memisahkan dengan jelas data **Customer Prospect** dari data **Customer Existing**.
- Menentukan data yang dapat dibawa dari Prospect, data yang harus dilengkapi Administrator, data yang berasal dari master data, dan metadata yang dibuat sistem.
- Menjaga sumber Prospect dan snapshot Google Places untuk audit. Istilah "pindah modul" berarti record tidak lagi tampil sebagai Prospect aktif dan mulai tampil sebagai Customer Existing; record Prospect sumber tidak dihapus.
- Google Places hanya menjadi sumber identitas tempat, alamat, koordinat, kategori, dan informasi publik yang memang tersedia. Google Places bukan sumber data pajak, bank, billing, ERP, atau struktur perusahaan.
- Nilai yang belum dapat ditentukan dari kebutuhan mentor diberi label **Needs Business Confirmation**.

## 2. Prospect-to-Customer Conversion Flow

### 2.1 Alur poin 10–14

| Poin | Aktor | Aksi | Prasyarat | Hasil |
|---:|---|---|---|---|
| 10 | Sales Executive | Memperbarui pipeline dan menetapkan hasil Prospect | Prospect dimiliki Sales Executive aktif dan telah berada pada tahap `FOLLOW_UP` | Prospect menjadi `LOST` atau `WON` |
| 10A | Sales Executive | Memilih hasil tidak berhasil | Alasan gagal tersedia | Status menjadi `LOST`; proses berhenti dan konversi dilarang |
| 10B | Sales Executive | Memilih hasil berhasil | Catatan kemenangan tersedia | Status menjadi `WON`; Prospect masuk antrean review Administrator |
| 11 | Administrator | Membuka dan meninjau Prospect `WON` | Status `WON`, belum pernah dikonversi, dan tidak memiliki link Customer Existing | Administrator dapat melanjutkan atau menunda konversi |
| 12 | Administrator | Memulai konversi | Review kelayakan selesai | Form Customer Existing dibuka dengan data yang dapat diprefill |
| 13 | Administrator | Melengkapi dan memvalidasi data Customer Existing | Data sumber telah ditampilkan; aturan duplikasi dijalankan | Form siap dikonversi apabila semua field wajib valid |
| 14 | System | Menjalankan konversi atomik | Prospect masih `WON`, belum dikonversi, Customer Code serta Parent Code berhasil dihasilkan/diambil dan unik, dan form valid | Parent Company dibuat/ditautkan, Customer Site dibuat, Prospect menjadi `CONVERTED`, dan Customer tampil di Customer Existing serta `My Customer` |

### 2.2 Transisi status

```text
FOLLOW_UP -> LOST
FOLLOW_UP -> WON -> CONVERTED
```

Aturan transisi:

- `LOST` bersifat terminal dan tidak dapat dikonversi.
- `WON` belum berarti Customer Existing. Status ini hanya membuat Prospect layak direview Administrator.
- `CONVERTED` hanya ditetapkan setelah seluruh operasi konversi berhasil.
- Kegagalan pada pembuatan/link Parent Company, pembuatan Customer Site, penutupan assignment, atau pencatatan metadata/history konversi minimum membatalkan seluruh konversi.
- Setelah `CONVERTED`, Prospect dikeluarkan dari antrean Prospect aktif, tetap tersedia sebagai sumber audit, dan tidak dapat dikonversi kedua kali.

### 2.3 Tanggung jawab aktor

**Sales Executive** bertanggung jawab atas hasil kunjungan, catatan follow-up, dan keputusan `WON` atau `LOST`. Sales Executive tidak membuat Customer Existing.

**Administrator** bertanggung jawab meninjau Prospect `WON`, memilih atau membuat Parent Company, memeriksa data hasil prefill, melengkapi field manual, memilih master data, dan menyetujui konversi.

**System** bertanggung jawab memeriksa duplikasi dan status, mengunci data sumber saat konversi, menghasilkan kode dan metadata internal yang diperlukan, membuat/link entitas secara atomik, memindahkan record ke antrean modul yang tepat, dan menyimpan riwayat transisi status.

## 3. Klasifikasi Data Sumber

### 3.1 Data Google Places yang dapat dipakai

Data berikut hanya digunakan jika tersedia pada response API dan telah disimpan sebagai snapshot Prospect:

- Google Place ID: identitas provider dan kontrol duplikasi; bukan field bisnis yang dapat diedit.
- Place Name: kandidat nama Customer Site.
- Formatted Address: kandidat Preview Address Customer Site.
- Latitude dan Longitude: kandidat lokasi Customer Site.
- Google Place Types / Category: kandidat Customer Category, tetap memerlukan review/pemetaan.
- Phone Number: kandidat nomor telepon Customer Site.
- Website dan Google Maps URL: data referensi Prospect; target form saat ini tidak menyediakan field tujuan.

Google Places tidak dianggap menyediakan Province, District, Sub-District, Village sebagai field terpisah berdasarkan daftar data yang diberikan. Komponen tersebut tidak boleh diautofill sampai response aktual dan aturan pemetaan wilayah telah dikonfirmasi.

### 3.2 Data Prospect dan Sales Visit yang tidak menjadi field master langsung

| Data sumber | Pemakaian saat konversi |
|---|---|
| Prospect status dan pipeline status | Menentukan kelayakan `WON`; ditampilkan terkunci |
| Assigned Sales Executive | Prefill Sales Executive pada Customer Site; dapat ditinjau Administrator |
| Visit outcome, visit notes, follow-up notes | Bahan review; tidak disalin menjadi master Customer tanpa field tujuan yang disetujui |
| Duplicate-check result | Ditampilkan sebagai hasil kontrol sebelum konversi |
| Google Place ID | Link sumber dan kontrol duplikasi; disimpan sebagai metadata sumber, bukan input manual |

## 4. Conversion Field Mapping Matrix

Keterangan mode:

- **Autofill**: sistem membawa nilai yang memang tersedia; Administrator tetap melakukan review kecuali dinyatakan terkunci.
- **Manual**: Administrator mengetik atau memilih nilai.
- **Generated**: sistem membuat nilai tanpa input bisnis manual.
- **Master-data selected**: Administrator memilih nilai dari data master; bukan data Google Places.

| No | Customer Existing Field | Entity | Source Field | Source System | Autofill / Manual / Generated | Conversion Rule | Required at Conversion | Validation / Notes |
|----|-------------------------|--------|--------------|---------------|-------------------------------|-----------------|------------------------|-------------------|
| 1 | Customer Name / Outlet / Branch / Store | Customer Site | Place Name | Google Places | Autofill, editable | Ambil dari snapshot Prospect; Administrator dapat memperbaiki nama operasional tanpa mengubah snapshot sumber | Ya | Tidak boleh kosong; perubahan harus tercatat sebagai nilai hasil review |
| 2 | Customer Segment | Customer Site | Customer Segment | Administrator Input | Manual atau master-data selected | Administrator memilih segment dari daftar yang disetujui | Needs Business Confirmation | Daftar segment dan kewajibannya belum diberikan |
| 3 | Customer Category | Customer Site | Google Place Types / Category | Google Places | Autofill sebagai saran, editable | Type Google harus dipetakan ke kategori bisnis; bila tidak ada mapping, Administrator memilih manual | Needs Business Confirmation | Google type tidak boleh langsung dianggap sebagai kategori ERP tanpa mapping |
| 4 | Parent Code | Parent Company | Generated Parent Code / Parent Code dari company terpilih | System Generated / ERP / Master Data | Generated atau autofill, locked preview | Company baru: sistem menghasilkan preview Parent Code; Existing Company: ambil Parent Code dari master terpilih | Ya | Tidak dapat diedit bebas; harus unik. Format, prefix, sequence, dan kepemilikan generator adalah Needs Business Confirmation |
| 5 | Customer Code | Customer Site | Generated Customer Code | System Generated | Generated, locked preview | Sistem menghasilkan Customer Code sebelum konversi selesai dan menampilkannya sebagai preview | Ya | Tidak dapat diedit bebas; harus unik. Format, prefix, sequence, dan kepemilikan generator adalah Needs Business Confirmation |
| 6 | Company selection method | Parent Company | Pilihan Manual Entry / Company Name Matches Customer Name / Existing Company | Administrator Input | Manual | Tepat satu metode harus dipilih sebelum melanjutkan | Ya | Mengontrol sumber nama dan field company lainnya |
| 7 | Customer Company / Parent | Parent Company | Nama company sesuai metode pilihan | Administrator Input | Manual atau autofill kondisional | Manual Entry: ketik nama; Name Matches: salin nama Customer Site; Existing Company: ambil master terpilih | Ya | Existing Company harus memakai ID master, bukan hanya teks nama |
| 8 | Site Address — Search by Gmaps / Manual | Customer Site | Metode input alamat | Administrator Input | Manual | Administrator memilih sumber pengisian alamat Site | Ya | Pergantian metode tidak boleh menghapus snapshot Google sumber |
| 9 | Site Address — Province | Customer Site | Province | Administrator Input | Manual atau master-data selected | Isi/pilih wilayah setelah review alamat | Needs Business Confirmation | Autofill hanya boleh jika komponen API dan mapping wilayah telah disetujui |
| 10 | Site Address — District | Customer Site | District | Administrator Input | Manual atau master-data selected | Isi/pilih wilayah setelah Province | Needs Business Confirmation | Definisi District terhadap Kabupaten/Kota perlu konfirmasi bisnis |
| 11 | Site Address — Sub-District | Customer Site | Sub-District | Administrator Input | Manual atau master-data selected | Isi/pilih wilayah setelah District | Needs Business Confirmation | Mapping Kecamatan perlu konfirmasi |
| 12 | Site Address — Village | Customer Site | Village | Administrator Input | Manual atau master-data selected | Isi/pilih wilayah setelah Sub-District | Needs Business Confirmation | Mapping Kelurahan/Desa perlu konfirmasi |
| 13 | Site Address — Latitude | Customer Site | Latitude | Google Places | Autofill, editable melalui pemilihan peta | Ambil koordinat snapshot; perubahan dilakukan melalui pilihan lokasi, bukan teks bebas | Ya | Valid range `-90..90`; diperlukan untuk lokasi Customer Site |
| 14 | Site Address — Longitude | Customer Site | Longitude | Google Places | Autofill, editable melalui pemilihan peta | Ambil koordinat snapshot; perubahan dilakukan melalui pilihan lokasi, bukan teks bebas | Ya | Valid range `-180..180`; diperlukan untuk lokasi Customer Site |
| 15 | Site Address — Preview Address | Customer Site | Formatted Address | Google Places | Autofill, editable | Ambil formatted address snapshot; Administrator memverifikasi dan boleh melengkapi | Ya | Tidak menggantikan komponen wilayah terstruktur bila field tersebut diwajibkan |
| 16 | Company Address — Search by Gmaps / Manual | Parent Company | Metode input alamat company | Administrator Input | Manual | Administrator memilih metode; untuk Existing Company gunakan alamat master | Kondisional | Wajib untuk company baru; perilaku untuk existing company mengikuti master |
| 17 | Company Address — Province | Parent Company | Province company | Administrator Input | Manual atau master-data selected | Diisi untuk company baru; existing company mengambil master dan locked | Needs Business Confirmation | Tidak boleh otomatis disamakan dengan Site tanpa pilihan eksplisit |
| 18 | Company Address — District | Parent Company | District company | Administrator Input | Manual atau master-data selected | Diisi untuk company baru; existing company mengambil master dan locked | Needs Business Confirmation | Struktur wilayah mengikuti keputusan master data |
| 19 | Company Address — Sub-District | Parent Company | Sub-District company | Administrator Input | Manual atau master-data selected | Diisi untuk company baru; existing company mengambil master dan locked | Needs Business Confirmation | Struktur wilayah mengikuti keputusan master data |
| 20 | Company Address — Village | Parent Company | Village company | Administrator Input | Manual atau master-data selected | Diisi untuk company baru; existing company mengambil master dan locked | Needs Business Confirmation | Struktur wilayah mengikuti keputusan master data |
| 21 | Company Address — Latitude | Parent Company | Lokasi company yang dipilih | Administrator Input | Manual melalui Gmaps atau input terkontrol | Tidak diwarisi dari Site kecuali Administrator memilih bahwa alamat sama | Needs Business Confirmation | Valid range `-90..90` |
| 22 | Company Address — Longitude | Parent Company | Lokasi company yang dipilih | Administrator Input | Manual melalui Gmaps atau input terkontrol | Tidak diwarisi dari Site kecuali Administrator memilih bahwa alamat sama | Needs Business Confirmation | Valid range `-180..180` |
| 23 | Company Address — Preview Address | Parent Company | Alamat company yang dipilih/ditulis | Administrator Input | Manual | Existing company menggunakan alamat master; company baru memerlukan review | Needs Business Confirmation | Google Place Prospect hanya mewakili Site, bukan otomatis kantor Parent Company |
| 24 | Customer Site Contact — Contact Name | Customer Site Contact | Contact Name | Administrator Input | Manual | Administrator menambah kontak Site bila tersedia | Needs Business Confirmation | Google Places tidak menyediakan nama PIC bisnis |
| 25 | Customer Site Contact — Position | Customer Site Contact | Position | Administrator Input | Manual atau master-data selected | Isi jabatan kontak Site | Needs Business Confirmation | Daftar jabatan belum diberikan |
| 26 | Customer Site Contact — Phone Number | Customer Site Contact | Phone Number | Google Places | Autofill jika tersedia, editable | Nomor publik menjadi kandidat; Administrator memastikan apakah nomor outlet atau PIC | Needs Business Confirmation | Format dan kewajiban nomor perlu konfirmasi; jangan menganggap nomor publik sebagai nomor PIC |
| 27 | Customer Site Contact — Email Address | Customer Site Contact | Email Address | Administrator Input | Manual | Isi email kontak Site | Needs Business Confirmation | Google Places tidak dijadikan sumber email pada kebutuhan yang diberikan |
| 28 | Company Contact — Contact Name | Company Contact | Contact Name | Administrator Input | Manual atau master-data selected | Existing company mengambil kontak master; company baru diisi manual | Needs Business Confirmation | Kebijakan jumlah kontak belum diberikan |
| 29 | Company Contact — Position | Company Contact | Position | Administrator Input | Manual atau master-data selected | Existing company mengambil master; company baru diisi manual | Needs Business Confirmation | Daftar jabatan belum diberikan |
| 30 | Company Contact — Phone Number | Company Contact | Phone Number | Administrator Input | Manual | Tidak mengambil nomor Site kecuali Administrator menyalin secara eksplisit | Needs Business Confirmation | Validasi format perlu ditetapkan |
| 31 | Company Contact — Email Address | Company Contact | Email Address | Administrator Input | Manual | Isi email kontak company | Needs Business Confirmation | Validasi format email; kewajiban belum ditetapkan |
| 32 | PPN | Customer Site | PPN | Administrator Input | Manual atau master-data selected | Administrator memilih status/nilai PPN | Needs Business Confirmation | Bentuk data dan pilihan yang sah belum diberikan; bukan dari Google Places |
| 33 | ID TKU Number | Customer Site | ID TKU Number | Administrator Input | Manual | Isi berdasarkan dokumen bisnis yang diverifikasi | Needs Business Confirmation | Format, panjang, dan keunikan belum diberikan |
| 34 | NIK | Customer Site | NIK | Administrator Input | Manual | Isi hanya bila diperlukan dan memiliki dasar akses | Needs Business Confirmation | Data sensitif; aturan format, enkripsi/masking, dan kewajiban perlu konfirmasi |
| 35 | Company NPWP Name | Parent Company | Company NPWP Name | Administrator Input | Manual atau master-data selected | Existing company mengambil master; company baru diisi dari dokumen pajak | Needs Business Confirmation | Tidak boleh berasal dari Google Places |
| 36 | Company NPWP Address | Parent Company | Company NPWP Address | Administrator Input | Manual atau master-data selected | Existing company mengambil master; company baru diisi dari dokumen pajak | Needs Business Confirmation | Tidak otomatis sama dengan alamat Site atau Company Address |
| 37 | Company NPWP Number | Parent Company | Company NPWP Number | Administrator Input | Manual atau master-data selected | Existing company mengambil master; company baru diisi dan diverifikasi | Needs Business Confirmation | Format, keunikan, enkripsi/masking, dan normalisasi perlu konfirmasi |
| 38 | Shipment Cost | Customer Site | Shipment Cost | ERP / Master Data | Master-data selected atau manual | Pilih nilai sesuai kebijakan master | Needs Business Confirmation | Satuan, tipe nilai, dan daftar pilihan belum diberikan |
| 39 | Invoice Type | Customer Site | Invoice Type | ERP / Master Data | Master-data selected | Pilih jenis invoice dari master | Needs Business Confirmation | Tidak boleh diambil dari Google Places |
| 40 | Bank Account | Customer Site | Bank Account | Administrator Input | Manual atau master-data selected | Pilih/isi rekening sesuai kebijakan verifikasi | Needs Business Confirmation | Kepemilikan, format, masking, dan apakah rekening Site diperbolehkan perlu konfirmasi |
| 41 | Term of Payment | Parent Company | Term of Payment | ERP / Master Data | Master-data selected | Existing company mengambil master; company baru memilih nilai yang disetujui | Needs Business Confirmation | Daftar term dan otoritas persetujuan belum diberikan |
| 42 | Bill To Source | Customer Site | Bill To Source | ERP / Master Data | Master-data selected | Administrator memilih sumber penagihan yang sah | Needs Business Confirmation | Nilai pilihan dan relasinya ke Site/Company belum diberikan |
| 43 | Ship To Source | Customer Site | Ship To Source | ERP / Master Data | Master-data selected | Administrator memilih sumber pengiriman yang sah | Needs Business Confirmation | Nilai pilihan dan relasinya ke Site/Company belum diberikan |
| 44 | Sales Executive | Sales Assignment | Assigned Sales Executive | Prospect | Autofill, editable melalui pilihan user aktif | Warisi assignment aktif Prospect; Administrator dapat memilih Sales Executive aktif lain saat konversi | Ya | Harus hanya role Sales Executive aktif; perubahan tidak mengubah riwayat assignment Prospect |
| 45 | Sales Assignment — Start Month | Sales Assignment | Start Month | Administrator Input | Manual | Isi periode assignment Customer Site | Needs Business Confirmation | Apakah default dari bulan konversi belum disetujui |
| 46 | Sales Assignment — Start Year | Sales Assignment | Start Year | Administrator Input | Manual | Isi periode assignment Customer Site | Needs Business Confirmation | Harus konsisten dengan Start Month |
| 47 | Sales Assignment — End | Sales Assignment | End | Administrator Input | Manual | Isi akhir periode bila assignment berbatas waktu | Needs Business Confirmation | Tipe nilai dan apakah boleh kosong belum diberikan |
| 48 | Key Account Manager | KAM Assignment | Key Account Manager | ERP / Master Data | Master-data selected | Administrator memilih KAM yang sah untuk Parent Company | Needs Business Confirmation | Role KAM tidak termasuk dua role aplikasi; sumber master dan kewenangan perlu konfirmasi |
| 49 | KAM Assignment — Start Month | KAM Assignment | Start Month | Administrator Input | Manual | Isi periode KAM jika assignment diperlukan | Needs Business Confirmation | Default periode belum disetujui |
| 50 | KAM Assignment — Start Year | KAM Assignment | Start Year | Administrator Input | Manual | Isi periode KAM jika assignment diperlukan | Needs Business Confirmation | Harus konsisten dengan Start Month |
| 51 | KAM Assignment — End | KAM Assignment | End | Administrator Input | Manual | Isi akhir periode bila KAM berbatas waktu | Needs Business Confirmation | Tipe nilai dan apakah boleh kosong belum diberikan |

## 5. Required Conversion Form Rules

### 5.1 Prefilled dan editable

- Customer Name dari Place Name.
- Customer Category sebagai saran dari Google Place Types, hanya setelah tersedia mapping kategori.
- Site Preview Address dari Formatted Address.
- Site Latitude dan Longitude dari snapshot Google, tetapi perubahan hanya melalui kontrol lokasi/peta.
- Customer Site Phone Number jika Google Places menyediakan nomor; Administrator harus memastikan konteks nomor.
- Sales Executive dari assignment Prospect aktif; Administrator dapat memilih Sales Executive aktif lain sesuai flow yang disetujui.

Nilai prefill harus berasal dari snapshot Prospect, bukan memanggil Google secara diam-diam saat submit. Refresh Google, bila nanti disediakan, harus menjadi aksi terpisah dan tidak menimpa data tanpa review.

### 5.2 Prefilled tetapi locked

- Google Place ID.
- ID/link Prospect sumber.
- Status `WON` dan hasil duplicate check yang menjadi dasar kelayakan.
- Customer Code hasil generator sistem ditampilkan sebagai preview.
- Parent Code hasil generator sistem untuk Parent Company baru ditampilkan sebagai preview.
- Parent Code dan data master Parent Company saat Existing Company dipilih.
- Identitas master company terpilih; Administrator harus mengganti pilihan company, bukan mengedit teks master secara lokal.

Administrator tidak dapat mengedit Customer Code atau Parent Code secara bebas. Jika preview gagal dihasilkan, duplikat, atau tidak valid, konversi tidak dapat diselesaikan.

### 5.3 Minimum mandatory yang sudah dapat dipastikan

- Prospect masih berstatus `WON` dan belum dikonversi.
- Customer Name tidak kosong.
- Customer Code berhasil dihasilkan, ditampilkan sebagai preview, dan unik sebelum konversi selesai.
- Parent Code berhasil dihasilkan untuk Parent Company baru atau berhasil diambil dari Existing Company, ditampilkan sebagai preview, dan unik/valid.
- Company selection method dipilih.
- Parent Company berhasil dibuat atau ditautkan.
- Site Preview Address tersedia.
- Site Latitude dan Longitude valid.
- Sales Executive aktif tersedia.

Kewajiban field wilayah terstruktur, kontak, pajak, billing, bank, periode assignment, dan KAM tetap mengikuti label **Needs Business Confirmation** pada matrix. Setelah mentor menetapkan field wajib tambahan, field tersebut harus memblokir submit sampai valid.

### 5.4 Generated oleh sistem

Field form berikut dihasilkan atau diselesaikan sistem dan ditampilkan sebagai preview terkunci:

- Customer Code untuk Customer Site baru.
- Parent Code untuk Parent Company baru.
- Parent Code dari Existing Company tidak dibuat ulang; sistem mengambil nilai master company terpilih dan menampilkannya secara terkunci.

Administrator memilih metode Parent Company dan melengkapi data bisnis, tetapi tidak dapat mengetik bebas kedua kode tersebut. Format, prefix, sequence, serta keputusan apakah generator nantinya dimiliki CRM atau service ERP tetap **Needs Business Confirmation**.

Selain kode di atas, sistem menghasilkan metadata internal minimum berikut:

- ID internal Customer Site.
- Link ke Prospect sumber.
- Status Prospect `CONVERTED` setelah transaksi berhasil.
- Conversion timestamp.
- Actor Administrator ID.
- Converted Customer Site ID.
- Status transition history.

Full audit-event platform dan advanced idempotency infrastructure bukan syarat slice simulasi pertama dan dicatat sebagai production hardening yang ditunda.

### 5.5 Field yang bergantung pada Existing Company

Saat Existing Company dipilih:

- Administrator mencari dan memilih record master yang spesifik.
- Customer Company / Parent, Parent Code, company address, company contact, company tax, Term of Payment, dan KAM yang sudah menjadi bagian master ditampilkan dari record terpilih.
- Field master ditampilkan locked pada form konversi; perubahan master harus melalui proses master-data terpisah, bukan dengan membuat salinan berbeda pada Customer Site.
- Customer Site tetap record baru dan memiliki nama, alamat, kontak, billing/shipment, serta Sales Assignment sendiri.

### 5.6 Field manual karena tidak disediakan Google Places

- Customer Segment dan kategori bisnis final bila mapping tidak tersedia.
- Seluruh detail Parent Company selain opsi menyamakan nama.
- Province, District, Sub-District, dan Village sampai mapping aktual tersedia.
- Nama PIC, posisi, email, dan konteks nomor telepon.
- PPN, ID TKU, NIK, serta seluruh data NPWP.
- Shipment Cost, Invoice Type, Bank Account, dan Term of Payment.
- Bill To Source dan Ship To Source.
- Periode Sales Assignment.
- KAM dan periode KAM Assignment.

Customer Code dan Parent Code tidak termasuk field manual. Keduanya berupa preview terkunci sesuai metode company yang dipilih.

### 5.7 Posisi Key Account Manager

- Role login aplikasi tetap hanya `Administrator` dan `Sales Executive`.
- Key Account Manager adalah referensi dari ERP/master data untuk Parent Company, bukan role login aplikasi ketiga.
- Tidak ada akun, autentikasi, menu, atau authorization role KAM dalam scope simulasi ini.
- Perubahan KAM menjadi pengguna aplikasi hanya dapat dilakukan apabila mentor secara eksplisit mengubah keputusan role.

## 6. Duplicate and Company Handling

### 6.1 Google Place ID duplicate logic

- Google Place ID harus unik pada Prospect.
- Konversi menggunakan Google Place ID dari snapshot yang tersimpan dan tidak menerima penggantian ID dari form.
- Bila Google Place ID sudah terkait dengan Prospect lain, pembuatan Prospect baru harus ditolak sebelum mencapai flow konversi.
- Bila ditemukan Customer Existing yang berasal dari Google Place ID yang sama, konversi harus dihentikan dan Administrator diarahkan ke record yang sudah ada.
- Perbandingan nama/alamat dapat menjadi peringatan kandidat duplikat, tetapi tidak boleh menjadi hard match sebelum aturan fuzzy matching disetujui mentor.

### 6.2 Existing Customer duplicate logic

Sebelum konversi, sistem memeriksa minimal:

1. Prospect sumber belum memiliki Customer Existing.
2. Tidak ada Customer Existing dengan `source_prospect_id` yang sama.
3. Customer Code berhasil dihasilkan dan belum digunakan.
4. Parent Code baru berhasil dihasilkan dan belum digunakan, atau Parent Code Existing Company cocok dengan master company yang dipilih.
5. Tidak ada Customer Existing yang telah terkait dengan Google Place ID yang sama.

Slice simulasi pertama wajib mencegah konversi ganda dengan constraint/check terhadap Source Prospect ID dan status `CONVERTED`. Jika operasi diulang setelah sukses, sistem tidak membuat Customer kedua dan mengarahkan hasil ke Customer yang sudah terkonversi atau memberikan conflict yang aman. Infrastruktur idempotency lanjutan bukan syarat slice pertama.

Jika generator menghasilkan Customer Code atau Parent Code yang telah digunakan, sistem tidak boleh menyelesaikan konversi dengan kode duplikat. Mekanisme retry/reservasi sequence mengikuti rancangan generator setelah format dan ownership disetujui mentor.

### 6.3 Existing Parent Company selection logic

- Pencarian company menggunakan master company yang tersedia, bukan teks bebas yang dianggap sebagai link.
- Pemilihan menyimpan identitas Parent Company yang dipilih.
- Customer Site baru ditautkan ke Parent Company tersebut.
- Parent Code dan data company yang dikelola master tidak disalin menjadi data bebas yang dapat menyimpang.
- Parent Code Existing Company ditampilkan sebagai preview terkunci dan tidak dibuat ulang.
- Jika kandidat nama mirip ditemukan, Administrator tetap harus memilih secara eksplisit; sistem tidak melakukan auto-link hanya karena kemiripan nama.

### 6.4 Manual company creation logic

- Administrator memilih `Manual Entry` atau `Company Name Matches Customer Name`.
- Sistem menjalankan pemeriksaan kandidat Parent Company sebelum membuat record baru.
- Administrator melengkapi field company yang diwajibkan setelah keputusan mentor.
- Sistem menghasilkan Parent Code baru dan menampilkannya sebagai preview terkunci sebelum konversi diselesaikan.
- `Company Name Matches Customer Name` hanya menyalin nama; alamat, kontak, pajak, billing, dan KAM tidak boleh otomatis disamakan.
- Format, prefix, sequence, serta ownership generator Parent Code dan batas minimum data untuk company baru adalah **Needs Business Confirmation**.
- Pembuatan Parent Company baru dan Customer Site harus berada dalam transaksi konversi yang sama.

### 6.5 Prospect already converted

- Tombol/aksi konversi tidak tersedia sebagai aksi normal.
- Sistem menampilkan link ke Customer Existing hasil konversi.
- Operasi konversi tetap wajib memeriksa kondisi ini untuk mencegah double conversion akibat request ulang atau concurrency.
- Prospect tetap dapat dilihat sebagai riwayat, tetapi tidak kembali ke antrean Prospect aktif dan tidak dapat diubah menjadi `WON`/`LOST` kembali.

### 6.6 Batas scope simulasi dan production hardening

**Wajib untuk slice simulasi pertama:**

- Database transaction / atomic conversion.
- Duplicate prevention, termasuk proteksi double conversion dan keunikan kode.
- Conversion timestamp.
- Actor Administrator ID.
- Source Prospect ID.
- Converted Customer Site ID.
- Status transition history.

**Deferred production hardening:**

- Full audit-event platform.
- Advanced idempotency infrastructure di luar safe duplicate prevention.
- Rate limiting.
- Security-alert systems.
- Complex observability packages.

Item deferred tidak boleh dijadikan acceptance blocker untuk slice simulasi pertama. Penundaan ini tidak mengurangi kewajiban transaksi atomik, integritas referensi, keunikan kode, dan pencegahan double conversion.

## 7. Simulation Acceptance Criteria

1. Sales Executive dapat mengubah Prospect `FOLLOW_UP` menjadi `WON` atau `LOST` sesuai precondition.
2. Prospect `LOST` tidak dapat dikonversi dan tidak muncul di antrean review Won.
3. Prospect `WON` terlihat oleh Administrator dalam antrean review dan belum muncul sebagai Customer Existing.
4. Administrator dapat membuka form konversi hanya untuk Prospect `WON` yang belum dikonversi.
5. Place Name, Formatted Address, koordinat, kategori sebagai saran, nomor telepon bila tersedia, serta assigned Sales Executive terbawa ke form sesuai matrix.
6. Google Place ID, link Prospect sumber, dan status kelayakan tidak dapat diedit.
7. Customer Code dibuat sistem, ditampilkan sebagai preview terkunci, dan terbukti unik sebelum konversi selesai.
8. Parent Code untuk company baru dibuat sistem sebagai preview terkunci; Parent Code Existing Company berasal dari master terpilih dan juga terkunci.
9. Administrator tidak dapat mengetik atau mengedit bebas Customer Code maupun Parent Code.
10. Field manual yang ditetapkan wajib memblokir konversi sampai lengkap dan valid.
11. Administrator wajib memilih metode company dan konversi harus membuat Parent Company baru atau menautkan Existing Company.
12. Konversi membuat Customer Site yang tertaut ke Parent Company dan Prospect sumber.
13. Seluruh operasi konversi bersifat atomik; kegagalan pada satu langkah tidak meninggalkan Customer atau status setengah jadi.
14. Konversi menyimpan conversion timestamp, Actor Administrator ID, Source Prospect ID, Converted Customer Site ID, dan status transition history.
15. Prospect yang berhasil dikonversi ditandai `CONVERTED` dan tidak dapat dikonversi dua kali.
16. Customer hasil konversi muncul pada modul Customer Existing dan di `My Customer` milik Sales Executive yang dipilih.
17. Prospect sumber tidak dihapus; ia keluar dari antrean Prospect aktif dan tetap tersedia sebagai riwayat konversi.
18. Website dan Google Maps URL tetap dapat menjadi referensi sumber, tetapi tidak dipaksakan masuk ke field target yang belum tersedia.

## 8. Open Questions for Mentor

Pertanyaan berikut adalah keputusan bisnis yang tidak dapat disimpulkan dari field dan flow yang diberikan:

1. Apa daftar nilai resmi dan field wajib untuk Customer Segment serta Customer Category? Bagaimana Google Place Types dipetakan ke kategori tersebut?
2. Apa format, panjang, prefix, sequence, dan aturan keunikan Customer Code? Apakah generator sementara dimiliki CRM simulator atau langsung menggunakan service ERP?
3. Apa format, panjang, prefix, sequence, dan aturan keunikan Parent Code untuk Parent Company baru? Apakah generator sementara dimiliki CRM simulator atau langsung menggunakan service ERP?
4. Apakah Province, District, Sub-District, dan Village wajib, dan master wilayah mana yang menjadi acuan? Apakah District berarti Kabupaten/Kota dan Sub-District berarti Kecamatan?
5. Apakah alamat Parent Company wajib untuk company baru, dan apakah tersedia opsi eksplisit `Same as Customer Site`?
6. Berapa banyak Customer Site Contact dan Company Contact yang diperbolehkan atau diwajibkan pada saat konversi?
7. Apa bentuk nilai PPN, aturan ID TKU, serta format dan kebijakan keamanan NIK/NPWP?
8. Apa master dan aturan untuk Shipment Cost, Invoice Type, Term of Payment, Bill To Source, dan Ship To Source?
9. Bank Account pada bagian Customer Site merujuk pada rekening milik siapa, bagaimana verifikasinya, dan apakah dipilih dari master atau diketik?
10. Apakah periode Sales Assignment otomatis dimulai pada bulan/tahun konversi atau wajib dipilih Administrator? Apa tipe dan arti field `End`?
11. Dari ERP/master data mana daftar Key Account Manager berasal dan apakah KAM wajib untuk semua Parent Company? KAM tetap bukan role login aplikasi kecuali mentor membuat keputusan eksplisit yang berbeda.
12. Apakah field company dari Existing Company selalu locked, atau Administrator boleh mengajukan perubahan dalam flow konversi?
13. Apakah attendance radius yang terdapat pada arsitektur approved tetap wajib pada konversi meskipun tidak terdapat pada daftar field Customer Existing mentor?
14. Selain keberhasilan generation Customer Code dan kelengkapan Address, field mana yang dimaksud sebagai "additional required data" yang wajib memblokir konversi?

## 9. Ringkasan Klasifikasi

### 9.1 Autofill

- Customer Name dari Place Name.
- Site Preview Address dari Formatted Address.
- Site Latitude dan Longitude.
- Customer Category hanya sebagai saran apabila mapping tersedia.
- Customer Site Phone Number bila tersedia dan setelah review konteks.
- Sales Executive dari assignment aktif Prospect.
- Data Parent Company dari master ketika Existing Company dipilih.

### 9.2 Manual atau master-data selected

- Metode pemilihan company dan data bisnis Parent Company yang diwajibkan.
- Customer Segment/kategori final.
- Komponen wilayah terstruktur dan data Parent Company baru.
- Kontak Site dan Company.
- Seluruh data pajak/identitas.
- Shipment Cost, Invoice Type, Bank Account, Term of Payment.
- Bill To Source dan Ship To Source.
- Periode Sales Assignment serta KAM Assignment.

### 9.3 Generated

- Customer Code sebagai preview terkunci.
- Parent Code company baru sebagai preview terkunci; Parent Code Existing Company diambil dari master dan terkunci.
- ID internal Customer Site dan link Prospect sumber.
- Status `CONVERTED`, conversion timestamp, Actor Administrator ID, Converted Customer Site ID, dan status transition history.

### 9.4 Needs Business Confirmation

- Daftar nilai dan requiredness sebagian besar field master.
- Format, prefix, sequence, aturan keunikan, dan ownership generator Customer Code serta Parent Code.
- Struktur wilayah serta aturan alamat company.
- Aturan kontak, pajak, identitas, bank, billing, dan shipment.
- Periode Sales/KAM, sumber master KAM, serta arti field `End`.
- Status attendance radius pada form konversi.
- Definisi lengkap "additional required data".

## 10. Recommended Next Implementation Slice

Setelah dokumen ini disetujui mentor, slice implementasi pertama yang direkomendasikan adalah satu vertical slice poin 10–14: keputusan `WON/LOST`, antrean review Won untuk Administrator, form konversi dengan prefill yang telah disetujui, validasi field wajib minimum, transaksi create/link Parent Company dan create Customer Site, proteksi double conversion, lalu publikasi ke Customer Existing dan `My Customer`.

Slice tersebut harus dimulai dari acceptance test dan kontrak status/field yang telah disetujui. Google Maps lanjutan, modul CRM lain, deployment, dan perluasan UI tetap di luar scope sampai ada persetujuan terpisah.
