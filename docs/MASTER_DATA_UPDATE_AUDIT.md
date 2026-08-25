# Audit Pembaruan Master Data

Tanggal audit: 25 Agustus 2026

Branch sumber: `feat/master-data-management`

Basis branch: `origin/main` pada commit `6c42724`

Remote tujuan: `origin`

## Ringkasan

Pembaruan ini menambahkan pengelolaan master data Segment dan Category, termasuk soft delete, trash/restore, pemetaan kata kunci Google Places, serta pemakaian taxonomy B2B/B2C pada Prospect Finder. Perubahan backend, frontend, schema Prisma, dan migrasi database harus diintegrasikan sebagai satu unit.

## Dampak Perubahan

### Backend

- Modul baru `backend/internal/masterdata` dengan lapisan handler, service, repository, model, dan unit test.
- Endpoint baru di `/api/master-data` untuk list/create/update/delete Segment dan Category serta list/restore Trash.
- Akses endpoint dibatasi untuk `SUPER_ADMIN` dan `ADMINISTRATOR` yang telah mengganti password.
- Bootstrap dan server menerima dependency `masterDataService`.
- Google Places menggunakan taxonomy/category yang telah dipetakan dari master data.

### Database

Migrasi harus dijalankan sesuai urutan nama folder berikut:

1. `202608210001_master_data_segment_category`
2. `202608210002_master_data_trash`
3. `202608210003_master_data_place_api`
4. `202608240001_master_data_prospect_finder_categories`
5. `202608240002_master_data_customer_type`
6. `202608240003_master_data_b2b_b2c_segments`

Hasil akhir taxonomy adalah dua Segment utama, `B2B` dan `B2C`, dengan Category yang digunakan Prospect Finder. Migrasi kelima membuat struktur Customer Type sementara dan migrasi keenam menghapusnya setelah taxonomy direbase; keduanya tetap wajib dijalankan berurutan pada database yang belum pernah menerima rangkaian migrasi ini.

Sebelum deploy, buat backup database. Jangan menjalankan file SQL secara terpisah atau melompati migrasi karena migrasi terakhir bergantung pada struktur/data migrasi sebelumnya.

### Frontend

- Panel pengelolaan dan Trash Master Data tersedia pada layout Admin dan Sales yang memenuhi hak akses.
- Prospect Finder mengambil category dan keyword Places dari Master Data.
- Customer list, Sales dashboard, Tanya AI, dan AI Summary diselaraskan dengan taxonomy dan tampilan terbaru.
- Helper icon category dipusatkan di `frontend/src/utils/categoryIcons.ts`.

## Hasil Verifikasi

- `go test ./...`: lulus.
- `npm.cmd run typecheck`: lulus.
- `npm.cmd test`: lulus, 77 test setelah rebase ke `origin/main` terbaru.
- `git diff --check`: lulus setelah pembersihan trailing whitespace.
- `npm.cmd run build`: lulus setelah dijalankan dengan akses cache build normal; 563 modul ditransformasi.

## Prosedur Integrasi Tanpa Konflik

Developer yang tidak memiliki perubahan lokal:

```bash
git fetch origin
git switch main
git pull --ff-only origin main
git switch feat/master-data-management
```

Developer yang sedang bekerja pada branch sendiri:

```bash
git status
git add <file-yang-diubah>
git commit -m "wip: save work before master data integration"
git fetch origin
git rebase origin/main
git merge --no-ff origin/feat/master-data-management
```

Jika perubahan lokal belum siap di-commit, gunakan `git stash push -u` sebelum fetch/rebase dan `git stash pop` sesudah integrasi. Jangan pull langsung ke working tree yang kotor.

Untuk integrasi ke `main`, maintainer sebaiknya memperbarui branch fitur terhadap `origin/main`, menjalankan seluruh verifikasi, kemudian merge melalui pull request. Hindari push paksa ke `main`. Bila branch fitur perlu diperbarui setelah dipakai developer lain, gunakan commit tambahan; jangan rewrite history/force-push.

## Area Konflik yang Perlu Diwaspadai

- `backend/bootstrap/bootstrap.go` dan `backend/server/app.go` karena wiring dependency dan route.
- `backend/prisma/schema.prisma` dan seluruh folder migrasi karena perubahan schema harus tetap berurutan.
- `frontend/src/views/Admin/Prospect/ProspectFinderView.vue` karena integrasi taxonomy paling besar berada di sini.
- Layout Admin/Sales dan Sales Dashboard jika developer lain juga mengubah navigasi atau dashboard.

Saat konflik terjadi, pertahankan kontrak endpoint `/api/master-data`, urutan migrasi di atas, serta pemakaian taxonomy dari API. Setelah resolusi, ulangi test backend, typecheck, test frontend, dan build.
